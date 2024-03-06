package radar

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/iputil"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/thread"
	"github.com/vela-ssoc/vela-radar/host"
	"github.com/vela-ssoc/vela-radar/port"
	"github.com/vela-ssoc/vela-radar/port/syn"
	"github.com/vela-ssoc/vela-radar/port/tcp"
	"github.com/vela-ssoc/vela-radar/util"
)

type Dispatch interface {
	End()
	Callback(*Tx)
	Catch(error)
}

type Worker struct {
	Ping        *thread.PoolWithFunc
	Scan        *thread.PoolWithFunc
	FingerPrint *thread.PoolWithFunc
}

type WaitGroup struct {
	Ping        sync.WaitGroup
	Scan        sync.WaitGroup
	FingerPrint sync.WaitGroup
}

func (wg *WaitGroup) Wait(debug bool) {
	wg.Ping.Wait()
	if debug {
		xEnv.Infof("  wg.Ping.Wait() end")
	}
	wg.Scan.Wait()
	if debug {
		xEnv.Infof("  wg.Scan.Wait() end")
	}
	wg.FingerPrint.Wait()
	if debug {
		xEnv.Infof("  wg.FingerPrint.Wait() end")
	}

}

type Pool struct {
	Ping   int `json:"ping"`
	Scan   int `json:"scan"`
	Finger int `json:"finger"`
}

type Scanner interface {
	Scan(net.IP, uint16) error
	WaitLimiter() error
	Wait()
	Close()
}

type Task_Status int

const (
	Task_Status_Init Task_Status = iota
	Task_Status_Running
	Task_Status_Success
	Task_Status_Paused_By_Program
	Task_Status_Paused_Artificial
	Task_Status_Error
	Task_Status_Unknown
)

var Task_Status_Strings = [...]string{
	"Init",
	"Running",
	"Success",
	"paused_by_program",
	"paused_artificial",
	"Unknown",
}

func (s Task_Status) Detail() string {
	return Task_Status_Strings[s]
}

type Task struct {
	Name           string
	Id             string
	Debug          bool
	Status         Task_Status
	Count_all      uint64
	Count_success  uint64
	Count_asset    uint64
	Start_time     time.Time
	End_time       time.Time
	Timeuse_second float64
	Timeuse_msg    string
	Msg            string
	// FingerPrint_count_all        uint64 only for FingerPrint WaitGroup debug use
	// FingerPrint_count_success    uint64
	// Pause_signal                 int //暂停信号 1是机器自动暂停 2是用户人工手动暂停 0是正常运行
	co                           *lua.LState
	executionTimeMonitorStopChan chan struct{}
	rad                          *Radar
	Option                       Option
	Dispatch                     Dispatch
	Worker                       Worker
	WaitGroup                    WaitGroup
	ctx                          context.Context
	cancel                       context.CancelFunc
}

func (t *Task) String() string                         { return "" }
func (t *Task) Type() lua.LValueType                   { return lua.LTObject }
func (t *Task) AssertFloat64() (float64, bool)         { return 0, false }
func (t *Task) AssertString() (string, bool)           { return "", false }
func (t *Task) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (t *Task) Peek() lua.LValue                       { return t }

func (t *Task) close() error {
	if t.cancel == nil {
		return nil
	}
	t.cancel()
	return nil
}

func (t *Task) info() []byte {
	enc := kind.NewJsonEncoder()
	timeuse_second, timeuse_msg := t.get_timeuse_second()
	enc.Tab("")
	enc.KV("name", t.Name)
	enc.KV("id", t.Id)
	enc.KV("debug", t.Debug)
	enc.KV("status", t.Status)
	enc.KV("msg", t.Msg)
	enc.KV("start_time", t.Start_time)
	enc.KV("end_time", t.End_time)
	enc.KV("timeuse_second", timeuse_second)
	enc.KV("timeuse_msg", timeuse_msg)
	enc.KV("task_all_num", t.Count_all)
	enc.KV("task_success_num", t.Count_success)
	enc.KV("task_asset_num", t.Count_asset)
	enc.KV("task_process", fmt.Sprintf("%0.2f", float64(t.Count_success)/float64(t.Count_all)*100))
	enc.Raw("option", util.ToJsonBytes(t.Option))
	enc.End("}")
	return enc.Bytes()
}

func (t *Task) get_timeuse_second() (float64, string) {
	switch t.Status {
	case Task_Status_Init:
		t.CalculateTimeUse()
	case Task_Status_Running:
		t.CalculateTimeUse()
	case Task_Status_Paused_By_Program:
		t.CalculateTimeUse()
	case Task_Status_Paused_Artificial:
		t.CalculateTimeUse()
	default:

	}
	return t.Timeuse_second, t.Timeuse_msg
}

