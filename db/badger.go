package db

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

const (
	keyPrefix     = "pkey_"
	namePrefix    = "name_"
	maxParentKeys = 16
)

type BadgerDB struct {
	db *badger.DB
}

func NewBadgerDB(db *badger.DB) *BadgerDB {
	return &BadgerDB{db: db}
}

func (b *BadgerDB) RegisterKey(publicKey *ecc.PublicKey) error {
	publicKeyBase58 := base58.Encode(publicKey.SerializeCompressed())
	log.Info().Str("key", publicKeyBase58).Msg("registering public key")
	dbKey := keyPrefix + publicKeyBase58

	err := b.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbKey))
		if item != nil {
			return ErrRecordExists
		}

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

		// Register key address as name.

		nameKey := namePrefix + publicKey.Address()
		err = txn.Set([]byte(nameKey), publicKey.SerializeCompressed())
		if err != nil {
			return err
		}

		log.Info().Str("name", publicKey.Address()).Msg("name registered")
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}
	return nil
}

func (b *BadgerDB) RegisterKeyParent(child []byte, parent []byte) error {
	childBase58 := base58.Encode(child)
	childDbKey := keyPrefix + childBase58
	err := b.db.Update(func(txn *badger.Txn) error {

		// Retrieve the key.

		item, err := txn.Get([]byte(childDbKey))
		if err != nil {
			return fmt.Errorf("error getting key record: %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		keyRec := &pb.KeyRecord{}
		err = item.Value(func(val []byte) error {
			err := proto.Unmarshal(val, keyRec)
			if err != nil {
				return fmt.Errorf("failed to unmarshal key record: %w", err)
			}
			return nil
		})

		// Add parent to the key record.

		for _, p := range keyRec.GetParentKey() {
			if bytes.Compare(parent, p) == 0 {
				// This key is already registered as a parent.
				return ErrRecordExists
			}
		}

		keyRec.ParentKey = append(keyRec.ParentKey, parent)
		if len(keyRec.GetParentKey()) > maxParentKeys {
			return fmt.Errorf("too many parent keys")
		}

		// Update key registration.

		recBytes, err := proto.Marshal(keyRec)
		if err != nil {
			return fmt.Errorf("error marshaling key record: %w", err)
		}

		err = txn.Set([]byte(childDbKey), recBytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error registering key parent: %w", err)
	}

	return nil
}

func CheckSelfOrParent(txn *badger.Txn, originator, target []byte) (bool, error) {
	if bytes.Compare(originator, target) == 0 {
		return true, nil
	}

	targetBase58 := base58.Encode(target)
	key := keyPrefix + targetBase58
	item, err := txn.Get([]byte(key))
	if err != nil {
		return false, fmt.Errorf("error getting key record: %w", err)
	}
	if item == nil {
		return false, fmt.Errorf("target key not found")
	}

	keyRec := &pb.KeyRecord{}
	err = item.Value(func(val []byte) error {
		err := proto.Unmarshal(val, keyRec)
		if err != nil {
			return fmt.Errorf("failed to unmarshal key record: %w", err)
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("invalid key record: %w", err)
	}

	for _, p := range keyRec.GetParentKey() {
		if bytes.Compare(p, originator) == 0 {
			return true, nil
		}
	}
	return false, nil
}
