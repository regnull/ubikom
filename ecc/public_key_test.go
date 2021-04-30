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

func Test_PublicKey_FromSerializedCompressed(t *testing.T) {
	assert := assert.New(t)

	serialized, _ := new(big.Int).SetString("0357a4f368868a8a6d572991e484e664810ff14c05c0fa023275251151fe0e53d1", 16)
	publicKey, err := NewPublicFromSerializedCompressed(serialized.Bytes())
	assert.NoError(err)
	assert.NotNil(publicKey)
	assert.EqualValues("57a4f368868a8a6d572991e484e664810ff14c05c0fa023275251151fe0e53d1", fmt.Sprintf("%064x", publicKey.publicKey.X))
	assert.EqualValues("0d6cc87c5bc29b83368e17869e964f2f53d52ea3aa3e5a9efa1fa578123a0c6d", fmt.Sprintf("%064x", publicKey.publicKey.Y))
}

func Test_PublicKey_Address(t *testing.T) {
	assert := assert.New(t)

	secret, _ := new(big.Int).SetString("12345deadbeef", 16)
	privateKey := NewPrivateKey(secret)
	address := privateKey.PublicKey().Address()
	assert.Equal("1F1Pn2y6pDb68E5nYJJeba4TLg2U7B6KF1", address)
}
