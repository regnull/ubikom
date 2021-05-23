package ecc

import (
	"io/ioutil"
	"math/big"
	"math/rand"
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

func Test_PrivateKey_SerializeDeserialize(t *testing.T) {
	assert := assert.New(t)

	// Confirm that serialization/deserialization work as expected.
	// Serialize/deserialize a bunch of keys.

	r := rand.New(rand.NewSource(123))
	for i := 0; i < 1000; i++ {
		secret := r.Int63()
		privateKey := NewPrivateKey(big.NewInt(secret))
		serialized := privateKey.PublicKey().SerializeCompressed()
		publicKey, err := NewPublicFromSerializedCompressed(serialized)
		assert.NoError(err)
		assert.Equal(privateKey.PublicKey().publicKey.X, publicKey.publicKey.X)
		assert.Equal(privateKey.PublicKey().publicKey.Y, publicKey.publicKey.Y)
	}
}

func Test_PrivateKey_EncryptDecrypt(t *testing.T) {
	assert := assert.New(t)

	key, err := NewRandomPrivateKey()
	assert.NoError(err)

	encrypted, err := key.EncryptKeyWithPassphrase("super secret spies")
	assert.NoError(err)
	assert.NotNil(encrypted)

	key1, err := NewPrivateKeyFromEncryptedWithPassphrase(encrypted, "super secret spies")
	assert.NoError(err)
	assert.True(key1.privateKey.Equal(key.privateKey))
}

func Test_PrivateKey_FromPassword(t *testing.T) {
	assert := assert.New(t)

	key := NewPrivateKeyFromPassword([]byte("super secret spies"), []byte{0x11, 0x22, 0x33, 0x44})
	assert.NotNil(key)
}

func Test_PrivateKey_NewPrivateKeyFromEncryptedWithPassphrase_InvalidData(t *testing.T) {
	assert := assert.New(t)

	key, err := NewPrivateKeyFromEncryptedWithPassphrase([]byte("bad data"), "foo")
	assert.Nil(key)
	assert.Error((err))
}
