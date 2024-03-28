package _function

// from chatcpt
func RemoveLeadingZeros(data []byte) []byte {
	for i := 0; i < len(data); i++ {
		if data[i] != 0 {
			return data[i:]
		}
	}
	return []byte{0}
}
