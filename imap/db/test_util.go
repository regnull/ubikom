package db

import (
	"fmt"
	"os"
)

// CreateTestBadgerStore creates a temporary empty Badger store for testing.
// It returns the badger store, cleanup function which must be called to remove
// the test files, and error.
func CreateTestBadgerStore() (*Badger, func(), error) {
	dir, err := os.MkdirTemp("", "ubikom_badgerstore_test")
	if err != nil {
		return nil, func() {}, err
	}

	fmt.Printf("creating badger store at %s\n", dir)
	store, err := NewBadger(dir, 0)
	if err != nil {
		return nil, func() {}, err
	}
	return store, func() {
		fmt.Printf("cleaning up badger store at %s\n", dir)
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("error during cleanup: %s\n", err)
		}
	}, nil
}
