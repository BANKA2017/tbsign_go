// Code copied from https://www.php2golang.com. DO NOT EDIT.
// Code copied from https://www.php2golang.com. DO NOT EDIT.
// Code copied from https://www.php2golang.com. DO NOT EDIT.

package _function

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
)

func HtmlSpecialchars(html string) string {
	reg, err := regexp.Compile(`<([\w]+)(\s*[\w]+=([\w]+|"[^"]+"))*>([\S\s]*)<[/]?([\w]+)>`)
	if err != nil {
		return html
	}

	ret := html
	for reg.MatchString(ret) {
		ret = reg.ReplaceAllString(ret, "$4")
	}
	return ret
}

func Addslashes(str string) string {
	var tmpRune []rune
	strRune := []rune(str)
	for _, ch := range strRune {
		switch ch {
		case []rune{'\\'}[0], []rune{'"'}[0], []rune{'\''}[0]:
			tmpRune = append(tmpRune, []rune{'\\'}[0])
			tmpRune = append(tmpRune, ch)
		default:
			tmpRune = append(tmpRune, ch)
		}
	}
	return string(tmpRune)
}

func Md5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Sha1(str string) string {
	hash := sha1.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

func Sha256(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

func GenHMAC256(ciphertext, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(ciphertext))
	hmac := mac.Sum(nil)
	return hmac
}
