package port

import (
	_ "embed"
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/vela-ssoc/vela-radar/util"
)

type Scanner interface {
	Close()
	Wait()
	Scan(ip net.IP, dst uint16) error
	WaitLimiter() error
}

// OpenIpPort retChan
type OpenIpPort struct {
	Ip       net.IP
	Port     uint16
	Service  string
	HttpInfo *HttpInfo
}

func (op OpenIpPort) String() string {
	buf := strings.Builder{}
	buf.WriteString(op.Ip.String())
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(int(op.Port)))
	if op.Service != "" {
		buf.WriteString(" ")
		buf.WriteString(op.Service)
	}
	if op.HttpInfo != nil {
		buf.WriteString("\n")
		buf.WriteString(op.HttpInfo.String())
	}
	return buf.String()
}

// Option ...
type Option struct {
	Rate    int    // 每秒速度限制, 单位: s, 会在1s内平均发送, 相当于每个包之间的延迟
	Timeout int    // TCP连接响应延迟, 单位: ms
	NextHop string // pcap dev name
}

// HttpInfo Http服务基础信息
type HttpInfo struct {
	StatusCode    int      `json:"status_code"     bson:"status_code"`     // 响应状态码
	ContentLength int      `json:"content_length"  bson:"content_length"`  // 响应包大小
	URL           string   `json:"url"             bson:"url"`             // HTTP URL
	Location      string   `json:"location"        bson:"location"`        // 301/302 重定向地址
	Title         string   `json:"title"           bson:"title"`           // 网站 title
	Server        string   `json:"server"          bson:"server"`          // HTTP Header 中的 server 字段
	Body          string   `json:"body"            bson:"body"`            // HTTP body
	Header        string   `json:"header"          bson:"header"`          // HTTP Header 的数据
	FaviconMH3    string   `json:"favicon_mh3"     bson:"favicon_mh3"`     // favicon 的 mh3, mh3 一种比 md5 更快的算法
	FaviconMD5    string   `json:"favicon_md5"     bson:"favicon_md5"`     // favicon 的 md5
	ScreenshotURL string   `json:"screenshot_url"  bson:"screenshot_url"`  // 网站截图的 URL
	Fingerprints  []string `json:"fingerprints"    bson:"fingerprints"`    // 识别到的 web 指纹
	TLSCommonName string   `json:"tls_common_name" bson:"tls_common_name"` // TLS 证书的 CommonName
	TLSDNSNames   []string `json:"tls_dns_names"   bson:"tls_dns_names"`   // TLS 证书的 DNSName
}

func (hi *HttpInfo) String() string {
	if hi == nil {
		return ""
	}
	// var buf strings.Builder
	// buf.WriteString(fmt.Sprintf("[HttpInfo]%s StatusCode:%d ContentLen:%d Title:%s ", hi.URL, hi.StatusCode, hi.ContentLength, hi.Title))
	// if hi.Location != "" {
	// 	buf.WriteString("Location:" + hi.Location + " ")
	// }
	// if hi.TLSCommonName != "" {
	// 	buf.WriteString("TlsCN:" + hi.TLSCommonName + " ")
	// }
	// if len(hi.TLSDNSNames) > 0 {
	// 	buf.WriteString("TlsDNS:" + strings.Join(hi.TLSDNSNames, ",") + " ")
	// }
	// if hi.Server != "" {
	// 	buf.WriteString("Server:" + hi.Server + " ")
	// }
	// if len(hi.Fingerprints) != 0 {
	// 	buf.WriteString(fmt.Sprintf("Fingers:%s ", hi.Fingerprints))
	// }
	// return buf.String()

	return util.ToJsonStr(hi)
}

// ParsePortRangeStr 解析端口字符串
func ParsePortRangeStr(portStr string) (out [][]uint16, err error) {
	portsStrGroup := strings.Split(portStr, ",")
	var portsStrGroup3 []string
	var portStart, portEnd uint64
	for _, portsStrGroup2 := range portsStrGroup {
		if portsStrGroup2 == "top1000" {
			continue
		} else if portsStrGroup2 == "top5000" {
			continue
		} else if portsStrGroup2 == "top200" {
			continue
		}
		portsStrGroup3 = strings.Split(portsStrGroup2, "-")
		portStart, err = strconv.ParseUint(portsStrGroup3[0], 10, 16)
		if err != nil {
			return
		}
		portEnd = portStart
		if len(portsStrGroup3) == 2 {
			portEnd, err = strconv.ParseUint(portsStrGroup3[1], 10, 16)
		}
		if err != nil {
			return
		}
		out = append(out, []uint16{uint16(portStart), uint16(portEnd)})
	}
	return
}

// IsInPortRange 判断port是否在端口范围里
func IsInPortRange(port uint16, portRanges [][]uint16) bool {
	for _, portRange := range portRanges {
		if port >= portRange[0] && port <= portRange[1] {
			return true
		}
	}
	return false
}

// ShuffleParseAndMergeTopPorts shuffle parse portStr and merge TopTcpPorts
func ShuffleParseAndMergeTopPorts(portStr string) (ports []uint16, err error) {
	if portStr == "" {
		ports = TopTcpPorts_1000

		// ports = TopTcpPorts
		return
	}
	TopTcpPorts := TopTcpPorts_1000
	var portRanges [][]uint16
	portRanges, err = ParsePortRangeStr(portStr)
	if err != nil {
		return
	}
	// 优先发送top 1000端口
	selectTopPort := make(map[uint16]struct{}) // TopPort
	hasTopStr_200 := strings.Contains(portStr, "top200")
	hasTopStr_1000 := strings.Contains(portStr, "top1000")
	hasTopStr_5000 := strings.Contains(portStr, "top5000")
	if hasTopStr_200 {
		TopTcpPorts = TopTcpPorts_200
	} else if hasTopStr_1000 {
		TopTcpPorts = TopTcpPorts_1000
	} else if hasTopStr_5000 {
		TopTcpPorts = TopTcpPorts_5000
	}
	for _, _port := range TopTcpPorts {
		if hasTopStr_5000 || IsInPortRange(_port, portRanges) {
			selectTopPort[_port] = struct{}{}
			ports = append(ports, _port)
		} else if hasTopStr_1000 || IsInPortRange(_port, portRanges) {
			selectTopPort[_port] = struct{}{}
			ports = append(ports, _port)
		} else if hasTopStr_200 || IsInPortRange(_port, portRanges) {
			selectTopPort[_port] = struct{}{}
			ports = append(ports, _port)
		}
	}
	selectPort := make(map[uint16]struct{}) // OtherPort
	for _, portRange := range portRanges {
		var ok bool
		for _port := portRange[0]; _port <= portRange[1]; _port++ {
			if _port == 0 {
				continue
			}
			if _, ok = selectTopPort[_port]; ok {
				continue
			} else if _, ok = selectPort[_port]; ok {
				continue
			}
			selectPort[_port] = struct{}{}
			ports = append(ports, _port)
			if _port == 65535 {
				break
			}
		}
	}
	if len(ports) == 0 {
		err = errors.New("ports len is 0")
		return
	}
	// 端口随机化
	skip := uint64(len(selectTopPort)) // 跳过Top
	_ports := make([]uint16, len(ports))
	copy(_ports, ports)
	sf := util.NewShuffle(uint64(len(ports)) - skip)
	if sf != nil {
		for i := skip; i < uint64(len(_ports)); i++ {
			ports[i] = _ports[skip+sf.Get(i-skip)]
		}
	}
	return
}
