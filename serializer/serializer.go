package serializer

import "encoding/json"

func Bytes(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
