package ecc

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PublicKey_SerializeCompressed(t *testing.T) {
	assert := assert.New(t)

	privateKey := NewPrivateKey(big.NewInt(5001))
	publicKey := privateKey.PublicKey()
	serialized := publicKey.SerializeCompressed()
	assert.EqualValues("0357a4f368868a8a6d572991e484e664810ff14c05c0fa023275251151fe0e53d1", fmt.Sprintf("%x", serialized))
	assert.True(true)
}
