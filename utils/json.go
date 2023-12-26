package utils

import (
	"encoding/json"
	"log"
)

func MustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Panicf("json.MustMarshal failed: %v", err)
	}
	return data
}
