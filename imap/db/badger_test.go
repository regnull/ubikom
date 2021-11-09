package db

import (
	"fmt"
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
	assert.EqualValues("bar", mb.GetName())
}

func Test_GetMailboxes(t *testing.T) {
	assert := assert.New(t)

	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	assert.NoError(b.CreateMailbox("foo", "mb1"))
	assert.NoError(b.CreateMailbox("foo", "mb2"))
	assert.NoError(b.CreateMailbox("foo", "mb3"))

	mbs, err := b.GetMailboxes("foo")
	assert.NoError(err)
	assert.EqualValues(3, len(mbs))

	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("mb%d", i)
		found := false
		for _, mb := range mbs {
			if mb.GetName() == name {
				found = true
				break
			}
		}
		assert.True(found)
	}
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
