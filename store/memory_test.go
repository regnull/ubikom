package store

import (
	"fmt"
	"testing"

	"github.com/regnull/ubikom/pb"
	"github.com/stretchr/testify/assert"
)

func Test_Memory_SaveGetNext(t *testing.T) {
	assert := assert.New(t)

	msg := &pb.DMSMessage{
		Sender:   "alice",
		Receiver: "bob",
		Content:  []byte("the message"),
	}
	store := NewMemory()
	err := store.Save(msg, []byte("123"))
	assert.NoError(err)

	msg1, err := store.GetNext([]byte("123"))
	assert.NoError(err)
	assert.Equal(msg, msg1)
}

func Test_Memory_GetAll(t *testing.T) {
	assert := assert.New(t)

	store := NewMemory()
	var expectedMsgs []*pb.DMSMessage
	for i := 0; i < 10; i++ {
		msg := &pb.DMSMessage{
			Sender:   "alice",
			Receiver: "bob",
			Content:  []byte(fmt.Sprintf("message%d", i)),
		}
		err := store.Save(msg, []byte("123"))
		assert.NoError(err)
		expectedMsgs = append(expectedMsgs, msg)
	}

	msgs, err := store.GetAll([]byte("123"))
	assert.NoError(err)
	assert.True(len(msgs) == 10)
	for _, msg := range msgs {
		assert.Contains(expectedMsgs, msg)
	}
}

func Test_Memory_Remove(t *testing.T) {
	assert := assert.New(t)

	store := NewMemory()
	for i := 0; i < 10; i++ {
		msg := &pb.DMSMessage{
			Sender:   "alice",
			Receiver: "bob",
			Content:  []byte(fmt.Sprintf("message%d", i)),
		}
		err := store.Save(msg, []byte("123"))
		assert.NoError(err)
	}

	msgs, err := store.GetAll([]byte("123"))
	assert.NoError(err)
	assert.True(len(msgs) == 10)

	for i := 0; i < 10; i++ {
		err := store.Remove(msgs[0], []byte("123"))
		msgs = msgs[1:]
		assert.NoError(err)
		msgs1, err := store.GetAll([]byte("123"))
		assert.NoError(err)
		assert.True(len(msgs1) == 9-i)
	}
}
