package ecc

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
)

// PublicKey represents elliptic curve cryptography private key.
type PublicKey struct {
	publicKey *ecdsa.PublicKey
}

func NewPublicFromSerializedCompressed(serialized []byte) (*PublicKey, error) {
	if len(serialized) != 33 {
		return nil, fmt.Errorf("invalid serialized compressed public key")
	}

	even := false
	if serialized[0] == 0x02 {
		even = true
	} else if serialized[0] == 0x03 {
		even = false
	} else {
		return nil, fmt.Errorf("invalid serialized compressed public key")
	}
	x := new(big.Int).SetBytes(serialized[1:])
	P := btcec.S256().CurveParams.P
	sqrtExp := new(big.Int).Add(P, big.NewInt(1))
	sqrtExp = sqrtExp.Div(sqrtExp, big.NewInt(4))
	alpha := new(big.Int).Exp(x, big.NewInt(3), P)
	alpha.Add(alpha, btcec.S256().B)
	beta := new(big.Int).Exp(alpha, sqrtExp, P)
	var evenBeta *big.Int
	var oddBeta *big.Int
	if new(big.Int).Div(beta, big.NewInt(2)).Cmp(big.NewInt(0)) == 0 {
		evenBeta = beta
		oddBeta = new(big.Int).Sub(P, beta)
	} else {
		evenBeta = new(big.Int).Sub(P, beta)
		oddBeta = beta
	}
	var y *big.Int
	if even {
		y = evenBeta
	} else {
		y = oddBeta
	}
	return &PublicKey{publicKey: &ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     x,
		Y:     y}}, nil
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
