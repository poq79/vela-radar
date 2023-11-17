package radar

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"
	"github.com/vela-ssoc/vela-kit/vela"
)

const init_err = "[API /api/v1/arr/agent/radar/runscan]Parameter initialization failed. Please check whether the parameter data type is correct--"

func (rad *Radar) define(r vela.Router) {
	r.GET("/api/v1/arr/agent/radar/status", xEnv.Then(func(ctx *fasthttp.RequestCtx) error {
		info := rad.Info()
		fmt.Println(string(info))
		ctx.Response.Header.SetContentType("application/json")
		ctx.Response.SetBody(info)
		return nil
	}))
	r.POST("/api/v1/arr/agent/radar/runscan", xEnv.Then(func(ctx *fasthttp.RequestCtx) error {
		if rad.TaskStatus() == "working" {
			return errors.New("There are already scanning tasks running")
		}
		// 获取请求的 JSON 数据
		body := ctx.PostBody()

		// 解析 JSON 数据
		var data map[string]interface{}
		err := json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println("解析 JSON 数据失败:", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return err
		}
		if v, ok := data["target"].(string); ok && v != "" {
			rad.NewTask(v)
		} else {
			return errors.New(init_err + "target")
		}
		for key, value := range data {
			switch key {
			case "location":
				if v, ok := value.(string); ok {
					rad.task.Option.Location = v
				} else {
					return errors.New(init_err + key)
				}
			case "name":
				if v, ok := value.(string); ok {
					rad.task.Name = v
				} else {
					return errors.New(init_err + key)
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
			case "ping":
				if v, ok := value.(bool); ok {
					rad.task.Option.Ping = v
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
			}
		}
		go rad.task.GenRun()
		ctx.Response.Header.SetContentType("application/json")
		ctx.Response.SetBody(rad.task.info())
		return nil
	}))
}
