package radar

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/valyala/fasthttp"
)

const taskParameterInitErr = "[API /api/v1/arr/agent/radar/runscan]Parameter initialization failed. Please check whether the parameter data type is correct--"

func (rad *Radar) TaskPath() string {
	// Generate URLs with specific information eg: name,location,workgroup..
	// ..
	return "/api/v1/arr/agent/radar/runscan"
}

func (rad *Radar) StatusPath() string {
	// Generate URLs with specific information  eg: name,location,workgroup..
	// ..
	return "/api/v1/arr/agent/radar/status"
}

func (rad *Radar) TaskHandle(ctx *fasthttp.RequestCtx) error {
	if rad.TaskStatus() == "working" {
		return errors.New("there are already scanning tasks running")
	}
	// 获取请求的 JSON 数据
	body := ctx.PostBody()

	// 解析 JSON 数据
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return err
	}
	if v, ok := data["target"].(string); ok && v != "" {
		rad.NewTask(v)
	} else {
		return errors.New(taskParameterInitErr + "target")
	}
	for key, value := range data {
		switch key {
		case "location":
			if v, ok := value.(string); ok {
				rad.task.Option.Location = v
			} else {
				return errors.New(taskParameterInitErr + key)
			}
		case "name":
			if v, ok := value.(string); ok {
				rad.task.Name = v
			} else {
				return errors.New(taskParameterInitErr + key)
			}
		case "mode":
			if v, ok := value.(string); ok {
				rad.task.Option.Mode = v
			} else {
				// return errors.New(init_err)
			}
		case "port":
			if v, ok := value.(string); ok {
				rad.task.Option.Port = v
			} else {
				// return errors.New(init_err)
			}
		case "rate":
			if v, ok := value.(int); ok {
				rad.task.Option.set_rate(v)
			} else {
				// return errors.New(init_err)
			}
		case "timeout":
			if v, ok := value.(int); ok {
				rad.task.Option.set_timeout(v)
			} else {
				// return errors.New(init_err)
			}
		case "httpx":
			if v, ok := value.(bool); ok {
				rad.task.Option.Httpx = v
			} else {
				// return errors.New(init_err)
			}
		case "fingerDB":
			if v, ok := value.(string); ok {
				rad.task.Option.FingerDB = v
			} else {
				// return errors.New(init_err)
			}
		case "ping":
			if v, ok := value.(bool); ok {
				rad.task.Option.Ping = v
			} else {
				// return errors.New(init_err)
			}
		case "screenshot":
			if v, ok := value.(bool); ok {
				rad.task.Option.Screenshot = v
			} else {
				// return errors.New(init_err)
			}
		case "pool_ping":
			if v, ok := value.(int); ok {
				rad.task.Option.set_pool_ping(v)
			} else {
				// return errors.New(init_err)
			}
		case "pool_scan":
			if v, ok := value.(int); ok {
				rad.task.Option.set_pool_scan(v)
			} else {
				// return errors.New(init_err)
			}
		case "pool_finger":
			if v, ok := value.(int); ok {
				rad.task.Option.set_pool_finger(v)
			} else {
				// return errors.New(init_err)
			}
		case "excludeTimeRange":
			if v, ok := value.(string); ok {
				elements := strings.Split(v, ",")
				if len(elements) == 3 {
					rad.task.Option.ExcludeTimeRange.Daily = elements[0]
					rad.task.Option.ExcludeTimeRange.Begin = elements[1]
					rad.task.Option.ExcludeTimeRange.End = elements[2]
				}
			} else {
				// return errors.New(init_err)
			}

		}
	}
	go rad.task.GenRun()
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody(rad.task.info())
	return nil
}

func (rad *Radar) StatusHandle(ctx *fasthttp.RequestCtx) error {
	info := rad.Info()
	// fmt.Println(string(info))
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody(info)
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
