package web

import (
	"testing"
	"time"
)

func TestName(t *testing.T) {
	t.Log(ProbeHttpInfo("http://127.0.0.1:8008", 5*time.Second))
}
