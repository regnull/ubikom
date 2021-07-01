package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateName(t *testing.T) {
	assert := assert.New(t)

	assert.True(ValidateName("foobar"))
	assert.True(ValidateName("foo123"))
	assert.True(ValidateName("12345678"))
	assert.True(ValidateName("foo-bar"))

	assert.False(ValidateName("x"))
	assert.False(ValidateName("%%%%"))
	assert.False(ValidateName("some-name-here-$$$"))
	assert.False(ValidateName("some-very-long-name-here-yes-sir-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890"))
}

func Test_GetConfigFromArgs(t *testing.T) {
	assert := assert.New(t)

	c := GetConfigFromArgs([]string{"whatever", "--config", "foo"})
	assert.EqualValues("foo", c)

	c = GetConfigFromArgs([]string{"whatever", "--config"})
	assert.EqualValues("", c)

	c = GetConfigFromArgs([]string{"foo", "bar", "baz"})
	assert.EqualValues("", c)

	c = GetConfigFromArgs([]string{"whatever", "--config=blah"})
	assert.EqualValues("blah", c)
}
