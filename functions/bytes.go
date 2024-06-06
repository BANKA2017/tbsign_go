package _function

import "strings"

// from chatgpt
func RemoveLeadingZeros(data []byte) []byte {
	for i := 0; i < len(data); i++ {
		if data[i] != 0 {
			return data[i:]
		}
	}
	return []byte{0}
}

func AppendStrings(s ...string) string {
	return strings.Join(s, "")
}
