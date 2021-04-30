package util

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/ripemd160"
)

const (
	allowedChars      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	defaultHomeSubDir = ".ubikom"
	defaultKeyFile    = "key"
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

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(sha256(b)).
func Hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
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

func GetDefaultKeyLocation() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory, %w", err)
	}
	dir := path.Join(homeDir, defaultHomeSubDir)
	keyFile := path.Join(dir, defaultKeyFile)
	return keyFile, nil
}
