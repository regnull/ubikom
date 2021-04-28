package pow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_VerifyLeadingZeros(t *testing.T) {
	assert := assert.New(t)

	data := []byte{0b00001111}
	assert.True(verifyLeadingZeros(data, 4))
	assert.False(verifyLeadingZeros(data, 5))

	data = []byte{0, 0, 0b00111111}
	assert.True(verifyLeadingZeros(data, 8))
	assert.True(verifyLeadingZeros(data, 16))
	assert.True(verifyLeadingZeros(data, 18))
	assert.False(verifyLeadingZeros(data, 19))

	data = []byte{0, 0xFF}
	assert.True(verifyLeadingZeros(data, 8))
	assert.False(verifyLeadingZeros(data, 9))
	assert.False(verifyLeadingZeros(data, 999))
}

func Test_GenerateVerify(t *testing.T) {
	assert := assert.New(t)

	data := []byte("hello there")
	pow := Compute(data, 10)
	assert.True(Verify(data, pow, 10))
	pow[0] += 7
	assert.False(Verify(data, pow, 10))
}
