package protoutil

import (
	"testing"
	"time"

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

func Test_VerifyIdentity(t *testing.T) {
	assert := assert.New(t)

	key, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	ts := time.Now()
	signed := IdentityProof(key, ts)
	assert.NoError(VerifyIdentity(signed, ts, 10.0))

	ts1 := ts.Add(time.Minute)
	assert.Error(VerifyIdentity(signed, ts1, 10.0))
}
