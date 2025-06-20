package _function

// from chatgpt
func RemoveLeadingZeros(data []byte) []byte {
	for i := range data {
		if data[i] != 0 {
			return data[i:]
		}
	}
	return []byte{0}
}
