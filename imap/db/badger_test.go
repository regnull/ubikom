package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateGetMailbox(t *testing.T) {
	assert := assert.New(t)

	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	mb, err := b.GetMailbox("foo", "bar")
	assert.Nil(mb)
	assert.EqualValues(ErrNotFound, err)

	assert.NoError(b.CreateMailbox("foo", "bar"))
	mb, err = b.GetMailbox("foo", "bar")
	assert.NotNil(mb)
	assert.NoError(err)
}

func createTestBadgerStore() (*Badger, func(), error) {
	dir, err := os.MkdirTemp("", "ubikom_badgerstore_test")
	if err != nil {
		return nil, func() {}, err
	}

	store, err := NewBadger(dir)
	if err != nil {
		return nil, func() {}, err
	}
	return store, func() { os.RemoveAll(dir) }, nil
}