func (t *Task) end() {
	t.End_time = time.Now()
	t.CalculateTimeUse()
	t.Status = Task_Status_Success
	audit.NewEvent("PortScanTask.end").Subject("调试信息").From(t.co.CodeVM()).Msg(fmt.Sprintf("scan task succeed, id=%s, time use:%s", t.Id, t.Timeuse_msg)).Log().Put()
	t.Dispatch.End()
	if t.rad.cfg.Debug || t.Debug {
		xEnv.Infof("task end")
	}
}

func (t *Task) endWithErr(msg string) {
	t.Msg = msg
	t.End_time = time.Now()
	t.CalculateTimeUse()
	t.Status = Task_Status_Error
	audit.NewEvent("PortScanTask.error").Subject("调试信息").From(t.co.CodeVM()).Msg(msg).Log().Put()
	close(t.executionTimeMonitorStopChan)
	t.Dispatch.End()
	if t.rad.cfg.Debug || t.Debug {
		xEnv.Infof("task end with error : %s", msg)
	}
}

func (t *Task) CalculateTimeUse() {
	timeuse := time.Since(t.Start_time)
	timeuseMsg := fmt.Sprintf("%d小时%02d分钟%02d秒", int(timeuse.Hours()), int(timeuse.Minutes())%60, int(timeuse.Seconds())%60)
	t.Timeuse_second = timeuse.Seconds()
	t.Timeuse_msg = timeuseMsg
}

func (t *Task) executionMonitor() {
	// 可做全局监视器debug用
	t.executionTimeMonitorStopChan = make(chan struct{})
	if t.Option.ExcludeTimeRange.Daily == "" {
		if t.rad.cfg.Debug || t.Debug {
			xEnv.Infof("没有设置排除时间，直接执行")
		}
		return
	}
	for {
		select {
		case <-t.executionTimeMonitorStopChan:
			if t.rad.cfg.Debug || t.Debug {
				xEnv.Infof("扫描任务结束, 接收到终止信号, 退出执行时间监控器协程")
			}
			return
		default:
			isInTimeRange, err := util.IsWithinRange(t.Option.ExcludeTimeRange)
			if err != nil {
				t.Status = Task_Status_Paused_By_Program
				xEnv.Errorf("task execution time monitor fail %v", err)
				return
			}
			if t.Status != Task_Status_Paused_Artificial && isInTimeRange {
				t.Status = Task_Status_Paused_By_Program
			} else if t.Status != Task_Status_Paused_Artificial && !isInTimeRange {
				t.Status = Task_Status_Running
			}
			// 等待5秒
			time.Sleep(5 * time.Second)
		}

	}
}

