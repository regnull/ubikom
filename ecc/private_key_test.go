package ecc

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PrivateKey_NewRandom(t *testing.T) {
	assert := assert.New(t)

	pk, err := NewRandomPrivateKey()
	assert.NoError(err)
	assert.NotNil(pk)
}

func Test_PrivateKey_Save(t *testing.T) {
	assert := assert.New(t)

	pk, err := NewRandomPrivateKey()
	assert.NoError(err)

	dir, err := ioutil.TempDir("", "pktest")
	assert.NoError(err)

	fileName := path.Join(dir, "private_key")
	err = pk.Save(fileName)
	assert.NoError(err)

	fi, err := os.Stat(fileName)
	assert.NoError(err)
	assert.EqualValues(32, fi.Size())

	assert.NoError(os.RemoveAll(dir))
}

func Test_PrivateKey_Load(t *testing.T) {
	assert := assert.New(t)

	pk, err := NewRandomPrivateKey()
	assert.NoError(err)

	dir, err := ioutil.TempDir("", "pktest")
	assert.NoError(err)

	fileName := path.Join(dir, "private_key")
	err = pk.Save(fileName)
	assert.NoError(err)

	pkCopy, err := LoadPrivateKey(fileName)
	assert.NoError(err)
	assert.NotNil(pkCopy)
	assert.EqualValues(pk.privateKey.D, pkCopy.privateKey.D)
	assert.EqualValues(pk.privateKey.PublicKey.X, pkCopy.privateKey.PublicKey.X)
	assert.EqualValues(pk.privateKey.PublicKey.Y, pkCopy.privateKey.PublicKey.Y)

	assert.NoError(os.RemoveAll(dir))
}
