package web

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-kit/fileutil"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
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
	Cfg       *ScreenshotCfg
	ctx       context.Context
	state     uint32
	Avaliable bool
	chrome    []chromedp.ExecAllocatorOption
	queue     chan *ScreenshotTask
	Logger    vela.Log
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
	st.queue = make(chan *ScreenshotTask)
	for i := 0; i < st.Cfg.Thread; i++ {
		go st.navigate(i, st.chrome)
	}
	return nil
}

func (st *ScreenshotServer) navigate(workerNum int, option []chromedp.ExecAllocatorOption) {

	pCtx, pCancel := chromedp.NewExecAllocator(st.ctx, option...)
	defer pCancel()

	ctx, cancel := chromedp.NewContext(pCtx) // chrome tab ctx
	defer cancel()

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

	screen := func(target *ScreenshotTask) error {
		var buf []byte
		target.TargetCtx, target.TimeoutCancel = context.WithTimeout(ctx, time.Duration(st.Cfg.Timeout)*time.Second)
		if err := chromedp.Run(target.TargetCtx, fullScreenshot(target.Url, 100, &buf)); err != nil {
			st.Logger.Errorf("[-] Failed to take (URL:%s) screenshot: %v", target.Url, err)
			target.Done <- -1
			chromedp.Run(ctx, chromedp.Navigate("about:blank"))
			return err
		}

		ScreenshotURL, err := util.UploadToMinio(st.Cfg.Miniocfg, target.Radartaskid+".png", bytes.NewReader(buf), int64(len(buf)))
		if err != nil {
			log.Errorf("[-] Failed to write file %v", err)
			target.Done <- -1
			return err
		}
		target.ScreenshotURL = ScreenshotURL

		if st.Cfg.Save {
			pngFilePath := fmt.Sprintf("./%s/%s.png", st.Cfg.ResultDir, target.Radartaskid)
			if e := os.WriteFile(pngFilePath, buf, 0644); e != nil {
				st.Logger.Errorf("[-] Failed to write file %v", err)
				return e
			}
		}

		if ctx.Err() != nil {
			st.Logger.Infof("[+] screenshot failed, timeout! %s", target.Url)
			// target.Wg.Done()
			target.Done <- -1
			chromedp.Run(ctx, chromedp.Tasks{chromedp.Navigate("about:blank")})
			return err
		}

		chromedp.Run(ctx, chromedp.Tasks{chromedp.Navigate("about:blank")})
		target.Done <- 1
		return nil
	}

	for target := range st.queue {
		screen(target)
	}
}

func (st *ScreenshotServer) Push(target *ScreenshotTask) {
	if st.Avaliable {
		st.queue <- target
	}
}

func (st *ScreenshotServer) Close() {
	defer func() {
		if e := recover(); e != nil {
			st.Logger.Errorf("%v", e)
		}
	}()
	if st.queue == nil || len(st.queue) == 0 {
		return
	}
	close(st.queue)
}

func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		//chromedp.WaitVisible("style"),
		chromedp.WaitReady("body"),
		chromedp.Sleep(3 * time.Second),
		//chromedp.OuterHTML(`document.querySelector("body")`, &htmlContent, chromedp.ByJSPath),
		chromedp.FullScreenshot(res, quality),
	}
}
