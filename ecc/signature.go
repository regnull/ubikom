package ecc

import (
	"crypto/ecdsa"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

func (sig *Signature) Verify(key *PublicKey, hash []byte) bool {
	return ecdsa.Verify(key.publicKey, hash, sig.R, sig.S)
}
