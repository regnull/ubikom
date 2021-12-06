package imap

import (
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap/db"
	"github.com/regnull/ubikom/mail"
	"github.com/stretchr/testify/assert"
)

func Test_NewMailbox(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	badger, cleanup, err := db.CreateTestBadgerStore()
	assert.NoError(err)
	defer cleanup()
	mb, err := NewMailbox("foo", "bar", badger, nil, nil, privateKey)
	assert.NoError(err)
	assert.NoError(mb.Save())
	assert.EqualValues("foo", mb.User())
	assert.EqualValues("bar", mb.Name())
	status, err := mb.Status([]imap.StatusItem{imap.StatusUidValidity})
	assert.NoError(err)
	assert.True(status.UidValidity >= 1000)
}

func Test_MailboxSubscribed(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	badger, cleanup, err := db.CreateTestBadgerStore()
	assert.NoError(err)
	defer cleanup()
	mb, err := NewMailbox("foo", "bar", badger, nil, nil, privateKey)
	assert.NoError(err)
	assert.NoError(mb.Save())

	s, err := badger.Subscribed("foo", "bar", privateKey)
	assert.NoError(err)
	assert.False(s)

	mb.SetSubscribed(true)
	mb.Save()
	s, err = badger.Subscribed("foo", "bar", privateKey)
	assert.NoError(err)
	assert.True(s)
}

func Test_MailboxInfo(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	badger, cleanup, err := db.CreateTestBadgerStore()
	assert.NoError(err)
	defer cleanup()
	mb, err := NewMailbox("foo", "bar", badger, nil, nil, privateKey)
	assert.NoError(err)
	assert.NoError(mb.Save())

	info, err := mb.Info()
	assert.NoError(err)
	assert.Nil(info.Attributes)
	assert.Equal(DELIMITER, info.Delimiter)
	assert.Equal("bar", info.Name)
}

func Test_MailboxStatus(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	badger, cleanup, err := db.CreateTestBadgerStore()
	assert.NoError(err)
	defer cleanup()
	mb, err := NewMailbox("foo", "bar", badger, nil, nil, privateKey)
	assert.NoError(err)
	assert.NoError(mb.Save())

	lit := getLiteral("bob", "joe", "hello", "how is it going?")
	err = mb.CreateMessage([]string{imap.RecentFlag}, time.Now(), lit)
	assert.NoError(err)

	status, err := mb.Status([]imap.StatusItem{imap.StatusMessages, imap.StatusUidNext,
		imap.StatusUidValidity, imap.StatusRecent, imap.StatusUnseen})
	assert.NoError(err)
	assert.EqualValues(1, status.Messages)
	assert.EqualValues(1001, status.UidNext)
	assert.EqualValues(1000, status.UidValidity)
	assert.EqualValues(1, status.Recent)
	assert.EqualValues(1, status.Unseen)
}

type dummyLiteral struct {
	*strings.Reader
	len int
}

func (d *dummyLiteral) Len() int {
	return d.len
}

func getLiteral(to, from string, subject string, body string) imap.Literal {
	s := mail.NewMessage(to, from, subject, body)
	buf := strings.NewReader(s)
	return &dummyLiteral{Reader: buf, len: len(s)}
}
