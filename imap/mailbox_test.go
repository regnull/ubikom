package imap

import (
	"testing"

	"github.com/emersion/go-imap"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap/db"
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
