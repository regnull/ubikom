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
	assert.EqualValues(6, len(mbs)) // Our mailboxes, plus inbox.

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

func Test_MailboxMessageID(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	// First mailbox must have ID of 1000.

	mbid, err := b.GetNextMailboxID("foo", privateKey)
	assert.NoError(err)
	assert.EqualValues(1000, mbid)

	// Increment functions return the next ID, and then increment it.
	mbid, err = b.IncrementMailboxID("foo", privateKey)
	assert.NoError(err)
	assert.EqualValues(1000, mbid)

	mbid, err = b.GetNextMailboxID("foo", privateKey)
	assert.NoError(err)
	assert.EqualValues(1001, mbid)

	mbid, err = b.IncrementMailboxID("foo", privateKey)
	assert.NoError(err)
	assert.EqualValues(1001, mbid)

	mbid, err = b.GetNextMailboxID("foo", privateKey)
	assert.NoError(err)
	assert.EqualValues(1002, mbid)

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

	msgid, err = b.GetNextMessageID("foo", "bar", privateKey)
	assert.NoError(err)
	assert.EqualValues(1002, msgid)

	// Create another mailbox to verify that the next message ID is different.
	err = b.CreateMailbox("foo", &pb.ImapMailbox{
		Name:           "baz",
		Uid:            1002,
		NextMessageUid: 1000,
	}, privateKey)
	assert.NoError(err)

	msgid, err = b.GetNextMessageID("foo", "baz", privateKey)
	assert.NoError(err)
	assert.EqualValues(1000, msgid)

	msgid, err = b.IncrementMessageID("foo", "baz", privateKey)
	assert.NoError(err)
	assert.EqualValues(1000, msgid)

	msgid, err = b.GetNextMessageID("foo", "baz", privateKey)
	assert.NoError(err)
	assert.EqualValues(1001, msgid)
}

func Test_SaveGetMessage(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	msg := &pb.ImapMessage{
		Content:           []byte("this is my message"),
		Flag:              []string{"flag1", "flag2"},
		ReceivedTimestamp: 12345,
		Size:              555,
		Uid:               1001,
	}
	err = b.SaveMessage("foo", 1000, msg, privateKey)
	assert.NoError(err)

	messages, err := b.GetMessages("foo", 1000, privateKey)
	assert.NoError(err)
	assert.EqualValues(1, len(messages))

	msg = messages[0]
	assert.EqualValues("this is my message", string(msg.Content))
	assert.Contains(msg.GetFlag(), "flag1")
	assert.Contains(msg.GetFlag(), "flag2")
	assert.EqualValues(12345, msg.GetReceivedTimestamp())
	assert.EqualValues(555, msg.GetSize())
	assert.EqualValues(1001, msg.GetUid())

	err = b.DeleteMessage("foo", 1000, 1001)
	assert.NoError(err)

	messages, err = b.GetMessages("foo", 1000, privateKey)
	assert.NoError(err)
	assert.EqualValues(0, len(messages))
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
