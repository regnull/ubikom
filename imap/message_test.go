package imap

import (
	"testing"

	"github.com/regnull/ubikom/pb"
	"github.com/stretchr/testify/assert"
)

func Test_NewMessageFromProto(t *testing.T) {
	assert := assert.New(t)

	p := &pb.ImapMessage{
		Content:           []byte("content"),
		Flag:              []string{"someFlag"},
		ReceivedTimestamp: 12345,
		Size:              555,
		Uid:               1001,
	}
	m := NewMessageFromProto(p)
	assert.EqualValues("content", m.Body)
	assert.EqualValues([]string{"someFlag"}, m.Flags)
	assert.EqualValues(12345, m.Date.UnixNano()/1000000)
	assert.EqualValues(555, m.Size)
	assert.EqualValues(1001, m.Uid)
}

func Test_MessageToProto(t *testing.T) {
	assert := assert.New(t)

	p := &pb.ImapMessage{
		Content:           []byte("content"),
		Flag:              []string{"someFlag"},
		ReceivedTimestamp: 12345,
		Size:              555,
		Uid:               1001,
	}
	m := NewMessageFromProto(p)
	p1 := m.ToProto()
	assert.Equal(p, p1)
}
