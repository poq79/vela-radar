package radar

import (
	"github.com/vela-ssoc/vela-kit/lua"
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
	default:
		//todo
	}

	return lua.LNil
}
