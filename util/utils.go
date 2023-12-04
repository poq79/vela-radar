package util

import (
	"encoding/json"
	"strings"

	"github.com/vela-ssoc/vela-kit/iputil"
)

func ToJsonStr(i interface{}) string {
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	jsonString := string(jsonBytes)
	return jsonString
}

func ToJsonBytes(i interface{}) []byte {
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		return []byte{}
	}
	return jsonBytes
}

func IpstrWithCommaToMap(ipstr string) map[string]bool {
	result := make(map[string]bool)

	items := strings.Split(ipstr, ",")
	for _, ip := range items {
		if iputil.IsIPv4(ip) {
			result[ip] = true
		} else {
			ip_set, err := iputil.GenIpSet(ip)
			if err != nil {
				continue
			}
			for _, _ip := range ip_set {
				result[_ip.String()] = true
			}
		}
	}
	return result
}
