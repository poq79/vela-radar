package radar

import (
	"time"

	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"github.com/vela-ssoc/vela-radar/fingerprintx/scan"
	"github.com/vela-ssoc/vela-radar/util"
)

type Config struct {
	//字段名
	name       string
	co         *lua.LState
	thread     int
	FxConfig   *scan.Config
	MinioCfg   *util.MinioCfg
	ReportDoer string
	ReportUri  string
	Chains     *pipe.Chains
}

func NewConfig(L *lua.LState) *Config {
	cfg := &Config{
		co:     xEnv.Clone(L), //产生子虚拟机
		thread: 25,
		Chains: pipe.New(pipe.Env(xEnv)),
		FxConfig: &scan.Config{
			DefaultTimeout: time.Second,
			FastMode:       false,
			UDP:            false,
		},
		MinioCfg:   &util.MinioCfg{},
		ReportDoer: "/api/v1/broker/proxy/siem/",
		ReportUri:  "/api/netapp/mono",
	}

	tab := L.CheckTable(1)

	tab.Range(func(s string, value lua.LValue) {
		cfg.NewIndex(L, s, value)
	})

	return cfg
}

func (cfg *Config) Finger() scan.Config {
	return scan.Config{
		DefaultTimeout: cfg.FxConfig.DefaultTimeout,
		FastMode:       cfg.FxConfig.FastMode,
		UDP:            cfg.FxConfig.UDP,
	}
}

func (cfg *Config) FingerConfig(L *lua.LState, val lua.LValue) {
	if val.Type() != lua.LTTable {
		L.RaiseError("finger config must table , got %s", val.Type().String())
		return
	}

	tab := val.(*lua.LTable)

	tab.Range(func(key string, value lua.LValue) {
		switch key {
		case "timeout":
			n := lua.IsInt(value)
			if n < 100 {
				n = 100
			}
			cfg.FxConfig.DefaultTimeout = time.Duration(n) * time.Millisecond
		case "fast":
			cfg.FxConfig.FastMode = lua.IsTrue(value)
		case "verbose":
			cfg.FxConfig.Verbose = lua.IsTrue(value)
		case "udp":
			cfg.FxConfig.UDP = lua.IsTrue(value)
		}
	})

}

func (cfg *Config) MinioConfig(L *lua.LState, val lua.LValue) {
	if val.Type() != lua.LTTable {
		L.RaiseError("minio config must table , got %s", val.Type().String())
		return
	}
	tab := val.(*lua.LTable)
	tab.Range(func(key string, value lua.LValue) {
		switch key {
		case "accessKey":
			cfg.MinioCfg.AccessKey = lua.IsString(value)
		case "secretKey":
			cfg.MinioCfg.SecretKey = lua.IsString(value)
		case "endpoint":
			cfg.MinioCfg.Endpoint = lua.IsString(value)
		case "name":
			cfg.MinioCfg.Name = lua.IsString(value)
		case "useSSL":
			cfg.MinioCfg.UseSSL = lua.IsTrue(value)
		}
	})

}

func (cfg *Config) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "name":
		cfg.name = val.String()
	case "thread":
		n := lua.IsInt(val)
		if n < 25 {
			n = 25
		}
		cfg.thread = n
	case "finger":
		cfg.FingerConfig(L, val)
	case "minio":
		cfg.MinioConfig(L, val)
	case "reportDoer":
		cfg.ReportDoer = val.String()
	case "reportUri":
		cfg.ReportUri = val.String()
	//todo
	default:
		return
	}

}
