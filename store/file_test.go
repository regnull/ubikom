package store

import (
	"os"
	"testing"

	"github.com/regnull/ubikom/ecc"
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

	fileStore := NewFile(dir)

	pk1, _ := ecc.NewRandomPrivateKey()
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
