package db

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/regnull/easyecc"

	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/ubikom/pb"
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
		key, err := easyecc.NewRandomPrivateKey()
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
		childKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		childPublicKey := childKey.PublicKey()

		err = b.RegisterKey(childPublicKey)
		assert.NoError(err)

		parentKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		parentPublicKey := parentKey.PublicKey()
		err = b.RegisterKeyParent(childPublicKey, parentPublicKey)
		assert.NoError(err)

		// Make sure another key can't be registered as a parent.
		anotherKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		anotherPublicKey := anotherKey.PublicKey()
		err = b.RegisterKeyParent(childPublicKey, anotherPublicKey)
		assert.Error(err)

		// Make sure public key was successfully registered.
		rec, err := b.GetKey(childPublicKey)
		assert.NoError(err)
		assert.EqualValues(1, len(rec.GetParentKey()))
		assert.True(parentPublicKey.EqualSerializedCompressed(rec.GetParentKey()[0]))

		// Make sure some random guy can't disable the child.
		someKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		somePublicKey := someKey.PublicKey()
		err = b.DisableKey(somePublicKey, childPublicKey)
		assert.Error(err)

		// Make sure parent can disable the child.
		err = b.DisableKey(parentPublicKey, childPublicKey)
		assert.NoError(err)

		// Make sure child is really disabled.
		rec, err = b.GetKey(childPublicKey)
		assert.NoError(err)
		assert.True(rec.GetDisabled())
		assert.True(parentPublicKey.EqualSerializedCompressed(rec.GetDisabledBy()))
	})

	t.Run("Test_RegisterName", func(t *testing.T) {
		childKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		childPublicKey := childKey.PublicKey()

		err = b.RegisterKey(childPublicKey)
		assert.NoError(err)

		parentKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		parentPublicKey := parentKey.PublicKey()
		err = b.RegisterKey(parentPublicKey)
		assert.NoError(err)

		err = b.RegisterKeyParent(childPublicKey, parentPublicKey)
		assert.NoError(err)

		err = b.RegisterName(parentPublicKey, childPublicKey, "bob")
		assert.NoError(err)

		// Confirm the registration.
		k, err := b.GetName("bob")
		assert.True(k.Equal(childPublicKey))

		// Make sure some random guy can't change the name.
		someKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		somePublicKey := someKey.PublicKey()

		err = b.RegisterName(somePublicKey, somePublicKey, "bob")
		assert.Error(err)

		// The original key can't change the registration (but parent can).
		err = b.RegisterName(childPublicKey, childPublicKey, "bob")
		assert.Error(err)

		// Parent can re-register key to another child.
		anotherChildKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		anotherChildPublicKey := anotherChildKey.PublicKey()
		err = b.RegisterKey(anotherChildPublicKey)
		assert.NoError(err)
		err = b.RegisterKeyParent(anotherChildPublicKey, parentPublicKey)
		assert.NoError(err)

		err = b.RegisterName(parentPublicKey, anotherChildPublicKey, "bob")
		assert.NoError(err)

		// Make sure the name registration is updated.
		k, err = b.GetName("bob")
		assert.True(anotherChildPublicKey.Equal(k))
	})

	t.Run("Test_Address", func(t *testing.T) {
		childKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		childPublicKey := childKey.PublicKey()

		err = b.RegisterKey(childPublicKey)
		assert.NoError(err)

		parentKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		parentPublicKey := parentKey.PublicKey()
		err = b.RegisterKey(parentPublicKey)
		assert.NoError(err)

		err = b.RegisterKeyParent(childPublicKey, parentPublicKey)
		assert.NoError(err)

		err = b.RegisterName(parentPublicKey, childPublicKey, "patrick")
		assert.NoError(err)

		// Test address registration.
		err = b.RegisterAddress(parentPublicKey, childPublicKey, "patrick", pb.Protocol_PL_DMS, "localhost:1122")
		assert.NoError(err)

		address, err := b.GetAddress("patrick", pb.Protocol_PL_DMS)
		assert.NoError(err)
		assert.EqualValues("localhost:1122", address)

		// Make sure some random guy can't change the registration.
		someKey, err := easyecc.NewRandomPrivateKey()
		assert.NoError(err)
		somePublicKey := someKey.PublicKey()
		err = b.RegisterAddress(somePublicKey, childPublicKey, "patrick", pb.Protocol_PL_DMS, "localhost:3344")
		assert.Error(err)

		// Make sure the original key can update the registration.
		err = b.RegisterAddress(parentPublicKey, childPublicKey, "patrick", pb.Protocol_PL_DMS, "localhost:5566")
		assert.NoError(err)

		address, err = b.GetAddress("patrick", pb.Protocol_PL_DMS)
		assert.NoError(err)
		assert.EqualValues("localhost:5566", address)
	})

	// Tear down.
	os.RemoveAll(dir)
}
