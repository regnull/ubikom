package imap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewMailbox(t *testing.T) {
	assert := assert.New(t)

	mb := NewMailbox("foo", "bar", nil)
	assert.EqualValues("foo", mb.User())
	assert.EqualValues("bar", mb.Name())
	status, err := mb.Status(nil)
	assert.NoError(err)
	assert.True(status.UidValidity > 1000000000)
}
