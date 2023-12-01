package web

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
	"github.com/vela-ssoc/vela-radar/util"
)

type ScreenshotCfg struct {
	Proxy     string
	Chrome    bool
	Thread    int
	Timeout   int
	Path      string
	ResultDir string
	Save      bool
	Debug     bool
	Miniocfg  util.MinioCfg
}

type ScreenshotTask struct {
	Radartaskid   string
	Url           string
	Name          string
	ScreenshotURL string
	Done          chan int
	Wg            sync.WaitGroup
	TargetCtx     context.Context
	TimeoutCancel context.CancelFunc
}

type ScreenshotServer struct {
	Taskchan   chan *ScreenshotTask
	Wg         sync.WaitGroup
	Cfg        *ScreenshotCfg
	mutex      sync.Mutex
	StopSignal chan struct{}
	Logger     *log.Logger
}

var WebScreenshotServer ScreenshotServer

var (
	execPathNotFound        = "[-] Chrome executable file not found, please install Chrome or specify the chrome.exe path with --path"
	createResultFolderError = "[-] failed to create result folder:%v"
)

func (st *ScreenshotServer) InitService(c ScreenshotCfg) {
	st.Cfg = &c
	if st.Cfg.Debug {
		log.Infoln("Screenshot module Logger [Debug Mode]")
		// st.Logger = log.New()
		// log.SetOutput(os.Stdout)
		// log.SetFormatter(log.New().Formatter)
		log.SetLevel(log.DebugLevel)
	}
	st.Taskchan = make(chan *ScreenshotTask)
	st.Wg = sync.WaitGroup{}
	err := CreateDirIfNotExists(st.Cfg.ResultDir)
	if err != nil {
		log.Errorf(createResultFolderError, err)
		return
	}

	log.Infoln("[+] init screenshot task threads")

	for i := 1; i < st.Cfg.Thread+1; i++ {
		st.Wg.Add(1)
		go st.navigate(i)
		log.Debugf("[+] start chrome Thread %d\n", i)
	}
}

