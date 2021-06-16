package store

import (
	"fmt"
	"testing"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func testGetRemove(t *testing.T, store Store) {
	assert := assert.New(t)

	pk1, _ := easyecc.NewRandomPrivateKey()
	key1 := pk1.PublicKey().SerializeCompressed()

	msg, err := store.GetNext(key1)
	assert.NoError(err)
	assert.Nil(msg)

	msg = &pb.DMSMessage{
		Sender:   "foo",
		Receiver: "bar",
		Content:  []byte("hello there"),
	}
	err = store.Save(msg, key1)
	assert.NoError(err)

	msg1, err := store.GetNext(key1)
	assert.NoError(err)
	assert.True(proto.Equal(msg, msg1))

	assert.NoError(store.Remove(msg1, key1))
	msg, err = store.GetNext(key1)
	assert.NoError(err)
	assert.Nil(msg)
}

func testGetAll(t *testing.T, store Store) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	serializedPublicKey := privateKey.PublicKey().SerializeCompressed()

	messages := make([]*pb.DMSMessage, 5)
	for i := 0; i < 5; i++ {
		msg := &pb.DMSMessage{
			Sender:   "foo",
			Receiver: "bar",
			Content:  []byte(fmt.Sprintf("this is message #%d", i)),
		}
		messages[i] = msg
		err = store.Save(msg, serializedPublicKey)
		assert.NoError(err)
	}

	allMessages, err := store.GetAll(serializedPublicKey)
	assert.NoError(err)
	assert.True(len(allMessages) == 5)
	for _, msg := range messages {
		assert.True(containsMessage(allMessages, msg))
	}

	// Delete one of the messages.
	err = store.Remove(messages[3], serializedPublicKey)
	allMessages, err = store.GetAll(serializedPublicKey)
	assert.NoError(err)
	assert.True(len(allMessages) == 4)
	for i, msg := range messages {
		if i == 3 {
			assert.False(containsMessage(allMessages, msg))
		} else {
			assert.True(containsMessage(allMessages, msg))
		}
	}

	// Delete all messages.
	for _, msg := range allMessages {
		assert.NoError(store.Remove(msg, serializedPublicKey))
	}

	allMessages, err = store.GetAll(serializedPublicKey)
	assert.NoError(err)
	assert.True(len(allMessages) == 0)
}
