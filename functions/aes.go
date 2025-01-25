package _function

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"
)

// GCM Encryption/Decryption
func AES256GCMEncrypt[T string | []byte](plaintext T, key []byte) ([]byte, error) {
	strByte := []byte{}
	var err error

	switch any(plaintext).(type) {
	case string:
		strByte = []byte(any(plaintext).(string))
	case []byte:
		strByte = any(plaintext).([]byte)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce, _ := RandomTokenBuilder(12)

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, strByte, nil)

	return append(nonce, ciphertext...), nil
}

func AES256GCMDecrypt[T string | []byte](ciphertext T, key []byte) ([]byte, error) {
	strByte := []byte{}
	var err error

	switch any(ciphertext).(type) {
	case string:
		strByte, err = base64.RawURLEncoding.DecodeString(strings.ReplaceAll(any(ciphertext).(string), "=", ""))
		if err != nil {
			return nil, err
		}
	case []byte:
		strByte = any(ciphertext).([]byte)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := strByte[:12]
	strByte = strByte[12:]

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, strByte, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
