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

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	mb, err := b.GetMailbox("foo", "bar", privateKey)
	assert.Nil(mb)
	assert.EqualValues(ErrNotFound, err)

	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "bar"}, privateKey))
	mb, err = b.GetMailbox("foo", "bar", privateKey)
	assert.NotNil(mb)
	assert.NoError(err)
	assert.EqualValues("bar", mb.GetName())
}

func Test_GetMailboxes(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "mb1"}, privateKey))
	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "mb2"}, privateKey))
	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "mb3"}, privateKey))

	mbs, err := b.GetMailboxes("foo", privateKey)
	assert.NoError(err)
	assert.EqualValues(4, len(mbs)) // Our mailboxes, plus inbox.

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

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "bar"}, privateKey))
	assert.NoError(b.RenameMailbox("foo", "bar", "baz", privateKey))

	mb, err := b.GetMailbox("foo", "baz", privateKey)
	assert.NoError(err)
	assert.NotNil(mb)
	assert.EqualValues("baz", mb.GetName())
}

func Test_DeleteMailbox(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	assert.NoError(b.CreateMailbox("foo", &pb.ImapMailbox{Name: "bar"}, privateKey))
	assert.NoError(b.DeleteMailbox("foo", "bar", privateKey))

	mb, err := b.GetMailbox("foo", "bar", privateKey)
	assert.EqualValues(ErrNotFound, err)
	assert.Nil(mb)
}

func Test_SubscribeUnsubscribe(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	s, err := b.Subscribed("foo", "bar", privateKey)
	assert.NoError(err)
	assert.False(s)

	assert.NoError(b.Subscribe("foo", "bar", privateKey))
	s, err = b.Subscribed("foo", "bar", privateKey)
	assert.NoError(err)
	assert.True(s)

	assert.NoError(b.Unsubscribe("foo", "bar", privateKey))
	s, err = b.Subscribed("foo", "bar", privateKey)
	assert.NoError(err)
	assert.False(s)
}

func Test_Info(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	mbid, err := b.IncrementMailboxID("foo", privateKey)
	assert.NoError(err)
	assert.EqualValues(1000, mbid)

	mbid, err = b.IncrementMailboxID("foo", privateKey)
	assert.NoError(err)
	assert.EqualValues(1001, mbid)

	err = b.CreateMailbox("foo", &pb.ImapMailbox{
		Name:           "bar",
		Uid:            1001,
		NextMessageUid: 1000,
	}, privateKey)
	assert.NoError(err)

	msgid, err := b.IncrementMessageID("foo", "bar", privateKey)
	assert.NoError(err)
	assert.EqualValues(1000, msgid)

	msgid, err = b.IncrementMessageID("foo", "bar", privateKey)
	assert.NoError(err)
	assert.EqualValues(1001, msgid)
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
