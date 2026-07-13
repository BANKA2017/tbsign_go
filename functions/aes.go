package _function

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"strings"

	"github.com/BANKA2017/tbsign_go/share"
)

const tcEncPrefix = "tc:enc:aes256gcm:"

const tcEncVerify = "tc:enc:verify"

const tcEncVerifyAAD = "tc:enc:aad"

func CreateVerifyEncStatus() error {
	val, err := AES256GCMEncrypt(tcEncVerify, share.DataEncryptKeyByte, []byte(tcEncVerifyAAD))
	if err != nil {
		return err
	}

	return SetOption("go_encrypt", val)
}

func VerifyEncStatus() error {
	if len(share.DataEncryptKeyByte) != 32 {
		return errors.New("invalid encrypt key length")
	}

	testValue := GetOption("go_encrypt")

	if testValue == "0" || testValue == "" {
		return errors.New("system never been encrypted")
	}

	dec, err := AES256GCMDecrypt(testValue, share.DataEncryptKeyByte, []byte(tcEncVerifyAAD))
	if err != nil {
		return err
	}

	if string(dec) != tcEncVerify {
		return errors.New("invalid encrypt key")
	}

	return nil
}

// AES256GCM Encryption
// tcenc should be idempotent
func AES256GCMEncrypt(plaintext string, key, aad []byte) (string, error) {
	if strings.HasPrefix(plaintext, tcEncPrefix) {
		return plaintext, nil
	}

	strByte := []byte(plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext, err
	}

	nonce, _ := RandomTokenBuilder(12)

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext, err
	}

	ciphertext := gcm.Seal(nil, nonce, strByte, aad)

	return tcEncPrefix + Base64URLEncode(append(nonce, ciphertext...)), nil
}

func AES256GCMDecrypt(ciphertext string, key, aad []byte) ([]byte, error) {
	if !strings.HasPrefix(ciphertext, tcEncPrefix) {
		return []byte(ciphertext), nil
	}

	strByte, err := Base64URLDecode(strings.TrimPrefix(ciphertext, tcEncPrefix))

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(strByte) < 12 {
		return nil, errors.New("ciphertext too short for nonce")
	}

	nonce := strByte[:12]
	strByte = strByte[12:]

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, strByte, aad)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
