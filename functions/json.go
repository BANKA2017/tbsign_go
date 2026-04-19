package _function

import (
	"encoding/json"
	"errors"
)

func JsonDecode[T any](jsonByte []byte, template *T) error {
	return json.Unmarshal(jsonByte, template)
}

func JsonEncode[T any](data T) ([]byte, error) {
	return json.Marshal(data)
}

var unmarshalTypeError *json.UnmarshalTypeError

func JsonIsUnmarshalTypeError(err error) bool {
	return errors.As(err, &unmarshalTypeError)
}
