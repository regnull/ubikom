package store

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Badger_StoreGetRemove(t *testing.T) {
	assert := assert.New(t)
	store, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()
	testGetRemove(t, store)
}

func Test_Badger_GetAll(t *testing.T) {
	assert := assert.New(t)
	store, cleanup, err := createTestBadgerStore()
	assert.NoError(err)
	defer cleanup()
	testGetAll(t, store)
}

func createTestBadgerStore() (Store, CleanupFunc, error) {
	dir, err := os.MkdirTemp("", "ubikom_badgerstore_test")
	if err != nil {
		return nil, func() {}, err
	}

	store, err := NewBadger(dir, time.Hour)
	if err != nil {
		return nil, func() {}, err
	}
	return store, func() { os.RemoveAll(dir) }, nil
}
