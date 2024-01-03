package radar

import (
	"time"

	"github.com/vela-ssoc/vela-radar/util"
)

type Option struct {
	Location         string         `json:"location"`
	Mode             string         `json:"mode"`
	Target           string         `json:"target"`
	ExcludedTarget   string         `json:"exclude_target"`
	Port             string         `json:"port"`
	Rate             int            `json:"rate"`
	Timeout          int            `json:"timeout"`
	Httpx            bool           `json:"httpx"`
	Ping             bool           `json:"ping"`
	FingerDB         string         `json:"fingerDB"`
	Screenshot       bool           `json:"screenshot"`
	Pool             Pool           `json:"pool"`
	ExcludeTimeRange util.TimeRange `json:"excludeTimeRange"`
	MinioCfg         util.MinioCfg  `json:"-"`
	Ctime            time.Time      `json:"-"`
}

func (o *Option) set_rate(n int) {
	if n > 2000 {
		o.Rate = 2000
	} else {
		o.Rate = n
	}
}

func (o *Option) set_pool_ping(n int) {
	if n > 1000 {
		o.Pool.Ping = 1000
	} else {
		o.Pool.Ping = n
	}
}

func (o *Option) set_pool_scan(n int) {
	if n > 500 {
		o.Pool.Scan = 500
	} else {
		o.Pool.Scan = n
	}
}

func (o *Option) set_pool_finger(n int) {
	if n > 500 {
		o.Pool.Finger = 500
	} else {
		o.Pool.Finger = n
	}
}

func (o *Option) set_timeout(n int) {
	if n > 10000 {
		o.Timeout = 10000
	} else if n < 100 {
		o.Timeout = 100
	} else {
		o.Timeout = n
	}
}
