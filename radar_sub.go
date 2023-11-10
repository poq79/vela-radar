package radar

import (
	"github.com/vela-ssoc/vela-radar/fingerprintx/plugins"
	"github.com/vela-ssoc/vela-radar/fingerprintx/scan"
	"github.com/vela-ssoc/vela-radar/port"
	"net/netip"
	"time"
)

// OnMessage naabu queue message
func (rad *Radar) OnMessage(v interface{}) {
	tx, ok := v.(port.OpenIpPort)
	if !ok {
		return
	}

	addr, _ := netip.AddrFromSlice(tx.Ip)

	target := plugins.Target{
		Address: netip.AddrPortFrom(addr, tx.Port),
		Host:    "localhost",
	}

	cfg := scan.Config{
		UDP:            false,
		FastMode:       false,
		DefaultTimeout: time.Second,
		Verbose:        false,
	}

	srv, err := scan.Do(target, cfg)
	if err != nil {
		return
	}

	if srv == nil {
		return
	}

	h := Service{
		IP:        tx.Ip,
		Port:      tx.Port,
		Protocol:  srv.Protocol,
		TLS:       srv.TLS,
		Transport: srv.Transport,
		Version:   srv.Version,
		Banner:    srv.Raw,
	}
	rad.handle(&h)
}

func (rad *Radar) OnClose() {

}
