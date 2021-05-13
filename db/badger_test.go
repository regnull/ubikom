package db

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/ubikom/ecc"
	"github.com/stretchr/testify/assert"
)

func Test_Badger(t *testing.T) {
	assert := assert.New(t)

	// Setup.
	dir, err := ioutil.TempDir("", "ubk-db-test")
	assert.NoError(err)
	db, err := badger.Open(badger.DefaultOptions(dir))
	assert.NoError(err)

	b := NewBadgerDB(db)

	t.Run("Test_Key", func(t *testing.T) {
		key, err := ecc.NewRandomPrivateKey()
		assert.NoError(err)
		publicKey := key.PublicKey()

		err = b.RegisterKey(publicKey)
		assert.NoError(err)

		rec, err := b.GetKey(publicKey)
		assert.NoError(err)
		assert.True(rec.GetRegistrationTimestamp() > 0)
		assert.False(rec.GetDisabled())
		assert.Zero(rec.GetDisabledTimestamp())

		err = b.DisableKey(publicKey, publicKey)
		assert.NoError(err)

		rec, err = b.GetKey(publicKey)
		assert.NoError(err)
		assert.True(rec.GetDisabled())
		assert.True(rec.GetDisabledTimestamp() > 0)
	})

	t.Run("Test_ParentKey", func(t *testing.T) {
		childKey, err := ecc.NewRandomPrivateKey()
		assert.NoError(err)
		childPublicKey := childKey.PublicKey()

		err = b.RegisterKey(childPublicKey)
		assert.NoError(err)

		parentKey, err := ecc.NewRandomPrivateKey()
		assert.NoError(err)
		parentPublicKey := parentKey.PublicKey()
		err = b.RegisterKeyParent(childPublicKey, parentPublicKey)
		assert.NoError(err)

		// Make sure public key was successfully registered.
		rec, err := b.GetKey(childPublicKey)
		assert.NoError(err)
		assert.EqualValues(1, len(rec.GetParentKey()))
		assert.True(parentPublicKey.EqualSerializedCompressed(rec.GetParentKey()[0]))

		// Make sure some random guy can't disable the child.
		someKey, err := ecc.NewRandomPrivateKey()
		assert.NoError(err)
		somePublicKey := someKey.PublicKey()
		err = b.DisableKey(childPublicKey, somePublicKey)
		assert.Error(err)

		// Make sure parent can disable the child.
		err = b.DisableKey(childPublicKey, parentPublicKey)
		assert.NoError(err)

		// Make sure child is really disabled.
		rec, err = b.GetKey(childPublicKey)
		assert.NoError(err)
		assert.True(rec.GetDisabled())
		assert.True(parentPublicKey.EqualSerializedCompressed(rec.GetDisabledBy()))
	})

	t.Run("Test_RegisterName", func(t *testing.T) {
		childKey, err := ecc.NewRandomPrivateKey()
		assert.NoError(err)
		childPublicKey := childKey.PublicKey()

		err = b.RegisterKey(childPublicKey)
		assert.NoError(err)

		parentKey, err := ecc.NewRandomPrivateKey()
		assert.NoError(err)
		parentPublicKey := parentKey.PublicKey()
		err = b.RegisterKeyParent(childPublicKey, parentPublicKey)
		assert.NoError(err)

		err = b.RegisterName(childPublicKey, "bob")
		assert.NoError(err)

		// Confirm the registration.
		k, err := b.GetName("bob")
		assert.True(k.Equal(childPublicKey))

		// Make sure some random guy can't change the name.
		someKey, err := ecc.NewRandomPrivateKey()
		assert.NoError(err)
		somePublicKey := someKey.PublicKey()

		err = b.RegisterName(somePublicKey, "bob")
		assert.Error(err)

		// The original key can change the registration, although
		// right now it won't do anything.
		err = b.RegisterName(childPublicKey, "bob")
		assert.NoError(err)

		// Parent can also change the registration, in this case,
		// the name is re-registered to the parent.
		err = b.RegisterName(parentPublicKey, "bob")
		assert.NoError(err)

		// Make sure the name registration is updated.
		k, err = b.GetName("bob")
		assert.True(parentPublicKey.Equal(k))
	})

	// Tear down.
	os.RemoveAll(dir)
}
