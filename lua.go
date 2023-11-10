package radar

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
)

var xEnv vela.Environment

func NewRadarL(L *lua.LState) int {
	cfg := NewConfig(L)
	vda := L.NewVelaData(cfg.name, typeof) //判断出 当前code 是否有相同的对象 名字和类型
	if vda.IsNil() {
		vda.Set(NewRadar(cfg))
		L.Push(vda)
	} else {
		old := vda.Data.(*Radar)
		old.cfg = cfg
		L.Push(vda)
	}
	return 1

}

/*
	local s = vela.scanner{
	    name   = "scanner",
		thread = 10,
		report = vela.report()
	}

	s.pipe(function(host)
		//host.peer host.port host.banner  host.app  host.http_raw
	end)

	s.start()

	s.task("172.31.61.0/24").port("top1000").location("办公区").ping(false).pool(200 , 300 , 100).start()
*/

func WithEnv(env vela.Environment) {
	xEnv = env
	tab := lua.NewUserKV()
	xEnv.Set("radar", lua.NewExport("vela.radar.export", lua.WithTable(tab), lua.WithFunc(NewRadarL)))
}
