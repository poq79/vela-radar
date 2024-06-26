package radar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/vela-ssoc/vela-radar/util"
	"github.com/vela-ssoc/vela-radar/web"
	"github.com/vela-ssoc/vela-radar/web/finder"

	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-radar/fingerprintx/plugins"
	"github.com/vela-ssoc/vela-radar/fingerprintx/scan"
	tunnel "github.com/vela-ssoc/vela-tunnel"
)

var typeof = reflect.TypeOf((*Radar)(nil)).String()

const (
	Idle uint32 = iota + 1
	Working
)

type Radar struct {
	lua.SuperVelaData
	Status   uint32
	cfg      *Config
	screen   *web.ScreenshotServer
	task     *Task
	lastTask *Task
	dr       tunnel.Doer
}

func (rad *Radar) IsWorking() bool {
	return atomic.LoadUint32(&rad.Status) == Working
}

func (rad *Radar) Info() []byte {
	enc := kind.NewJsonEncoder()

	enc.Tab("")
	enc.KV("name", rad.cfg.name)
	enc.KV("status", rad.TaskStatus())
	enc.KV("thread", rad.cfg.thread)
	enc.KV("udp", rad.cfg.FxConfig.UDP)
	enc.KV("fastmode", rad.cfg.FxConfig.FastMode)
	enc.KV("default_timeout", rad.cfg.FxConfig.DefaultTimeout.Milliseconds())
	if rad.task == nil {
		enc.KV("task", nil)
	} else {
		enc.Raw("task", rad.task.info())
	}
	if rad.lastTask == nil {
		enc.KV("last_task", nil)
	} else {
		enc.Raw("last_task", rad.lastTask.info())
	}
	enc.End("}")
	return enc.Bytes()
}

func (rad *Radar) TaskID() string {
	if rad.task == nil {
		return ""
	}

	return rad.task.Id
}

func (rad *Radar) TaskStatus() string {
	if rad.task == nil {
		return "idle"
	}
	return "working"
}

func (rad *Radar) Exception(err error) {
	xEnv.Errorf("radar handle fail %v", err)
}

func (rad *Radar) Catch(err error) {
	xEnv.Errorf("radar handle fail %v", err)
}

func (rad *Radar) Screen(tx *Tx, s *Service) {
	if rad.screen == nil {
		return
	}

	if !rad.screen.Avaliable {
		return
	}

	if s.HTTPInfo == nil {
		return
	}

	if !tx.Param.Screenshot {
		return
	}

	if s.HTTPInfo.Body == "" || s.HTTPInfo.Header == "Content-Type: text/plain\nContent-Length: 0\n" {
		return
	}

	target := &web.ScreenshotTask{
		Radartaskid:   uuid.NewString(),
		Url:           s.HTTPInfo.URL,
		Name:          "test",
		Wg:            sync.WaitGroup{},
		TargetCtx:     context.TODO(),
		Done:          make(chan int),
		TimeoutCancel: func() {},
	}
	//defer target.TimeoutCancel()
	rad.screen.Push(target)

	select {
	//case <-target.TargetCtx.Done():
	//	xEnv.Errorf(fmt.Sprintf("[-] Failed to take (URL:http://%s:%d) screenshot:", s.IP, s.Port) + "screenshot canceled or timed out.")
	case sig := <-target.Done:
		switch sig {
		case 1: //
			xEnv.Errorf("[-] http://%s:%d screenshot and upload succeed", s.IP, s.Port)
		case -1: //
			xEnv.Debugf("[-] http://%s:%d screenshot and upload fail", s.IP, s.Port)
		}
	}

	s.HTTPInfo.ScreenshotURL = target.ScreenshotURL

}

