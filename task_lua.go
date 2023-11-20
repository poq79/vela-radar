package radar

import "github.com/vela-ssoc/vela-kit/lua"

func (t *Task) runL(L *lua.LState) int {
	go t.GenRun()
	return 0
}

func (t *Task) portL(L *lua.LState) int {
	port := L.CheckString(1)
	t.Option.Port = port
	L.Push(t)
	return 1
}

func (t *Task) modeL(L *lua.LState) int {
	mode := L.CheckString(1)
	t.Option.Mode = mode
	L.Push(t)
	return 1
}

func (t *Task) locationL(L *lua.LState) int {
	location := L.CheckString(1)
	t.Option.Location = location
	L.Push(t)
	return 1
}

func (t *Task) httpxL(L *lua.LState) int {
	t.Option.Httpx = L.IsTrue(1)
	L.Push(t)
	return 1
}

func (t *Task) pingL(L *lua.LState) int {
	t.Option.Ping = L.IsTrue(1)
	L.Push(t)
	return 1
}

func (t *Task) rateL(L *lua.LState) int {
	n := L.IsInt(1)
	t.Option.set_rate(n)
	L.Push(t)
	return 1
}

func (t *Task) poolL(L *lua.LState) int {
	scan := L.IsInt(1)
	finger := L.IsInt(2)
	ping := L.IsInt(3)
	t.Option.set_pool_scan(scan)
	t.Option.set_pool_finger(finger)
	t.Option.set_pool_ping(ping)
	L.Push(t)
	return 1
}

func (t *Task) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "mode":
		return lua.NewFunction(t.modeL)
	case "location":
		return lua.NewFunction(t.locationL)
	case "rate":
		return lua.NewFunction(t.rateL)
	case "port":
		return lua.NewFunction(t.portL)
	case "httpx":
		return lua.NewFunction(t.httpxL)
	case "pool":
		return lua.NewFunction(t.poolL)
	case "ping":
		return lua.NewFunction(t.pingL)
	case "run":
		return lua.NewFunction(t.runL)
	default:
		return lua.LNil
	}

}
