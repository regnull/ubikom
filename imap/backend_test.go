package imap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewBackend_Login(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	b := NewBackend(nil, nil, nil, "foo", "bar", nil)
	require.NotNil(b)
	user, err := b.Login(nil, "foo", "bar")
	assert.NoError(err)
	assert.NotNil(user)

	user, err = b.Login(nil, "foo", "baz")
	assert.Error(ErrAccessDenied, err)
	assert.Nil(user)
}
