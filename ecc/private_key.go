package ecc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
)

// PrivateKey represents elliptic cryptography private key.
type PrivateKey struct {
	privateKey *ecdsa.PrivateKey
}

// NewRandomPrivateKey creates a new random private key.
func NewRandomPrivateKey() (*PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key, %w", err)
	}
	return &PrivateKey{privateKey: privateKey}, nil
}

// LoadPrivateKey loads private key from file.
func LoadPrivateKey(fileName string) (*PrivateKey, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	if len(b) != 32 {
		return nil, fmt.Errorf("invalid private key length")
	}

	d := new(big.Int)
	d.SetBytes(b)

	// We don't save public key, instead, we re-construct public key
	// from the private key.
	pk := &ecdsa.PrivateKey{
		D: d}
	pk.PublicKey.Curve = elliptic.P256()
	pk.PublicKey.X, pk.PublicKey.Y = pk.PublicKey.Curve.ScalarBaseMult(d.Bytes())

	return &PrivateKey{privateKey: pk}, nil
}

// SavePrivateKey saves the private key to the specified file.
func (pk *PrivateKey) Save(fileName string) error {
	return ioutil.WriteFile(fileName, []byte(pk.privateKey.D.Bytes()), 0600)
}
