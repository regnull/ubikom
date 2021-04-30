package util

import (
	"crypto/sha256"
	"strings"
	"time"
)

const (
	allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
)

// NowMs returns current time as milliseconds from epoch.
func NowMs() int64 {
	return time.Now().UnixNano() / 1000000
}

// Hash256 does two rounds of SHA256 hashing.
func Hash256(data []byte) []byte {
	h := sha256.Sum256(data)
	h1 := sha256.Sum256(h[:])
	return h1[:]
}

// ValidateName returns true if the name is valid.
func ValidateName(name string) bool {
	if len(name) < 5 || len(name) > 64 {
		return false
	}

	for _, c := range name {
		if !strings.ContainsRune(allowedChars, c) {
			return false
		}
	}
	return true
}
