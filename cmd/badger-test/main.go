package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
)

const (
	defaultHomeSubDir = ".ubikom"
	defaultDBSubDir   = "db"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := path.Join(homeDir, defaultHomeSubDir)
	_ = os.Mkdir(dir, 0700)
	dir = path.Join(dir, defaultDBSubDir)
	_ = os.Mkdir(dir, 0700)
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("answer"), []byte("42"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	var valCopy []byte
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("answer"))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		return err
	})
	fmt.Printf("The answer is: %s\n", valCopy)

	defer db.Close()
}
