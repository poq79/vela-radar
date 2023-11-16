package util

import "encoding/json"

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
