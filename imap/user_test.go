package imap

import (
	"os"
	"testing"

	"github.com/emersion/go-imap/backend"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap/db"
	"github.com/stretchr/testify/assert"
)

func Test_NewUser(t *testing.T) {
	assert := assert.New(t)

	u := NewUser("bob", nil, nil, nil, nil)
	assert.EqualValues("bob", u.Username())
}

func Test_ListMailboxes(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	b, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()

	u := NewUser("bob", b, privateKey, nil, nil)
	mailboxes, err := u.ListMailboxes(false)
	assert.NoError(err)

	// Even if no mailboxes were explicitly created, mailbox must be there.
	assert.EqualValues(1, len(mailboxes))
	assert.EqualValues("INBOX", mailboxes[0].Name())

	assert.NoError(u.CreateMailbox("foo"))
	assert.NoError(u.CreateMailbox("bar"))
	mailboxes, err = u.ListMailboxes(false)
	assert.NoError(err)
	assert.EqualValues(3, len(mailboxes))
	assert.True(containsMailbox(mailboxes, "foo"))
	assert.True(containsMailbox(mailboxes, "bar"))
	assert.False(containsMailbox(mailboxes, "zoo"))

	assert.NoError(u.DeleteMailbox("bar"))
	mailboxes, err = u.ListMailboxes(false)
	assert.NoError(err)
	assert.EqualValues(2, len(mailboxes))
	assert.True(containsMailbox(mailboxes, "foo"))
	assert.False(containsMailbox(mailboxes, "bar"))
	assert.True(containsMailbox(mailboxes, "INBOX"))

	// Can't delete inbox!
	assert.Error(u.DeleteMailbox("INBOX"))

	// Only subscribed mailboxes must be returned.
	mailboxes, err = u.ListMailboxes(true)
	assert.NoError(err)
	assert.EqualValues(0, len(mailboxes))

	assert.NoError(b.Subscribe("bob", "foo", privateKey))
	mailboxes, err = u.ListMailboxes(true)
	assert.NoError(err)
	assert.EqualValues(1, len(mailboxes))
	assert.True(containsMailbox(mailboxes, "foo"))
}

func containsMailbox(mailboxes []backend.Mailbox, name string) bool {
	for _, mb := range mailboxes {
		if mb.Name() == name {
			return true
		}
	}
	return false
}

func createTestBadgerStore() (*db.Badger, func(), error) {
	dir, err := os.MkdirTemp("", "ubikom_badgerstore_test")
	if err != nil {
		return nil, func() {}, err
	}

	store, err := db.NewBadger(dir)
	if err != nil {
		return nil, func() {}, err
	}
	return store, func() { os.RemoveAll(dir) }, nil
}
