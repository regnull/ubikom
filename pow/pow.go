package pow

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"math/rand"
	"time"
)

// Compute computes proof of work for the given chunk of data.
func Compute(data []byte, zeros int) []byte {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	start := r.Int63()
	nonce := big.NewInt(start)
	for {
		b := bytes.Join([][]byte{data, nonce.Bytes()}, nil)
		h := sha256.Sum256(b)
		if verifyLeadingZeros(h[:], zeros) {
			return nonce.Bytes()
		}
		nonce.Add(nonce, big.NewInt(1))
	}
}

// Verify verifies proof of work.
func Verify(data []byte, nonce []byte, zeros int) bool {
	b := bytes.Join([][]byte{data, nonce}, nil)
	h := sha256.Sum256(b)
	return verifyLeadingZeros(h[:], zeros)
}

func verifyLeadingZeros(data []byte, zeros int) bool {
	bytes := zeros / 8
	bits := zeros % 8
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
