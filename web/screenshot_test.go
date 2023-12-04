package web

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestShot(t *testing.T) {
	WebScreenshotServer.InitService(ScreenshotCfg{
		Proxy:     "",
		Chrome:    true,
		Thread:    5,
		Timeout:   5,
		ResultDir: "res",
	})

	for i := 1; i <= 10; i++ {
		t := ScreenshotTask{
			Radartaskid: uuid.NewString(),
			Url:         fmt.Sprintf("http://%s", "www.baidu.com"),
			Name:        "test",
		}
		WebScreenshotServer.Taskchan <- &t
		time.Sleep(1 * time.Second)
		//fmt.Printf("add target %d\n", i)
	}
	close(WebScreenshotServer.Taskchan)
	WebScreenshotServer.Wg.Wait()

}
