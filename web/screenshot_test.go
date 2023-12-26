package web

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestShot(t *testing.T) {
	s, err := NewScreenServer(context.Background(), &ScreenshotCfg{
		Proxy:     "",
		Chrome:    true,
		Thread:    5,
		Timeout:   10,
		ResultDir: "res",
	}, nil)
	if err != nil {
		t.Errorf("NewScreenServer ERR: ", err)
	}
	for i := 1; i <= 10; i++ {
		t := ScreenshotTask{
			Radartaskid: uuid.NewString(),
			Url:         fmt.Sprintf("http://%s", "www.baidu.com"),
			Name:        "test",
		}
		s.queue <- &t
		time.Sleep(1 * time.Second)
		//fmt.Printf("add target %d\n", i)
	}
	s.Close()

}
