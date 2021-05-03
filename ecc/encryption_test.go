package ecc

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
)

func TestEncryption(t *testing.T) {
	assert := assert.New(t)

	alicePrivateKey, err := NewRandomPrivateKey()
	assert.NoError(err)

	bobPrivateKey, err := NewRandomPrivateKey()
	assert.NoError(err)

	key1x, key1y := btcec.S256().ScalarMult(alicePrivateKey.PublicKey().X(), alicePrivateKey.PublicKey().Y(),
		bobPrivateKey.Secret().Bytes())

	key2x, key2y := btcec.S256().ScalarMult(bobPrivateKey.PublicKey().X(), bobPrivateKey.PublicKey().Y(),
		alicePrivateKey.Secret().Bytes())

	assert.Equal(key1x, key2x)
	assert.Equal(key1y, key2y)
}

func TestEncryptDecrypt(t *testing.T) {
	assert := assert.New(t)

	alicePrivateKey, err := NewRandomPrivateKey()
	assert.NoError(err)

	bobPrivateKey, err := NewRandomPrivateKey()
	assert.NoError(err)

	message := "attack at dawn"
	ciphertext, err := alicePrivateKey.Encrypt([]byte(message), bobPrivateKey.PublicKey())
	fmt.Printf("%x\n", ciphertext)
	assert.NoError(err)

	plaintext, err := bobPrivateKey.Decrypt(ciphertext, alicePrivateKey.PublicKey())
	assert.NoError(err)

	assert.True(bytes.Equal([]byte(message), plaintext))
}
