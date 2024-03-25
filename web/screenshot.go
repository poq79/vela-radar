package web

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-kit/fileutil"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"

	"github.com/chromedp/chromedp"
	"github.com/vela-ssoc/vela-radar/util"
)

type ScreenshotCfg struct {
	Proxy     string
	Chrome    bool
	Thread    int
	Timeout   int
	Path      string //chrome的安装路径, 安装在默认路径则不配置
	ResultDir string
	Save      bool
	Debug     bool
	Miniocfg  *util.MinioCfg
}

func (s *ScreenshotCfg) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "save":
		s.Save = lua.IsTrue(val)
	case "thread":
		s.Thread = lua.IsInt(val)
	case "timeout":
		s.Timeout = lua.IsInt(val)
	case "resultDir":
		s.ResultDir = lua.IsString(val)
	case "debug":
		s.Debug = lua.IsTrue(val)

	}

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
	Cfg               *ScreenshotCfg
	ctx               context.Context
	navigateWaitgroup sync.WaitGroup
	state             uint32
	Avaliable         bool
	chrome            []chromedp.ExecAllocatorOption
	queue             chan *ScreenshotTask
	Logger            vela.Log
}

var (
	execPathNotFound        = "[-] Chrome executable file not found, please install Chrome or specify the chrome.exe"
	createResultFolderError = "[-] failed to create result folder:%v"
)

func NewScreenServer(ctx context.Context, cfg *ScreenshotCfg, log vela.Log) (*ScreenshotServer, error) {
	screen := &ScreenshotServer{
		ctx:       ctx,
		Cfg:       cfg,
		Logger:    log,
		Avaliable: true,
	}

	if e := fileutil.CreateIfNotExists(cfg.ResultDir, true); e != nil {
		return nil, e
	}

	// create context
	chrome := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !cfg.Chrome),
		chromedp.ProxyServer(cfg.Proxy),
		chromedp.Flag("mute-audio", true),
		chromedp.IgnoreCertErrors,
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		//chromedp.ExecPath(st.Cfg.Path),
		chromedp.WindowSize(1280, 720),
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox)

	if cfg.Proxy != "" {
		chrome = append(chrome, chromedp.Flag("proxy-server", cfg.Proxy))
	}

	screen.chrome = chrome

	return screen, nil
}

func (st *ScreenshotServer) Start() error {
	level := st.Logger.LoggerLevel()
	if st.Cfg.Debug {
		_ = level.Set("debug")
	} else {
		_ = level.Set("error")
	}
	st.Logger.Infof("[+]ScreenshotServer starting...")
	st.queue = make(chan *ScreenshotTask)
	for i := 0; i < st.Cfg.Thread; i++ {
		go st.navigate(i, st.chrome)
	}
	return nil
}

