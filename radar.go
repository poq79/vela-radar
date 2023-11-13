package radar

import (
	"context"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-radar/fingerprintx/plugins"
	"github.com/vela-ssoc/vela-radar/fingerprintx/scan"
	"net/netip"
	"reflect"
	"sync/atomic"
	"time"
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

func (rad *Radar) TaskID() string {
	if rad.task == nil {
		return ""
	}

	return rad.task.Option.ID
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

	if tx.Param.Httpx && (s.Protocol == "http" || s.Protocol == "https") {
		tx.Web(&s)
	}

	rad.handle(&s)
}

func (rad *Radar) NewTask(target string) *Task {
	opt := Option{
		Target:  target,
		Port:    "top1000",
		Httpx:   false,
		Ping:    false,
		Ctime:   time.Now(),
		Rate:    1500,
		Timeout: 800,
		Pool: Pool{
			Ping:   50,
			Scan:   50,
			Finger: 500,
		},
	}

	ctx, cancel := context.WithCancel(xEnv.Context())
	t := &Task{Option: opt, Dispatch: rad, ctx: ctx, cancel: cancel}
	rad.task = t
	return t
}

func NewRadar(cfg *Config) *Radar {
	naa := &Radar{
		cfg:    cfg,
		Status: Idle,
	}

	return naa
}
