package radar

import (
	"github.com/vela-ssoc/vela-radar/port"
	"github.com/vela-ssoc/vela-radar/web"
	"time"
)

type Tx struct {
	Entry port.OpenIpPort
	Param Option
}

func (tx *Tx) Web(s *Service) {

	info, ok := web.ProbeHttpInfo(tx.Entry.Ip, tx.Entry.Port, s.Protocol, time.Second*2)
	if !ok {
		return
	}

	s.HttpInfo = info
	xEnv.Errorf("%+v", info)
}
