package radar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-radar/fingerprintx/plugins"
	"github.com/vela-ssoc/vela-radar/fingerprintx/scan"
)

var typeof = reflect.TypeOf((*Radar)(nil)).String()

const (
	Idle uint32 = iota + 1
	Working
)

type Radar struct {
	lua.SuperVelaData
	Status uint32
	cfg    *Config
	task   *Task
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
	enc.KV("defaultTimeout", rad.cfg.FxConfig.DefaultTimeout.Milliseconds())
	if rad.task == nil {
		enc.KV("task", nil)
	} else {
		enc.KV("task_all_num", rad.task.Count_all)
		enc.KV("task_success_num", rad.task.Count_success)
		enc.KV("task_process", fmt.Sprintf("%0.2f", float64(rad.task.Count_success)/float64(rad.task.Count_all)*100))
		enc.Raw("task", rad.task.info())
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

func (rad *Radar) handle(s *Service) {
	rad.cfg.Chains.Do(s, rad.cfg.co, func(err error) {
		rad.Exception(err)
	})
}

func (rad *Radar) End() {
	atomic.StoreUint32(&rad.Status, Idle)
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
		TLS:       srv.TLS,
		Transport: srv.Transport,
		Version:   srv.Version,
		Banner:    srv.Raw,
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

	rad.handle(&s)
}

func (rad *Radar) NewTask(target string) *Task {
	opt := Option{
		Target:  target,
		Port:    "top1000",
		Mode:    "pn", // syn or not
		Httpx:   false,
		Ping:    false,
		Ctime:   time.Now(),
		Rate:    500,
		Timeout: 800,
		Pool: Pool{
			Ping:   10,
			Scan:   10,
			Finger: 50,
		},
	}

	ctx, cancel := context.WithCancel(xEnv.Context())
	t := &Task{Option: opt, Dispatch: rad, ctx: ctx, cancel: cancel, co: xEnv.Clone(rad.cfg.co)}
	rad.task = t
	return t
}

func NewRadar(cfg *Config) *Radar {
	naa := &Radar{
		cfg:    cfg,
		Status: Idle,
	}
	naa.define(xEnv.R())
	return naa
}