func (t *Task) GenRun() {
	if t.rad.cfg.Debug || t.Debug {
		xEnv.Infof("[debug is enable]")
	}
	if t.Dispatch == nil {
		xEnv.Errorf("dispatch got nil")
		t.endWithErr("dispatch got nil")
		return
	}

	if t.rad.screen != nil {
		t.rad.screen.Start()
	}

	audit.NewEvent("PortScanTask.start").Subject("调试信息").From(t.co.CodeVM()).Msg(fmt.Sprintf("scan task start, id=%s", t.Id)).Log().Put()
	//fmt.Printf("scan task start, id=%s config: %s", t.Id, string(t.info()))
	var ss Scanner
	var err error
	wg := new(WaitGroup)
	// parse ip
	items := strings.Split(t.Option.Target, ",")
	excluded_ip_map := util.IpstrWithCommaToMap(t.Option.ExcludedTarget)
	// 解析端口字符串并且优先发送 TopTcpPorts 中的端口, eg: 1-65535,top1000
	ports, err := port.ShuffleParseAndMergeTopPorts(t.Option.Port)
	if err != nil {
		t.endWithErr(fmt.Sprintf("task port range parse fail %v", err))
		return
	}
	for n, ip := range items {
		if t.rad.cfg.Debug || t.Debug {
			xEnv.Infof("get Target[%d] %s", n, ip)
		}
		it, _, err := iputil.NewIter(ip)
		if err != nil {
			t.endWithErr(fmt.Sprintf("task ip range[%s] parse fail %v", ip, err))
			return
		}

		t.Count_all = t.Count_all + it.TotalNum()*uint64(len(ports))
	}

	// end init, start running
	t.Status = Task_Status_Running
	fingerPool, _ := thread.NewPoolWithFunc(t.Option.Pool.Finger, func(v interface{}) {
		defer wg.FingerPrint.Done()
		entry := v.(port.OpenIpPort)
		t.Dispatch.Callback(&Tx{Entry: entry, Param: t.Option})
		// atomic.AddUint64(&t.FingerPrint_count_success, 1)
	})
	defer fingerPool.Release()

	call := func(v port.OpenIpPort) {
		atomic.AddUint64(&t.Count_success, 1)
		if v.Ip == nil {
			return
		}
		// atomic.AddUint64(&t.FingerPrint_count_all, 1)
		wg.FingerPrint.Add(1)
		fingerPool.Invoke(v)
	}

	for n, ip := range items {
		it, startIp, err := iputil.NewIter(ip)
		if err != nil {
			t.endWithErr(fmt.Sprintf("task ip range[%s] parse fail (scanning): %v", ip, err))
			return
		}
		switch t.Option.Mode {
		case "syn":
			ss, err = syn.NewSynScanner(startIp, call, port.Option{
				Rate:    t.Option.Rate,
				Timeout: t.Option.Timeout,
			})
		default:
			ss, err = tcp.NewTcpScanner(call, port.Option{
				Rate:    t.Option.Rate,
				Timeout: t.Option.Timeout,
			})
		}

		// port scan func
		scanner := func(ip net.IP) {
			n := len(ports)
			if n == 1 {
				ss.WaitLimiter() // limit rate
				ss.Scan(ip, ports[0])
				return
			}

			for i := 0; i < n; i++ {
				ss.WaitLimiter() // limit rate
				ss.Scan(ip, ports[i])
			}
		}

		// host group scan func
		scan, _ := thread.NewPoolWithFunc(t.Option.Pool.Scan, func(v interface{}) {
			ip := v.(net.IP)
			for t.Status == Task_Status_Paused_By_Program || t.Status == Task_Status_Paused_Artificial {
				time.Sleep(3 * time.Second)
			}
			scanner(ip)
			wg.Scan.Done()
		})
		defer scan.Release()

		// Pool - ping and port scan
		ping, _ := thread.NewPoolWithFunc(t.Option.Pool.Ping, func(v interface{}) {
			ip := v.(net.IP)
			for t.Status == Task_Status_Paused_By_Program || t.Status == Task_Status_Paused_Artificial {
				time.Sleep(3 * time.Second)
			}
			ok := host.IsLive(ip.String(), false, 800*time.Millisecond)
			wg.Ping.Done()

			if ok {
				wg.Scan.Add(1)
				scan.Invoke(ip)
			} else {
				// atomic.AddUint64(&t.Count_success, uint64(len(ports)))
				atomic.AddUint64(&t.Count_success, 1)
				atomic.AddUint64(&t.Count_all, uint64(1-len(ports)))
			}
		})
		defer ping.Release()

		shuffle := util.NewShuffle(it.TotalNum())    // shuffle
		for i := uint64(0); i < it.TotalNum(); i++ { // ip index
			select {
			case <-t.ctx.Done():
				if len(items) == n+1 {
					goto done
				} else {
					continue
				}
			default:
				ip := make(net.IP, len(it.GetIpByIndex(0)))
				copy(ip, it.GetIpByIndex(shuffle.Get(i))) // Note: dup copy []byte when concurrent (GetIpByIndex not to do dup copy)
				// 黑名单ip
				for t.Status == Task_Status_Paused_By_Program || t.Status == Task_Status_Paused_Artificial {
					time.Sleep(3 * time.Second)
				}
				if excluded_ip_map[ip.String()] {
					// atomic.AddUint64(&t.Count_success, uint64(len(ports)))
					atomic.AddUint64(&t.Count_success, 1)
					atomic.AddUint64(&t.Count_all, uint64(1-len(ports)))
				} else if t.Option.Ping {
					wg.Ping.Add(1)
					_ = ping.Invoke(ip)
				} else {
					wg.Scan.Add(1)
					_ = scan.Invoke(ip)
				}
			}
		}
	}

done:
	wg.Wait(t.rad.cfg.Debug || t.Debug)
	if t.rad.cfg.Debug || t.Debug {
		xEnv.Infof("wg.Wait end")
		// fmt.Printf("task end\n")
	}
	ss.Wait()
	if t.rad.cfg.Debug || t.Debug {
		xEnv.Infof("ss.Wait end")
		// fmt.Printf("ss.Wait end\n")
	}
	ss.Close()
	if t.rad.cfg.Debug || t.Debug {
		xEnv.Infof("[%s]scanner closed", t.Option.Mode)
		// fmt.Printf("[%s]scanner closed\n")
	}
	close(t.executionTimeMonitorStopChan)
	if t.rad.cfg.Debug || t.Debug {
		xEnv.Infof("executionTimeMonitorStopChan closed")
	}
	if t.rad.task.Option.Screenshot && t.rad.screen != nil {
		t.rad.screen.Close()
		if t.rad.cfg.Debug || t.Debug {
			xEnv.Infof("ScreenshotServer closed")
		}
	}
	t.end()
}
