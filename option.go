package radar

import "time"

type Option struct {
	Location string    `json:"location"`
	Mode     string    `json:"mode"`
	Target   string    `json:"target"`
	Port     string    `json:"port"`
	Rate     int       `json:"rate"`
	Timeout  int       `json:"timeout"`
	Httpx    bool      `json:"httpx"`
	Ping     bool      `json:"ping"`
	Pool     Pool      `json:"pool"`
	Ctime    time.Time `json:"-"`
}

func (o *Option) set_rate(n int) {
	if n > 1000 {
		o.Rate = 1000
	} else {
		o.Rate = n
	}
}

func (o *Option) set_pool_ping(n int) {
	if n > 1000 {
		o.Pool.Ping = 1000
	} else {
		o.Rate = n
	}
}

func (o *Option) set_pool_scan(n int) {
	if n > 1000 {
		o.Pool.Scan = 1000
	} else {
		o.Rate = n
	}
}

func (o *Option) set_pool_finger(n int) {
	if n > 1000 {
		o.Pool.Scan = 1000
	} else {
		o.Rate = n
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
