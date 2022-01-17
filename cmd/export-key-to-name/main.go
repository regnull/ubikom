package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	db, err := badger.Open(badger.DefaultOptions("db"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize the database")
	}

	keysByName := make(map[string]string)
	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("name_")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			var keyBytes []byte
			err = item.Value(func(val []byte) error {
				keyBytes = append([]byte{}, val...)
				return nil
			})
			parts := strings.Split(string(k), "_")
			keysByName[parts[1]] = fmt.Sprintf("%0x", keyBytes)
		}
		return nil
	})

	f, err := os.Create("name_to_key.csv")
	w := bufio.NewWriter(f)
	for k, v := range keysByName {
		w.WriteString(k + "," + v + "\n")
	}
	defer f.Close()
}
