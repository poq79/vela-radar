package radar

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-ssoc/vela-kit/kind"
)

type TaskParam struct {
	ID                 string   `json:"id"`
	Host               []string `json:"host"`
	Mode               string   `json:"mode"`
	Ports              string   `json:"ports"`
	AllPort            bool     `json:"all_port"`
	Rate               int      `json:"rate"`
	Retries            int      `json:"retries"`
	Ping               bool     `json:"ping"`
	Verbose            bool     `json:"verbose"`
	ServiceDiscovery   bool     `json:"service_discovery"`
	SkipHostDiscovery  bool     `json:"skip_host_discovery"`
	DisableUpdateCheck bool     `json:"disable_update_check"`
	ServiceVersion     bool     `json:"service_version"`
}

func NewTaskParam() *TaskParam {
	return &TaskParam{
		Verbose: false,
		Mode:    "sn",
		// ScanAllIPS: true,
		// Ports: "80-10000",
		Retries:            2,
		Rate:               256, //wg thread worker goroutine
		Ping:               false,
		ServiceDiscovery:   true,
		SkipHostDiscovery:  true,
		DisableUpdateCheck: true,
		ServiceVersion:     true,
	}
}

func (tp *TaskParam) Clone() {
	/*
		option.Host = tp.Host
		option.Verbose = tp.Verbose
		option.ScanType = tp.Mode
		option.ScanAllIPS = tp.AllPort
		option.Ports = tp.Ports
		option.Retries = tp.Retries
		option.Rate = tp.Rate //wg thread worker goroutine
		option.Ping = tp.Ping
		option.ServiceDiscovery = tp.ServiceDiscovery
		option.SkipHostDiscovery = tp.SkipHostDiscovery
		option.DisableUpdateCheck = tp.DisableUpdateCheck
		option.ServiceVersion = tp.ServiceVersion

	*/
}

func (rad *Radar) TaskPath() string {
	return fmt.Sprintf("/api/v1/arr/agent/lua/radbu/%s/task", rad.Name())
}

func (rad *Radar) StatusPath() string {
	return fmt.Sprintf("/api/v1/arr/agent/lua/radbu/%s/status", rad.Name())
}

func (rad *Radar) TaskHandle(ctx *fasthttp.RequestCtx) error {
	/*
		param := NewTaskParam()
		body := ctx.PostBody()
		if len(body) == 0 {
			return fmt.Errorf("task handle fail got empty")
		}

		err := json.Unmarshal(body, &param)
		if err != nil {
			return err
		}

		task := rad.NewTask(param.Host)
		param.Clone(task.Option)
		ctx.Write(task.info())
		go task.GenRun()
	*/
	return nil
}

func (rad *Radar) StatusHandle(ctx *fasthttp.RequestCtx) error {
	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.KV("task_id", rad.TaskID())
	enc.KV("task_status", rad.TaskStatus())
	enc.End("}")
	ctx.Write(enc.Bytes())
	return nil
}

func (rad *Radar) Define() {
	r := xEnv.R()
	r.POST(rad.TaskPath(), xEnv.Then(rad.TaskHandle))
	r.GET(rad.StatusPath(), xEnv.Then(rad.StatusHandle))
}

func (rad *Radar) UndoDefine() {
	r := xEnv.R()
	r.Undo(fasthttp.MethodPost, rad.TaskPath())
	r.Undo(fasthttp.MethodGet, rad.StatusPath())
}
