package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
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

	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "bar"}))
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

	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "mb1"}))
	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "mb2"}))
	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "mb3"}))

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

func Test_RenameMailbox(t *testing.T) {
	assert := assert.New(t)

	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "bar"}))
	assert.NoError(b.RenameMailbox("foo", "bar", "baz"))

	mb, err := b.GetMailbox("foo", "baz")
	assert.NoError(err)
	assert.NotNil(mb)
	assert.EqualValues("baz", mb.GetName())
}

func Test_DeleteMailbox(t *testing.T) {
	assert := assert.New(t)

	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "bar"}))
	assert.NoError(b.DeleteMailbox("foo", "bar"))

	mb, err := b.GetMailbox("foo", "bar")
	assert.EqualValues(ErrNotFound, err)
	assert.Nil(mb)
}

func Test_SubscribeUnsubscribe(t *testing.T) {
	assert := assert.New(t)

	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	s, err := b.Subscribed("foo", "bar")
	assert.NoError(err)
	assert.False(s)

	assert.NoError(b.Subscribe("foo", "bar"))
	s, err = b.Subscribed("foo", "bar")
	assert.NoError(err)
	assert.True(s)

	assert.NoError(b.Unsubscribe("foo", "bar"))
	s, err = b.Subscribed("foo", "bar")
	assert.NoError(err)
	assert.False(s)
}

func Test_Info(t *testing.T) {
	assert := assert.New(t)

	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	mbid, err := b.GetNextMailboxID("foo")
	assert.NoError(err)
	assert.EqualValues(1000, mbid)

	mbid, err = b.GetNextMailboxID("foo")
	assert.NoError(err)
	assert.EqualValues(1001, mbid)

	msgid, err := b.GetNextMessageID("foo")
	assert.NoError(err)
	assert.EqualValues(1000, msgid)

	msgid, err = b.GetNextMessageID("foo")
	assert.NoError(err)
	assert.EqualValues(1001, msgid)
}

func createTestBadgerStore() (*Badger, func(), error) {
	dir, err := os.MkdirTemp("", "ubikom_badgerstore_test")
	if err != nil {
		return nil, func() {}, err
	}

	privateKey, err := easyecc.NewRandomPrivateKey()
	if err != nil {
		return nil, func() {}, err
	}
	store, err := NewBadger(dir, privateKey)
	if err != nil {
		return nil, func() {}, err
	}
	return store, func() { os.RemoveAll(dir) }, nil
}
