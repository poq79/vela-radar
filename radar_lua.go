package radar

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-radar/web"
	"time"
)

func (rad *Radar) Type() string {
	return typeof
}

func (rad *Radar) Name() string {
	return rad.cfg.name
}

func (rad *Radar) Close() error {
	if rad.task != nil {
		rad.task.close()
	}
	rad.UndoDefine()

	return nil
}

func (rad *Radar) Start() error {
	rad.V(time.Now(), lua.VTRun)
	return nil
}

func (rad *Radar) pipeL(L *lua.LState) int {
	rad.cfg.Chains.CheckMany(L)
	return 0
}

func (rad *Radar) NewTaskL(L *lua.LState) int {
	if rad.IsWorking() {
		L.RaiseError("%+v\n", "scan task running")
		return 0
	}

	target := L.CheckString(1)
	task := rad.NewTask(target)
	L.Push(task)
	return 1
}

func (rad *Radar) defineL(L *lua.LState) int {
	rad.Define()
	return 0
}

func (rad *Radar) startL(L *lua.LState) int {
	xEnv.Start(L, rad).
		Err(func(err error) { L.RaiseError("%v", err) }).
		From(L.CodeVM()).
		Do()
	return 0
}

func (rad *Radar) chromeL(L *lua.LState) int {
	//todo create chrome object
	cfg := &web.ScreenshotCfg{
		Proxy:     "",
		Chrome:    false,
		Thread:    5,
		Timeout:   5,
		ResultDir: "res",
		Save:      true,
		Debug:     true,
		Miniocfg:  rad.cfg.MinioCfg,
	}

	v := L.Get(1)
	switch v.Type() {
	case lua.LTTable:
		v.(*lua.LTable).Range(func(k string, value lua.LValue) {
			cfg.NewIndex(L, k, value)
		})

	default:
		goto DONE

	}

DONE:
	s, err := web.NewScreenServer(L.Context(), cfg, xEnv)
	if err != nil {
		L.RaiseError("screen server init fail %v", err)
		return 0
	}

	rad.screen = s
	return 1
}

// rad.chrome("/aab//cc")

func (rad *Radar) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "start":
		return lua.NewFunction(rad.startL)
	case "pipe":
		return lua.NewFunction(rad.pipeL)

	case "task":
		return lua.NewFunction(rad.NewTaskL)

	case "define":
		return lua.NewFunction(rad.defineL)

	case "chrome":
		return lua.NewFunction(rad.chromeL)

	default:
		//todo
	}

	return lua.LNil
}
