package imap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewBackend_Login(t *testing.T) {
	assert := assert.New(t)

	b := NewBackend(nil, nil, nil, "foo", "bar", nil)
	assert.NotNil(b)
	user, err := b.Login(nil, "foo", "bar")
	assert.NoError(err)
	assert.NotNil(user)
}
