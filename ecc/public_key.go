package ecc

import (
	"crypto/ecdsa"
	"math/big"
)

// PublicKey represents elliptic curve cryptography private key.
type PublicKey struct {
	publicKey *ecdsa.PublicKey
}

// SerializeCompressed returns the private key serialized in SEC compressed format.
func (pbk *PublicKey) SerializeCompressed() []byte {
	buf := make([]byte, 33)
	if new(big.Int).Mod(pbk.publicKey.Y, big.NewInt(2)).Cmp(big.NewInt(0)) == 0 {
		// Even.
		buf[0] = 0x02
	} else {
		// Odd.
		buf[0] = 0x03
	}
	yBytes := pbk.publicKey.X.Bytes()
	for i := 1; i < 33; i++ {
		buf[i] = yBytes[i-1]
	}
	return buf
}
