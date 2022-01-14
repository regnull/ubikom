package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"google.golang.org/protobuf/proto"
)

const (
	keyPrefix  = "pkey_"
	namePrefix = "name_"
)

func main() {
	var (
		dbPath string
		key    string
		name   string
	)
	flag.StringVar(&dbPath, "db-path", "", "db path")
	flag.StringVar(&key, "key", "", "key")
	flag.StringVar(&name, "name", "", "name")
	flag.Parse()

	if dbPath == "" {
		log.Fatal("--db-path must be specified")
	}

	if key == "" {
		log.Fatal("--key must be specified")
	}

	if name == "" {
		log.Fatal("--name must be specified")
	}

	keyData, err := hex.DecodeString(key)
	if err != nil {
		log.Fatal(err)
	}

	publicKey, err := easyecc.NewPublicFromSerializedCompressed(keyData)
	if err != nil {
		log.Fatal(err)
	}

	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		log.Fatal(err)
	}

	// Register key.
	publicKeyBase58 := base58.Encode(publicKey.SerializeCompressed())
	dbKey := keyPrefix + publicKeyBase58

	err = db.Update(func(txn *badger.Txn) error {
		// Register the key.

		rec := &pb.KeyRecord{
			RegistrationTimestamp: util.NowMs()}
		recBytes, err := proto.Marshal(rec)
		if err != nil {
			return fmt.Errorf("error marshaling key record: %w", err)
		}

		err = txn.Set([]byte(dbKey), recBytes)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("key is registered successfully")

	// Register name.

	name = strings.ToLower(name)
	dbKey = namePrefix + name
	log.Printf("updating name: %s", name)
	err = db.Update(func(txn *badger.Txn) error {
		err = txn.Set([]byte(dbKey), publicKey.SerializeCompressed())
		return err
	})

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("name update successfully")

	var keyBytes []byte
	err = db.View(func(txn *badger.Txn) error {
		nameKey := namePrefix + name
		item, err := txn.Get([]byte(nameKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return err
			}
			return fmt.Errorf("error getting name, %w", err)
		}
		if item == nil {
			return fmt.Errorf("item is nil")
		}

		err = item.Value(func(val []byte) error {
			keyBytes = append([]byte{}, val...)
			return nil
		})
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("name is registered to key: %0x", keyBytes)
}
