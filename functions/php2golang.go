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
	"regexp"
	"strings"
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
	var b strings.Builder
	for _, ch := range str {
		if ch == '\\' || ch == '"' || ch == '\'' {
			b.WriteRune('\\')
		}
		b.WriteRune(ch)
	}
	return b.String()
}

func Md5(str string) string {
	_md5 := md5.Sum([]byte(str))
	return hex.EncodeToString(_md5[:])
}

func Sha1(str string) string {
	hash := sha1.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

func Sha256(str []byte) string {
	hash := sha256.New()
	hash.Write(str)
	return hex.EncodeToString(hash.Sum(nil))
}

func GenHMAC256(ciphertext, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(ciphertext))
	hmac := mac.Sum(nil)
	return hmac
}
