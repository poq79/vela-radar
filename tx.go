package radar

import (
	"fmt"
	"time"

	"github.com/vela-ssoc/vela-radar/port"
	"github.com/vela-ssoc/vela-radar/web"
)

type Tx struct {
	Entry port.OpenIpPort
	Param Option
}

func (tx *Tx) Web(s *Service) {
	url := fmt.Sprintf("%s://%s:%d/", s.Protocol, s.IP.String(), s.Port)

	info, ok := web.ProbeHttpInfo(url, time.Second*2)
	if !ok {
		// fmt.Printf("%s:%d  ProbeHttpInfo not OK", s.IP.String(), s.Port)
		return
	}
	// fmt.Printf("%s:%d  ProbeHttpInfo...\n", s.IP.String(), s.Port)
	s.HTTPInfo = info
}
