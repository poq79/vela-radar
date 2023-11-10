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

func (t *Task) poolL(L *lua.LState) int {
	scan := L.IsInt(1)
	finger := L.IsInt(2)
	ping := L.IsInt(3)

	if scan > 5 {
		t.Option.Pool.Scan = scan
	}

	if finger > 5 {
		t.Option.Pool.Finger = finger
	}

	if ping > 5 {
		t.Option.Pool.Ping = ping
	}
	L.Push(t)
	return 1
}

// r.task("172.31.61.0/24").mode("syn").location("宛平南路88号/办公区/3F").httpx(true).port("top1000").ping(true).run()

func (t *Task) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "mode":
		return lua.NewFunction(t.modeL)
	case "location":
		return lua.NewFunction(t.locationL)
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
