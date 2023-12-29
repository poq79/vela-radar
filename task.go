package radar

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
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

func (wg *WaitGroup) Wait() {
	wg.Ping.Wait()
	// fmt.Printf("  wg.Ping.Wait() end")
	wg.Scan.Wait()
	// fmt.Printf("  wg.Scan.Wait() end")
	wg.FingerPrint.Wait()
	// fmt.Printf("  wg.FingerPrint.Wait() end")

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

type Task struct {
	Name          string
	Id            string
	co            *lua.LState
	Count_all     uint64
	Count_success uint64
	// FingerPrint_count_all        uint64 only for FingerPrint WaitGroup debug use
	// FingerPrint_count_success    uint64
	Pause_signal                 int //暂停信号 1是机器自动暂停 2是用户人工手动暂停 0是正常运行
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
	enc.Tab("")
	enc.KV("name", t.Name)
	enc.KV("id", t.Id)
	enc.KV("status", "working")
	enc.KV("start_time", t.Option.Ctime.Format("2006-01-02 15:04:05"))
	enc.Raw("option", util.ToJsonBytes(t.Option))
	enc.End("}")
	return enc.Bytes()
}

func (t *Task) executionMonitor() {
	// 可做全局监视器debug用
	t.executionTimeMonitorStopChan = make(chan struct{})
	if t.Option.ExcludeTimeRange.Daily == "" {
		// fmt.Println("没有设置排除时间，直接执行")
		return
	}
	for {
		select {
		case <-t.executionTimeMonitorStopChan:
			// fmt.Println("扫描任务结束, 接收到终止信, 退出执行时间监控器协程")
			return
		default:
			isInTimeRange, err := util.IsWithinRange(t.Option.ExcludeTimeRange)
			if err != nil {
				t.Pause_signal = 2
				xEnv.Errorf("task execution time monitor fail %v", err)
				return
			}
			if t.Pause_signal != 2 && isInTimeRange {
				t.Pause_signal = 1
			} else if t.Pause_signal != 2 && !isInTimeRange {
				t.Pause_signal = 0
			}
			// 等待5秒
			time.Sleep(5 * time.Second)
		}

	}
}

func (t *Task) GenRun() {

	if t.Dispatch == nil {
		xEnv.Errorf("%s dispatch got nil")
		return
	}

	if t.rad.screen != nil {
		t.rad.screen.Start()
	}

	t.Id = uuid.NewString()
	audit.NewEvent("PortScanTask.start").Subject("调试信息").From(t.co.CodeVM()).Msg(fmt.Sprintf("scan task start, id=%s", t.Id)).Log().Put()
	//fmt.Printf("scan task start, id=%s config: %s", t.Id, string(t.info()))
	var ss Scanner
	var err error
	wg := new(WaitGroup)
	// parse ip
	it, startIp, err := iputil.NewIter(t.Option.Target)
	if err != nil {
		xEnv.Errorf("task ip range parse fail %v", err)
		audit.NewEvent("PortScanTask.error").Subject("调试信息").From(t.co.CodeVM()).Msg(fmt.Sprintf("task ip range parse fail %v", err)).Log().Put()
		t.Dispatch.End()
		return
	}

	excluded_ip_map := util.IpstrWithCommaToMap(t.Option.ExcludedTarget)

	// 解析端口字符串并且优先发送 TopTcpPorts 中的端口, eg: 1-65535,top1000
	ports, err := port.ShuffleParseAndMergeTopPorts(t.Option.Port)
	if err != nil {
		xEnv.Errorf("task port range parse fail %v", err)
		audit.NewEvent("PortScanTask.error").Subject("调试信息").From(t.co.CodeVM()).Msg(fmt.Sprintf("task port range parse fail %v", err)).Log().Put()
		t.Dispatch.End()
		return
	}
	// todo Support the input of multiple IP ranges
	t.Count_all = it.TotalNum() * uint64(len(ports))
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
		for t.Pause_signal == 1 {
			time.Sleep(3 * time.Second)
		}
		scanner(ip)
		wg.Scan.Done()
	})
	defer scan.Release()

	// Pool - ping and port scan
	ping, _ := thread.NewPoolWithFunc(t.Option.Pool.Ping, func(v interface{}) {
		ip := v.(net.IP)
		for t.Pause_signal == 1 {
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
			goto done
		default:
			ip := make(net.IP, len(it.GetIpByIndex(0)))
			copy(ip, it.GetIpByIndex(shuffle.Get(i))) // Note: dup copy []byte when concurrent (GetIpByIndex not to do dup copy)
			// 黑名单ip
			for t.Pause_signal == 1 {
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

done:
	wg.Wait()
	// fmt.Printf("wg.Wait end")
	ss.Wait()
	// fmt.Printf("ss.Wait end")
	ss.Close()
	close(t.executionTimeMonitorStopChan)
	timeuse := time.Since(t.Option.Ctime)
	hours := int(timeuse.Hours())
	minutes := int(timeuse.Minutes()) % 60
	seconds := int(timeuse.Seconds()) % 60
	timeuseMsg := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	audit.NewEvent("PortScanTask.end").Subject("调试信息").From(t.co.CodeVM()).Msg(fmt.Sprintf("scan task succeed, id=%s, time use:%s", t.Id, timeuseMsg)).Log().Put()
	t.Dispatch.End()
	// audit.Debug("task end").From(t.co.CodeVM()).Put()
}
