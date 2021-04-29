package ecc

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SignAndVerify(t *testing.T) {
	assert := assert.New(t)

	data := []byte("hello there")
	for i := 0; i < 1000; i++ {
		hash := sha256.Sum256(data)
		pkey, err := NewRandomPrivateKey()
		assert.NoError(err)
		sig, err := pkey.Sign(hash[:])
		assert.NoError(err)
		assert.True(sig.Verify(pkey.PublicKey(), hash[:]))
	}
}
