package radar

import "github.com/vela-ssoc/vela-kit/lua"

func (t *Task) runL(L *lua.LState) int {
	go t.GenRun()
	go t.executionMonitor()
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

func (t *Task) screenshotL(L *lua.LState) int {
	t.Option.Screenshot = L.IsTrue(1)
	L.Push(t)
	return 1
}

func (t *Task) excludeL(L *lua.LState) int {
	excludeip := L.CheckString(1)
	t.Option.ExcludedTarget = excludeip
	L.Push(t)
	return 1
}

func (t *Task) fingerDBL(L *lua.LState) int {
	FingerDB := L.CheckString(1)
	t.Option.FingerDB = FingerDB
	L.Push(t)
	return 1
}

func (t *Task) timeoutL(L *lua.LState) int {
	timeout := L.CheckInt(1)
	t.Option.set_timeout(timeout)
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

func (t *Task) excludeTimeRangeL(L *lua.LState) int {
	Daily := L.CheckString(1)
	Begin := L.CheckString(2)
	End := L.CheckString(3)
	// t.Option.ExcludeTimeRange.Daily = Daily
	// t.Option.ExcludeTimeRange.Begin = Begin
	// t.Option.ExcludeTimeRange.End = End
	err := t.Option.set_ExcludeTimeRange_Daily(Daily)
	if err != nil {
		L.RaiseError("set_ExcludeTimeRange_Daily fail %v", err)
		return 0
	}
	err = t.Option.set_ExcludeTimeRange_Begin(Begin)
	if err != nil {
		L.RaiseError("set_ExcludeTimeRange_Begin fail %v", err)
		return 0
	}
	err = t.Option.set_ExcludeTimeRange_End(End)
	if err != nil {
		L.RaiseError("set_ExcludeTimeRange_End fail %v", err)
		return 0
	}
	L.Push(t)
	return 1
}

func (t *Task) debugL(L *lua.LState) int {
	t.Debug = L.IsTrue(1)
	L.Push(t)
	return 1
}

func (t *Task) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "exclude":
		return lua.NewFunction(t.excludeL)
	case "mode":
		return lua.NewFunction(t.modeL)
	case "location":
		return lua.NewFunction(t.locationL)
	case "rate":
		return lua.NewFunction(t.rateL)
	case "timeout":
		return lua.NewFunction(t.timeoutL)
	case "port":
		return lua.NewFunction(t.portL)
	case "httpx":
		return lua.NewFunction(t.httpxL)
	case "pool":
		return lua.NewFunction(t.poolL)
	case "ping":
		return lua.NewFunction(t.pingL)
	case "debug":
		return lua.NewFunction(t.debugL)
	case "screenshot":
		return lua.NewFunction(t.screenshotL)
	case "fingerDB":
		return lua.NewFunction(t.fingerDBL)
	case "excludeTimeRange":
		return lua.NewFunction(t.excludeTimeRangeL)
	case "run":
		return lua.NewFunction(t.runL)
	default:
		return lua.LNil
	}

}
