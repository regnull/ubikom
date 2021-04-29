package ecc

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
)

// PrivateKey represents elliptic cryptography private key.
type PrivateKey struct {
	privateKey *ecdsa.PrivateKey
}

// NewRandomPrivateKey creates a new random private key.
func NewRandomPrivateKey() (*PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key, %w", err)
	}
	return &PrivateKey{privateKey: privateKey}, nil
}

// NewPrivateKey returns new private key created from the secret.
func NewPrivateKey(secret *big.Int) *PrivateKey {
	privateKey := &ecdsa.PrivateKey{
		D: secret}
	privateKey.PublicKey.Curve = btcec.S256()
	privateKey.PublicKey.X, privateKey.PublicKey.Y = privateKey.PublicKey.Curve.ScalarBaseMult(secret.Bytes())
	return &PrivateKey{privateKey: privateKey}
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
	pk.PublicKey.Curve = btcec.S256()
	pk.PublicKey.X, pk.PublicKey.Y = pk.PublicKey.Curve.ScalarBaseMult(d.Bytes())

	return &PrivateKey{privateKey: pk}, nil
}

// SavePrivateKey saves the private key to the specified file.
func (pk *PrivateKey) Save(fileName string) error {
	return ioutil.WriteFile(fileName, []byte(pk.privateKey.D.Bytes()), 0600)
}

// PublicKey returns the public key derived from this private key.
func (pk *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{publicKey: &pk.privateKey.PublicKey}
}

func (pk *PrivateKey) Sign(hash []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, pk.privateKey, hash)
	if err != nil {
		return nil, err
	}
	return &Signature{R: r, S: s}, nil
}

// TODO: remove this.
func (pk *PrivateKey) GetKey() *big.Int {
	return pk.privateKey.D
}