func (st *ScreenshotServer) navigate(workerNum int) {

	// create context
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", !st.Cfg.Chrome), chromedp.ProxyServer(st.Cfg.Proxy), chromedp.Flag("mute-audio", true), chromedp.IgnoreCertErrors, chromedp.DisableGPU, chromedp.NoFirstRun,
		//chromedp.ExecPath(st.Cfg.Path),
		chromedp.WindowSize(1280, 720), chromedp.NoDefaultBrowserCheck, chromedp.NoSandbox)

	if st.Cfg.Proxy != "" {
		opts = append(opts, chromedp.Flag("proxy-server", st.Cfg.Proxy))
	}
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer func() {
		log.Debugf("Screenshot chrome worker [%d] closed!!\n", workerNum)
		st.Wg.Done()
		cancel()
	}()
	// ctx, cancel = context.WithTimeout(ctx, time.Duration(st.cfg.timeout)*time.Second)
	ctx, cancel = chromedp.NewContext(ctx)
	// Check & Make Browser not close
	err := chromedp.Run(ctx, chromedp.Navigate("about:blank"))
	// err := chromedp.Run(ctx, chromedp.Tasks{
	// 	chromedp.Navigate("about:blank"),
	// })
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			log.Fatalln(execPathNotFound)
		} else {
			log.Fatalln(err)
		}
	}
	for target := range st.Taskchan {
		defer func() {
			// fmt.Printf("Screenshot chrome worker [%d] closed!!\n", workerNum)
			cancel()
		}()
		func() error {
			log.Debugf("[Worker%2d] starting screenshot task %s,remaining tasks:%d\n", workerNum, target.Url, len(st.Taskchan))
			var buf []byte
			// target.TargetCtx, target.TimeoutCancel = context.WithTimeout(context.Background(), 2*time.Second)
			target.TargetCtx, target.TimeoutCancel = context.WithTimeout(ctx, 10*time.Second)
			defer func() {
				target.TimeoutCancel()
				// log.Errorf("[Worker%2d] screenshot task %s  timeout....", workerNum, target.Url)
				// cancel()
				// ctx, cancel = chromedp.NewContext(ctx)
			}()
			// defer func() {
			// 	target.Done <- -1
			// 	target.TimeoutCancel()
			// 	return
			// }()

			// ctx, cancel = chromedp.NewContext(
			// 	ctx,
			// )
			// target.Wg.Add(1)
			if err := chromedp.Run(target.TargetCtx, fullScreenshot(target.Url, 100, &buf)); err != nil {
				log.Errorf("[-] Failed to take (URL:%s) screenshot: %v", target.Url, err)
				target.Done <- -1
				chromedp.Run(ctx, chromedp.Navigate("about:blank"))
				// c
				return err
			}
			ScreenshotURL, err := util.UploadToMinio(&st.Cfg.Miniocfg, "radar-screenshoot", target.Radartaskid+".png", bytes.NewReader(buf), int64(len(buf)))
			if err != nil {
				log.Errorf("[-] Failed to write file %v", err)
				target.Done <- -1
				return err
			}
			target.ScreenshotURL = ScreenshotURL

			if st.Cfg.Save {
				pngFilePath := fmt.Sprintf("./%s/%s.png", st.Cfg.ResultDir, target.Radartaskid)
				if err := os.WriteFile(pngFilePath, buf, 0644); err != nil {
					log.Errorf("[-] Failed to write file %v", err)
					return err
				}
			}

			// if err := page.Close().Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Target)); err != nil {
			// 	log.Errorln(err)
			// }

			log.Debugf("[Worker%2d] finished screenshot task %s,remaining tasks:%d\n", workerNum, target.Url, len(st.Taskchan))
			if ctx.Err() != nil {
				log.Infof("[+] screenshot failed, timeout! %s", target.Url)
				// target.Wg.Done()
				target.Done <- -1
				chromedp.Run(ctx, chromedp.Tasks{
					chromedp.Navigate("about:blank"),
				})
				return err
			}
			chromedp.Run(ctx, chromedp.Tasks{
				chromedp.Navigate("about:blank"),
			})
			target.Done <- 1
			return nil
		}()
	}
}

func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		//chromedp.WaitVisible("style"),
		chromedp.WaitReady("body"),
		chromedp.Sleep(3 * time.Second),
		//chromedp.OuterHTML(`document.querySelector("body")`, &htmlContent, chromedp.ByJSPath),
		chromedp.FullScreenshot(res, quality),
		//chromedp.ActionFunc(func(ctx context.Context) error {q
		//	_, _, _, _, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
		//	if err != nil {
		//		return err
		//	}
		//
		//	width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))
		//
		//		WithScreenOrientation(&emulation.ScreenOrientation{
		//			Type:  emulation.OrientationTypePortraitPrimary,
		//			Angle: 0,
		//		}).
		//		Do(ctx)
		//	if err != nil {
		//		return err
		//	}
		//
		//	*res, err = page.CaptureScreenshot().
		//		WithQuality(quality).
		//		WithClip(&page.Viewport{
		//			X:      contentSize.X,
		//			Y:      contentSize.Y,
		//			Width:  contentSize.Width,
		//			Height: contentSize.Height,
		//			Scale:  1,
		//		}).Do(ctx)
		//	if err != nil {
		//		return err
		//	}
		//	return nil
		//}),
	}
}

func CreateDirIfNotExists(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseResult(filename string) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var resultList [][]string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() != "" {
			splitStr := strings.Split(scanner.Text(), ", ")
			resultList = append(resultList, splitStr)
		}
	}
	return resultList, err
}

func AddScreenshotTarget(target *ScreenshotTask, cfg ScreenshotCfg) {
	if WebScreenshotServer.Cfg == nil {
		WebScreenshotServer.InitService(cfg)
	}
	WebScreenshotServer.Taskchan <- target
}
