package _function

import (
	"regexp"
	"strings"
)

func VerifyEmail(email string) bool {
	if email == "" {
		return false
	}
	subEmailStr := strings.Split(email, "@")
	if len(subEmailStr) != 2 {
		return false
	}

	if len(subEmailStr[0]) > 64 {
		return false
	}

	return len(regexp.MustCompile(`(?m)^[\w\.\-]+@\w+(?:[\.\-]\w+)*$`).FindAllString(email, -1)) == 1
}
