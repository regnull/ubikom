package protoutil

import (
	"testing"

	"github.com/regnull/easyecc"
	"github.com/stretchr/testify/assert"
)

func Test_CreateSigned(t *testing.T) {
	assert := assert.New(t)

	key, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	content := []byte("something to be signed")
	signed, err := CreateSigned(key, content)
	assert.NoError(err)
	assert.NotNil(signed)

	assert.True(VerifySignature(signed.Signature, signed.Key, content))
}
