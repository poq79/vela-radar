package radar

import (
	"encoding/json"
	"net"

	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/strutil"
	"github.com/vela-ssoc/vela-radar/port"
)

type Service struct {
	IP        net.IP          `json:"ip"         bson:"ip"`   // IP
	Port      uint16          `json:"port"       bson:"port"` // 端口
	Host      string          `json:"host"       bson:"host"` // 主机域名 (扫内网的时候一般为空或者填 IP)
	Location  string          `json:"location"`               // 地理位置 *
	TLS       bool            `json:"tls"`
	Banner    json.RawMessage `json:"banner"`    // tcp服务的banner信息
	Protocol  string          `json:"protocol"`  // 应用层协议
	Transport string          `json:"transport"` // 传输层协议 tcp/udp
	Version   string          `json:"version"`   // 应用(或者协议)版本
	Component []string        `json:"component"` // 组件标签
	Comment   string          `json:"Comment"`   // 备注信息
	HTTPInfo  *port.HttpInfo  `json:"http_info"` // web服务的指纹以及相关信息
}

func (s *Service) String() string                         { return strutil.B2S(s.Bytes()) }
func (s *Service) Type() lua.LValueType                   { return lua.LTObject }
func (s *Service) AssertFloat64() (float64, bool)         { return 0, false }
func (s *Service) AssertString() (string, bool)           { return "", false }
func (s *Service) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (s *Service) Peek() lua.LValue                       { return s }

func (s *Service) Bytes() []byte {
	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.KV("ip", s.IP)
	enc.KV("port", s.Port)
	enc.KV("tls", s.TLS)
	enc.KV("host", s.Host)
	enc.KV("protocol", s.Protocol)
	enc.KV("transport", s.Transport)
	enc.KV("version", s.Version)
	enc.KV("comment", s.Comment)
	enc.KV("component", s.Component)
	enc.Raw("http_info", s.HTTPInfo.Json())
	enc.KV("banner", string([]byte(s.Banner)))
	enc.End("}")
	return enc.Bytes()
}

func (s *Service) WebFinder() {

}
