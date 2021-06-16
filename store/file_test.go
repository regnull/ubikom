package store

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/regnull/ubikom/pb"
	"github.com/stretchr/testify/assert"
)

func Test_File_StoreGetRemove(t *testing.T) {
	assert := assert.New(t)
	store, cleanup, err := createTestFileStore()
	assert.NoError(err)
	defer cleanup()
	testGetRemove(t, store)
}

func Test_File_GetAll(t *testing.T) {
	assert := assert.New(t)
	store, cleanup, err := createTestFileStore()
	assert.NoError(err)
	defer cleanup()
	testGetAll(t, store)
}

func containsMessage(messages []*pb.DMSMessage, message *pb.DMSMessage) bool {
	for _, m := range messages {
		if bytes.Equal(m.Content, message.GetContent()) {
			return true
		}
	}
	return false
}

type CleanupFunc func()

func createTestFileStore() (Store, CleanupFunc, error) {
	dir, err := os.MkdirTemp("", "ubikom_filestore_test")
	if err != nil {
		return nil, func() {}, err
	}
	fileStore := NewFile(dir, time.Hour)
	return fileStore, func() { os.RemoveAll(dir) }, nil
}
