package radar

import (
	"encoding/json"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/strutil"
	"github.com/vela-ssoc/vela-radar/port"
	"net"
)

type Service struct {
	IP        net.IP          `json:"ip"`
	Port      uint16          `json:"port"`
	TLS       bool            `json:"tls"`
	Banner    json.RawMessage `json:"banner"`
	Protocol  string          `json:"protocol"`
	Transport string          `json:"transport"`
	Version   string          `json:"version"`
	HttpInfo  *port.HttpInfo  `json:"http_info"`
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
	enc.KV("protocol", s.Protocol)
	enc.KV("transport", s.Transport)
	enc.KV("version", s.Version)
	enc.Raw("banner", s.Banner)
	enc.End("}")
	return enc.Bytes()
}

func (s *Service) WebFinder() {

}
