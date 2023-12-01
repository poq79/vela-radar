package radar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vela-ssoc/vela-radar/port"
	"github.com/vela-ssoc/vela-radar/web"
)

type Tx struct {
	Entry port.OpenIpPort
	Param Option
}

func (tx *Tx) Web(s *Service) {

	info, ok := web.ProbeHttpInfo(s.IP, s.Port, s.Protocol, time.Second*2)
	if !ok {
		// fmt.Printf("%s:%d  ProbeHttpInfo not OK", s.IP.String(), s.Port)
		return
	}
	// fmt.Printf("%s:%d  ProbeHttpInfo...\n", s.IP.String(), s.Port)
	s.HTTPInfo = info
	if tx.Param.Screenshot {
		DoScreenshot(s, &web.ScreenshotCfg{
			Proxy:     "",
			Chrome:    false,
			Thread:    5,
			Timeout:   5,
			ResultDir: "res",
			Save:      true,
			Debug:     true,
			Miniocfg:  tx.Param.MinioCfg,
		})
	}
}

func DoScreenshot(s *Service, cfg *web.ScreenshotCfg) {
	// targetCtx, timeoutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	target := web.ScreenshotTask{
		Radartaskid:   uuid.NewString(),
		Url:           s.HTTPInfo.URL,
		Name:          "test",
		Wg:            sync.WaitGroup{},
		TargetCtx:     context.TODO(),
		Done:          make(chan int),
		TimeoutCancel: func() {},
	}
	web.AddScreenshotTarget(&target, *cfg)
	defer target.TimeoutCancel()
	select {
	case <-target.TargetCtx.Done():
		fmt.Println(fmt.Sprintf("http://%s:%d", s.IP, s.Port) + " canceled or timed out.")
		// target.Wg.Done()
	case sig := <-target.Done:
		if sig == 1 {
			fmt.Println(fmt.Sprintf("http://%s:%d", s.IP, s.Port) + " screenshot and upload succeed")
		} else if sig == -1 {
			fmt.Println(fmt.Sprintf("http://%s:%d", s.IP, s.Port) + " screenshot and upload ERROR")
		}
		// target.Wg.Done()
	}
	// target.Wg.Wait()
	s.HTTPInfo.ScreenshotURL = target.ScreenshotURL
}
