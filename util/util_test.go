package util

import (
	"fmt"
	"testing"
	"time"

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

func Test_NowUint32(t *testing.T) {
	assert := assert.New(t)

	now := NowUint32()
	assert.True(now > 1636658188)
}

func Test_ClearFlag(t *testing.T) {
	assert := assert.New(t)

	flags := []string{"foo", "bar", "baz", "xyz"}
	newFlags := ClearFlag(flags, "zzz")
	assert.Equal(flags, newFlags)

	newFlags = ClearFlag(flags, "baz")
	assert.EqualValues(3, len(newFlags))
	assert.Contains(newFlags, "foo")
	assert.Contains(newFlags, "bar")
	assert.Contains(newFlags, "xyz")
	assert.NotContains(newFlags, "baz")
}

func Test_FileNameNoExtension(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("foo", FileNameNoExtension("foo"))
	assert.Equal("foo", FileNameNoExtension("foo.bar"))
	assert.Equal("foo", FileNameNoExtension("/bar/baz/foo.xyz"))
}

func Test_FixName(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("foo", FixName("foo"))
	assert.Equal("foo", FixName("FoO"))
	assert.Equal("foo", FixName(" fOo  "))
	assert.Equal("foo", FixName("foo@bar.com"))
}

func Test_GetPrivateKeyFromNameAndPassword(t *testing.T) {
	assert := assert.New(t)

	pk := GetPrivateKeyFromNameAndPassword("foo", "bar")
	bitcoinAddress, _ := pk.PublicKey().BitcoinAddress()
	assert.Equal("1M6DhqJEyo6XVfrVH7qvrAGPyj4tE38UFU", bitcoinAddress)

	pk = GetPrivateKeyFromNameAndPassword(" fOo@zzz.xxx   ", "bar")
	bitcoinAddress, _ = pk.PublicKey().BitcoinAddress()
	assert.Equal("1M6DhqJEyo6XVfrVH7qvrAGPyj4tE38UFU", bitcoinAddress)
}

func Test_TimeMs(t *testing.T) {
	assert := assert.New(t)

	ms := NowMs()
	tm := TimeFromMs(ms)

	assert.True(time.Since(tm) < time.Second*5)
}

func Test_Hash160(t *testing.T) {
	assert := assert.New(t)

	h := Hash160([]byte("hello there"))
	assert.Equal("598f9fd736a8b4ae157504f20f8f0c64e11b95fa", fmt.Sprintf("%x", h))
}

func Test_GetDefaultKeyLocation(t *testing.T) {
	assert := assert.New(t)

	location, err := GetDefaultKeyLocation()
	assert.NoError(err)
	assert.True(len(location) > 0)
}
