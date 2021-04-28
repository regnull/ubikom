package util

import "time"

// NowMs returns current time as milliseconds from epoch.
func NowMs() int64 {
	return time.Now().UnixNano() / 1000000
}

// VerifyPOW returns true if n first bits of data are all zeros.
func VerifyPOW(data []byte, n int) bool {
	bytes := n / 8
	bits := n % 8
	minLengthBytes := bytes
	if bits > 0 {
		minLengthBytes++
	}
	if len(data) < minLengthBytes {
		return false
	}
	for i := 0; i < bytes; i++ {
		if data[0] != 0 {
			return false
		}
		data = data[1:]
	}
	b := data[0]
	for i := 0; i < bits; i++ {
		if b&0x80 != 0 {
			return false
		}
		b <<= 1
	}
	return true
}