func (rad *Radar) handle(s *Service) {
	//count
	atomic.AddUint64(&rad.task.Count_asset, 1)

	// todo ignore (use cnd )

	rad.cfg.Chains.Do(s, rad.cfg.co, func(err error) {
		rad.Exception(err)
	})

	if rad.task.Report {
		// todo upload (use tunnel)
		// res, err := xEnv.Fetch("/api/v1/broker/proxy/siem/api/netapp/mono", bytes.NewReader(s.Bytes()), nil)
		req, err := http.NewRequest("POST", rad.cfg.ReportUri, bytes.NewReader(s.Bytes()))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			xEnv.Errorf("create request fail %v", err)
			return
		}
		res, err := rad.dr.Do(req)
		if err != nil {
			xEnv.Errorf("upload siem info fail %v", err)
			return
		}

		if res.StatusCode != 200 {
			xEnv.Errorf("upload siem info not ok %v", err)
		}
	}
}

func (rad *Radar) End() {
	atomic.StoreUint32(&rad.Status, Idle)
	rad.lastTask = rad.task
	rad.task = nil
}

func (rad *Radar) Callback(tx *Tx) {
	addr, _ := netip.AddrFromSlice(tx.Entry.Ip)

	target := plugins.Target{
		Address: netip.AddrPortFrom(addr, tx.Entry.Port),
		Host:    "localhost",
	}

	cfg := rad.cfg.Finger()

	srv, err := scan.Do(target, cfg)
	if err != nil {
		return
	}

	if srv == nil {
		return
	}

	s := Service{
		IP:        tx.Entry.Ip,
		Port:      tx.Entry.Port,
		Protocol:  srv.Protocol,
		Location:  rad.task.Option.Location,
		TLS:       srv.TLS,
		Transport: srv.Transport,
		Version:   srv.Version,
		Banner:    srv.Raw,
		TaskId:    rad.task.Id,
	}

	if tx.Param.Httpx && s.Protocol == "http" {
		var raw plugins.ServiceHTTP
		json.Unmarshal(srv.Raw, &raw)
		s.Component = raw.Technologies
		s.Banner = []byte{}
		tx.Web(&s)
	} else if tx.Param.Httpx && s.Protocol == "https" {
		var raw plugins.ServiceHTTPS
		json.Unmarshal(srv.Raw, &raw)
		s.Component = raw.Technologies
		s.Banner = []byte{}
		tx.Web(&s)
	}

	rad.Screen(tx, &s)
	rad.handle(&s)
}

func (rad *Radar) NewTask(target string) *Task {
	opt := Option{
		Target:           target,
		Port:             "top1000",
		Mode:             "pn", // syn or not
		Httpx:            false,
		Ping:             false,
		Rate:             500,
		Timeout:          800,
		ExcludeTimeRange: util.TimeRange{},
		Pool: Pool{
			Ping:   10,
			Scan:   10,
			Finger: 50,
		},
		FingerDB: "",
		MinioCfg: *rad.cfg.MinioCfg,
	}
	if opt.FingerDB != "" {
		info, err := xEnv.Third(opt.FingerDB)
		if err != nil {
			xEnv.Errorf("get 3rd %s ERROR %v", opt.FingerDB, err)
		}
		err = finder.LoadWebFingerData("./3rd/" + info.Name)
		if err != nil {
			xEnv.Errorf("load WebFingerData [%s] ERROR", info.Name)
		}
		xEnv.Infof("use 3rd WebFingerData [%s]..", info.Name)
	} else {
		err := finder.ParseWebFingerData(web.FingerData)
		if err != nil {
			xEnv.Errorf("get Built-in FingerData ERROR %v", opt.FingerDB, err)
		}
	}
	ctx, cancel := context.WithCancel(xEnv.Context())
	t := &Task{Option: opt, Dispatch: rad, ctx: ctx, cancel: cancel, co: xEnv.Clone(rad.cfg.co), rad: rad}
	rad.task = t
	return t
}

func NewRadar(cfg *Config) *Radar {
	d, err := xEnv.Doer(cfg.ReportDoer)
	if err != nil {
		fmt.Println("xEnv.Doer ERR")
	}
	rad := &Radar{
		cfg:    cfg,
		Status: Idle,
		dr:     d,
	}
	return rad
}
