package store

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_File_StoreGetRemove(t *testing.T) {
	assert := assert.New(t)

	dir, err := os.MkdirTemp("", "ubikom_filestore_test")
	assert.NoError(err)
	assert.NotEmpty(dir)
	defer os.RemoveAll(dir)

	fileStore := NewFile(dir, time.Hour)

	pk1, _ := easyecc.NewRandomPrivateKey()
	key1 := pk1.PublicKey().SerializeCompressed()

	msg, err := fileStore.GetNext(key1)
	assert.NoError(err)
	assert.Nil(msg)

	msg = &pb.DMSMessage{
		Sender:   "foo",
		Receiver: "bar",
		Content:  []byte("hello there"),
	}
	err = fileStore.Save(msg, key1)
	assert.NoError(err)

	msg1, err := fileStore.GetNext(key1)
	assert.NoError(err)
	assert.True(proto.Equal(msg, msg1))

	assert.NoError(fileStore.Remove(msg1, key1))
	msg, err = fileStore.GetNext(key1)
	assert.NoError(err)
	assert.Nil(msg)
}

func Test_File_GetAll(t *testing.T) {
	assert := assert.New(t)

	dir, err := os.MkdirTemp("", "ubikom_filestore_test")
	assert.NoError(err)
	assert.NotEmpty(dir)
	defer os.RemoveAll(dir)

	fileStore := NewFile(dir, time.Hour)

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
		err = fileStore.Save(msg, serializedPublicKey)
		assert.NoError(err)
	}

	allMessages, err := fileStore.GetAll(serializedPublicKey)
	assert.NoError(err)
	assert.True(len(allMessages) == 5)
	for _, msg := range messages {
		assert.True(containsMessage(allMessages, msg))
	}

	// Delete one of the messages.
	err = fileStore.Remove(messages[3], serializedPublicKey)
	allMessages, err = fileStore.GetAll(serializedPublicKey)
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
		assert.NoError(fileStore.Remove(msg, serializedPublicKey))
	}

	allMessages, err = fileStore.GetAll(serializedPublicKey)
	assert.NoError(err)
	assert.True(len(allMessages) == 0)
}

func containsMessage(messages []*pb.DMSMessage, message *pb.DMSMessage) bool {
	for _, m := range messages {
		if bytes.Equal(m.Content, message.GetContent()) {
			return true
		}
	}
	return false
}
