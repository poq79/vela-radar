package web

import (
	"net"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	t.Log(ProbeHttpInfo(net.ParseIP("61.129.129.241"), 443, "https", 5*time.Second))
}
