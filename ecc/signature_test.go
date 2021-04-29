package ecc

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SignAndVerify(t *testing.T) {
	assert := assert.New(t)

	hashStr := []byte("f4a982e4268b35e069d0cf3bf1026489a4038bb434f2e0d73918ea0cbd771e24")
	hash := make([]byte, 32)
	hex.Decode(hash, hashStr)
	fmt.Printf("hash: %x\n", hash)

	//data := []byte("hello there")
	//	for i := 0; i < 1000; i++ {
	//hash := sha256.Sum256(data)
	secret, _ := new(big.Int).SetString("c5e05b56182c7cef2ecf7edd3f27764095c524ba74db470c2b1838a1a7234bde", 16)
	pkey := NewPrivateKey(secret)
	//		pkey, err := NewRandomPrivateKey()
	//		assert.NoError(err)
	sig, err := pkey.Sign(hash[:])
	fmt.Printf("R: %x, S: %x\n", sig.R, sig.S)
	assert.NoError(err)
	assert.True(sig.Verify(pkey.PublicKey(), hash[:]))
	//	}
}
