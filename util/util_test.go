package util

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/pb/mocks"
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

func Test_GetKeyFromNamePassword(t *testing.T) {
	assert := assert.New(t)

	// Test key from name as hash 256.
	name := "spongebob"
	password := "krabbypatty"

	nameHash := Hash256([]byte(name))
	expectedKey := easyecc.NewPrivateKeyFromPassword([]byte(password), nameHash)

	ctx := context.Background()
	lookup := new(mocks.LookupServiceClient)
	lookup.On("LookupKey", ctx, &pb.LookupKeyRequest{
		Key: expectedKey.PublicKey().SerializeCompressed()}).Return(&pb.LookupKeyResponse{}, nil)

	actualKey, err := GetKeyFromNamePassword(ctx, name, password, lookup)
	assert.NoError(err)
	assert.NotNil(actualKey)
	assert.Equal(expectedKey, actualKey)

	// Make sure we can strip the domain.
	lookup.On("LookupKey", ctx, &pb.LookupKeyRequest{
		Key: expectedKey.PublicKey().SerializeCompressed()}).Return(&pb.LookupKeyResponse{}, nil)
	actualKey, err = GetKeyFromNamePassword(ctx, name+"@bikinibottom.com", password, lookup)
	assert.NoError(err)
	assert.NotNil(actualKey)
	assert.Equal(expectedKey, actualKey)

	// Test key from name as base-58 encoded random value.
	var saltArr [8]byte
	_, err = rand.Read(saltArr[:])
	assert.NoError(err)
	salt := saltArr[:]
	name = base58.Encode(salt[:])
	candidateKey := easyecc.NewPrivateKeyFromPassword([]byte(password), Hash256([]byte(name)))
	expectedKey = easyecc.NewPrivateKeyFromPassword([]byte(password), salt)
	lookup.On("LookupKey", ctx, &pb.LookupKeyRequest{
		Key: candidateKey.PublicKey().SerializeCompressed()}).Return(nil, errors.New("not found"))
	lookup.On("LookupKey", ctx, &pb.LookupKeyRequest{
		Key: expectedKey.PublicKey().SerializeCompressed()}).Return(&pb.LookupKeyResponse{}, nil)
	actualKey, err = GetKeyFromNamePassword(ctx, name, password, lookup)
	assert.NoError(err)
	assert.NotNil(actualKey)
	assert.Equal(expectedKey, actualKey)
}

func Test_NowUint32(t *testing.T) {
	assert := assert.New(t)

	now := NowUint32()
	assert.True(now > 1636658188)
}

func Test_FileNameNoExtension(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("foo", FileNameNoExtension("foo"))
	assert.Equal("foo", FileNameNoExtension("foo.bar"))
	assert.Equal("foo", FileNameNoExtension("/bar/baz/foo.xyz"))
}