func (st *ScreenshotServer) navigate(workerNum int, option []chromedp.ExecAllocatorOption) {
	st.navigateWaitgroup.Add(1)
	defer st.navigateWaitgroup.Done()
	sctx, sCancel := context.WithCancel(context.Background())
	pCtx, pCancel := chromedp.NewExecAllocator(sctx, option...)
	ctx, cancel := chromedp.NewContext(pCtx) // Chrome tab ctx
	defer cancel()
	defer pCancel()
	defer sCancel()

	// Check & Make Browser not close
	if e := chromedp.Run(ctx, chromedp.Navigate("about:blank")); e != nil {
		st.Avaliable = false
		if strings.Contains(e.Error(), "executable file not found") {
			st.Logger.Error(execPathNotFound)
			return
		}
		st.Logger.Error(e)
		return
	}

	switchToBlank := func() error {
		nCtx, nCancel := context.WithTimeout(ctx, time.Duration(3)*time.Second)
		defer nCancel()
		err := chromedp.Run(nCtx, chromedp.Navigate("about:blank"))
		if err != nil {
			nCancel()
			// 如果超时则重新打开一个新的
			cancel()
			ctx, cancel = chromedp.NewContext(pCtx) // Chrome tab ctx
			n2Ctx, n2Cancel := context.WithTimeout(ctx, time.Duration(3)*time.Second)
			defer n2Cancel()
			err = chromedp.Run(n2Ctx, chromedp.Navigate("about:blank"))
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	}

	screen := func(target *ScreenshotTask) error {
		var buf []byte
		st.Logger.Debugf("[+]ScreenshotServer [debug 169] navigate [%d] start screen (URL:%s)", workerNum, target.Url)
		target.TargetCtx, target.TimeoutCancel = context.WithTimeout(ctx, time.Duration(st.Cfg.Timeout)*time.Second)
		defer target.TimeoutCancel()
		st.Logger.Debugf("[+]ScreenshotServer [debug 172] navigate [%d] start screen (URL:%s)", workerNum, target.Url)
		chromedp.ListenTarget(target.TargetCtx, func(ev interface{}) {
			if ev, ok := ev.(*page.EventJavascriptDialogOpening); ok {
				st.Logger.Infof("[+]ScreenshotServer navigate [%d] closing alert [%s] (URL:%s)", workerNum, ev.Message, target.Url)
				go func() {
					if err := chromedp.Run(ctx,
						page.HandleJavaScriptDialog(true),
					); err != nil {
						st.Logger.Errorf("[-] Failed to take (navigate: %d URL:%s) screenshot(HandleJavaScriptDialog): %v", workerNum, target.Url, err)
					}
				}()
			}
		})
		if err := chromedp.Run(target.TargetCtx, fullScreenshot(target.Url, 100, &buf)); err != nil {
			st.Logger.Errorf("[-] Failed to take (navigate: %d URL:%s) screenshot: %v", workerNum, target.Url, err)
			target.Done <- -1
			st.Logger.Debugf("[+]ScreenshotServer [debug 188] navigate [%d] start screen (URL:%s)", workerNum, target.Url)
			swerr := switchToBlank()
			if swerr != nil {
				st.Logger.Errorf("[-] Failed to take (navigate: %d URL:%s) screenshot: (switchToBlank EEROR)%v", swerr)
				return err
			}
			st.Logger.Debugf("[+]ScreenshotServer [debug 194] navigate [%d] start screen (URL:%s)", workerNum, target.Url)
			return err
		}
		st.Logger.Debugf("[+]ScreenshotServer [debug 197] navigate [%d] start screen (URL:%s)", workerNum, target.Url)
		ScreenshotURL, err := util.UploadToMinio(st.Cfg.Miniocfg, target.Radartaskid+".png", bytes.NewReader(buf), int64(len(buf)))
		if err != nil {
			st.Logger.Errorf("[-] Failed to upload to minio %v", err)
			target.Done <- -1
			return err
		}
		target.ScreenshotURL = ScreenshotURL

		if st.Cfg.Save {
			pngFilePath := fmt.Sprintf("./%s/%s.png", st.Cfg.ResultDir, target.Radartaskid)
			if err = os.WriteFile(pngFilePath, buf, 0644); err != nil {
				st.Logger.Errorf("[-] Failed to write file %v", err)
				target.Done <- -1
				return err
			}
		}
		st.Logger.Debugf("[+]ScreenshotServer [debug 214] navigate [%d] start screen (URL:%s)", workerNum, target.Url)
		if ctx.Err() != nil {
			st.Logger.Errorf("[+] screenshot (URL:%s) failed, ctx error: %s", target.Url, ctx.Err())
			// target.Wg.Done()
			ctx, cancel = chromedp.NewContext(pCtx)
			target.Done <- -1
			//_ = chromedp.Run(ctx, chromedp.Tasks{chromedp.Navigate("about:blank")})
			return err
		}
		st.Logger.Debugf("[+]ScreenshotServer [debug 223] navigate [%d] start screen (URL:%s)", workerNum, target.Url)
		_ = chromedp.Run(ctx, chromedp.Tasks{chromedp.Navigate("about:blank")})
		st.Logger.Debugf("[+]ScreenshotServer [debug 225] navigate [%d] start screen (URL:%s)", workerNum, target.Url)

		target.Done <- 1
		st.Logger.Debugf("[+]ScreenshotServer [debug 228] navigate [%d] start screen (URL:%s)", workerNum, target.Url)
		return nil
	}

	for target := range st.queue {
		st.Logger.Debugf("[+]ScreenshotServer navigate [%d] start screen (URL:%s)", workerNum, target.Url)
		_ = screen(target)
		st.Logger.Debugf("[+]ScreenshotServer navigate [%d] end screen  (URL:%s)", workerNum, target.Url)
	}
	st.Logger.Infof("[+]ScreenshotServer navigate [%d] closed", workerNum)
}

func (st *ScreenshotServer) Push(target *ScreenshotTask) {
	if st.Avaliable {
		st.queue <- target
	}
}

func (st *ScreenshotServer) Close() {
	defer func() {
		if e := recover(); e != nil {
			st.Logger.Errorf("ScreenshotServer Closer ERR:%v", e)
		}
	}()
	if st.queue == nil {
		st.Logger.Infof("ScreenshotServer queue为nil")
		return
	}
	if len(st.queue) == 0 {
		st.Logger.Infof("ScreenshotServer queue队列为空")
	}
	close(st.queue)
	st.Logger.Infof("ScreenshotServer queue closed")
	st.navigateWaitgroup.Wait()
	st.Logger.Infof("ScreenshotServer ended")
}

func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		//chromedp.WaitVisible("style"),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1 * time.Second),
		//chromedp.OuterHTML(`document.querySelector("body")`, &htmlContent, chromedp.ByJSPath),
		chromedp.FullScreenshot(res, quality),
	}
}
