package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_VerifyPOW(t *testing.T) {
	assert := assert.New(t)

	data := []byte{0b00001111}
	assert.True(VerifyPOW(data, 4))
	assert.False(VerifyPOW(data, 5))

	data = []byte{0, 0, 0b00111111}
	assert.True(VerifyPOW(data, 8))
	assert.True(VerifyPOW(data, 16))
	assert.True(VerifyPOW(data, 18))
	assert.False(VerifyPOW(data, 19))

	data = []byte{0, 0xFF}
	assert.True(VerifyPOW(data, 8))
	assert.False(VerifyPOW(data, 9))
	assert.False(VerifyPOW(data, 999))
}
