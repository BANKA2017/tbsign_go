package _function

import (
	"bytes"
	"encoding/json"
)

func JsonDecode[T any](jsonByte []byte, template *T) error {
	if err := json.Unmarshal(jsonByte, template); err != nil {
		return err
	}
	return nil
}

func JsonEncode[T any](data T) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(data)
	return bytes.TrimRight(buffer.Bytes(), "\n"), err
}
