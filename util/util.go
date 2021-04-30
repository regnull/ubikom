package util

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pow"
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

// CreateSignedWithPOW creates a request signed with the given private key and generates POW of the given strength.
func CreateSignedWithPOW(privateKey *ecc.PrivateKey, content []byte, powStrength int) (*pb.SignedWithPow, error) {
	compressedKey := privateKey.PublicKey().SerializeCompressed()

	log.Info().Msg("generating POW...")
	reqPow := pow.Compute(content, powStrength)
	log.Info().Hex("pow", reqPow).Msg("POW found")

	hash := Hash256(content)
	sig, err := privateKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request, %w", err)
	}

	req := &pb.SignedWithPow{
		Content: content,
		Pow:     reqPow,
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		Key: compressedKey,
	}
	return req, nil
}
