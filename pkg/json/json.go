package json

import (
	"encoding/json"
	"errors"
)

// Decode ...
func Decode(jsonStr string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	b := []byte(jsonStr)
	var f interface{}
	err := json.Unmarshal(b, &f)
	if err != nil {
		err = errors.New("json.Decode err:" + err.Error())
	} else {
		m = f.(map[string]interface{})
	}
	return m, err
}

// Encode ...
func Encode(jsonMap map[string]interface{}) ([]byte, error) {
	b, err := json.Marshal(jsonMap)
	if err != nil {
		err = errors.New("json.Encode err:" + err.Error())
	}

	return b, err
}
