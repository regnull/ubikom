package db

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/regnull/easyecc"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoio"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

const (
	keyPrefix     = "pkey_"
	namePrefix    = "name_"
	maxParentKeys = 1 // For now, only a single parent is allowed.
	addressPrefix = "address_"
)

type BadgerDB struct {
	db *badger.DB
}

func NewBadgerDB(db *badger.DB) *BadgerDB {
	return &BadgerDB{db: db}
}

func (b *BadgerDB) RegisterKey(publicKey *easyecc.PublicKey) error {
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
	return err
}

func (b *BadgerDB) RegisterKeyParent(childKey *easyecc.PublicKey, parentKey *easyecc.PublicKey) error {
	childBase58 := base58.Encode(childKey.SerializeCompressed())
	childDbKey := keyPrefix + childBase58
	err := b.db.Update(func(txn *badger.Txn) error {
		log.Debug().Str("parent", parentKey.Address()).Str("child", childKey.Address()).Msg("registering child key")

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
			if bytes.Compare(parentKey.SerializeCompressed(), p) == 0 {
				// This key is already registered as a parent.
				return ErrRecordExists
			}
		}

		keyRec.ParentKey = append(keyRec.ParentKey, parentKey.SerializeCompressed())
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
	return err
}

func (b *BadgerDB) RegisterName(originator, target *easyecc.PublicKey, name string) error {
	name = strings.ToLower(name)
	dbKey := namePrefix + name
	err := b.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbKey))
		var prev *easyecc.PublicKey
		if item != nil {
			// If the name is already registered, we can change the registration,
			// as long as we are authorized to do so.
			var previousKeyBytes [33]byte
			err = item.Value(func(val []byte) error {
				if copy(previousKeyBytes[:], val) != 33 {
					return fmt.Errorf("invalid name registration record")
				}
				return nil
			})
			if err != nil {
				return err
			}
			prev, err = easyecc.NewPublicFromSerializedCompressed(previousKeyBytes[:])
			if err != nil {
				return err
			}
		}
		authorized, err := CheckRegisterNameAuthorization(txn, originator, target, prev)
		if err != nil {
			return err
		}
		if !authorized {
			return ErrNotAuthorized
		}

		err = txn.Set([]byte(dbKey), target.SerializeCompressed())
		return err
	})
	return err
}

func (b *BadgerDB) RegisterAddress(originator, target *easyecc.PublicKey, name string, protocol pb.Protocol, address string) error {
	name = strings.ToLower(name)
	err := b.db.Update(func(txn *badger.Txn) error {
		// If the target has a parent, it must be the parent who sends the request.
		targetParent, err := GetParent(txn, target)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
		}
		if targetParent != nil && !originator.Equal(targetParent) {
			return ErrNotAuthorized
		}

		// Make sure the requested name is registered to this key.
		nameKey := namePrefix + name
		item, err := txn.Get([]byte(nameKey))
		if err != nil {
			return fmt.Errorf("error getting name, %w", err)
		}
		if item == nil {
			return fmt.Errorf("name %s is not registered", name)
		}

		// This is the key associated with the name. It must match the target key.
		var registeredKeyBytes []byte
		err = item.Value(func(val []byte) error {
			registeredKeyBytes = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return fmt.Errorf("error getting name, %w", err)
		}

		registeredKey, err := easyecc.NewPublicFromSerializedCompressed(registeredKeyBytes)
		if err != nil {
			return err
		}

		if !registeredKey.Equal(target) {
			return ErrNotAuthorized
		}

		addressKey := addressPrefix + name + "_" + protocol.String()
		err = txn.Set([]byte(addressKey), []byte(address))
		return err
	})
	return err
}

func (b *BadgerDB) DisableKey(originator *easyecc.PublicKey, key *easyecc.PublicKey) error {
	publicKeyBase58 := base58.Encode(key.SerializeCompressed())
	dbKey := keyPrefix + publicKeyBase58
	err := b.db.Update(func(txn *badger.Txn) error {

		// If the target has a parent, it must be the parent who sends the request.
		keyParent, err := GetParent(txn, key)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
		}
		if keyParent != nil && !originator.Equal(keyParent) {
			return ErrNotAuthorized
		}
		// Retrieve the key.

		item, err := txn.Get([]byte(dbKey))
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

		// Key if the originator key is authorized to make changes.
		// For this operation to be authorized, the originator must be the key itself,
		// or its parent.

		authorized := false
		if originator.Equal(key) {
			authorized = true
		} else {
			for _, parentKey := range keyRec.GetParentKey() {
				if originator.EqualSerializedCompressed(parentKey) {
					authorized = true
					break
				}
			}
		}
		if !authorized {
			return ErrNotAuthorized
		}

		keyRec.Disabled = true
		keyRec.DisabledTimestamp = util.NowMs()
		keyRec.DisabledBy = originator.SerializeCompressed()

		// Update key registration.

		recBytes, err := proto.Marshal(keyRec)
		if err != nil {
			return fmt.Errorf("error marshaling key record: %w", err)
		}

		err = txn.Set([]byte(dbKey), recBytes)
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

func (b *BadgerDB) GetKey(key *easyecc.PublicKey) (*pb.KeyRecord, error) {
	dbKey := keyPrefix + base58.Encode(key.SerializeCompressed())
	keyRec := &pb.KeyRecord{}
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotFound
			}
			return fmt.Errorf("error getting key record: %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		err = item.Value(func(val []byte) error {
			err := proto.Unmarshal(val, keyRec)
			if err != nil {
				return fmt.Errorf("failed to unmarshal key record: %w", err)
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keyRec, nil
}

func (b *BadgerDB) GetName(name string) (*easyecc.PublicKey, error) {
	name = strings.ToLower(name)
	var keyBytes []byte
	err := b.db.View(func(txn *badger.Txn) error {
		nameKey := namePrefix + name
		item, err := txn.Get([]byte(nameKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotFound
			}
			return fmt.Errorf("error getting name, %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		err = item.Value(func(val []byte) error {
			keyBytes = append([]byte{}, val...)
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	key, err := easyecc.NewPublicFromSerializedCompressed(keyBytes)
	if err != nil {
		// Invalid key.
		return nil, err
	}
	return key, nil
}

func (b *BadgerDB) GetAddress(name string, protocol pb.Protocol) (string, error) {
	var addressBytes []byte
	err := b.db.View(func(txn *badger.Txn) error {
		addressKey := addressPrefix + name + "_" + protocol.String()
		item, err := txn.Get([]byte(addressKey))
		if err != nil {
			return fmt.Errorf("error getting name, %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		err = item.Value(func(val []byte) error {
			addressBytes = append([]byte{}, val...)
			return nil
		})
		return err
	})
	if err != nil {
		return "", err
	}
	return string(addressBytes), nil
}

func CheckSelfOrParent(txn *badger.Txn, originator, target *easyecc.PublicKey) (bool, error) {
	if originator.Equal(target) {
		return true, nil
	}

	targetBase58 := base58.Encode(target.SerializeCompressed())
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
		if originator.EqualSerializedCompressed(p) {
			return true, nil
		}
	}
	return false, nil
}

func GetParent(txn *badger.Txn, key *easyecc.PublicKey) (*easyecc.PublicKey, error) {
	targetBase58 := base58.Encode(key.SerializeCompressed())
	dbKey := keyPrefix + targetBase58
	item, err := txn.Get([]byte(dbKey))
	if err != nil {
		return nil, fmt.Errorf("error getting key record: %w", err)
	}
	if item == nil {
		return nil, fmt.Errorf("target key not found")
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
		return nil, fmt.Errorf("invalid key record: %w", err)
	}
	if len(keyRec.GetParentKey()) == 0 {
		// No parent key.
		return nil, nil
	}
	if len(keyRec.GetParentKey()) > 1 {
		return nil, fmt.Errorf("invalid key record")
	}
	parent, err := easyecc.NewPublicFromSerializedCompressed(keyRec.GetParentKey()[0])
	if err != nil {
		return nil, fmt.Errorf("invalid parent key: %w", err)
	}
	return parent, nil
}

func CheckRegisterNameAuthorization(txn *badger.Txn, originator, target, prev *easyecc.PublicKey) (bool, error) {
	var prevParent *easyecc.PublicKey
	if prev != nil {
		var err error
		prevParent, err = GetParent(txn, prev)
		if err != nil {
			return false, err
		}
	}

	targetParent, err := GetParent(txn, target)
	if err != nil {
		return false, err
	}

	// 1. Check if the originator has the authority.

	// If the target has a parent, only the parent can request the change.
	if targetParent != nil && !originator.Equal(targetParent) {
		return false, nil
	}

	// The request must be for the same key, or for the child key.
	if prev != nil {
		if !originator.Equal(prev) && !originator.Equal(prevParent) {
			return false, nil
		}
	}

	// 2. Check if the change can be done.

	// Can assign to a new key.
	if prev == nil {
		return true, nil
	}

	// Can re-assign name to the same key (noop).
	if prev.Equal(target) {
		return true, nil
	}

	// Can assign from child to  parent.
	if prevParent != nil && prevParent.Equal(target) {
		return true, nil
	}

	// Can assign from parent to child.
	if targetParent != nil && targetParent.Equal(prev) {
		return true, nil
	}

	// Can assign to another child.
	if targetParent != nil && targetParent.Equal(prevParent) {
		return true, nil
	}

	// Nothing else can be done.
	return false, nil
}

func (b *BadgerDB) WriteKeys(w protoio.Writer, cutoffTime uint64) error {
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(keyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			name := string(item.Key())
			publicKeyBase58 := name[len(keyPrefix):]
			keyRec := &pb.KeyRecord{}
			err := item.Value(func(val []byte) error {
				err := proto.Unmarshal(val, keyRec)
				if err != nil {
					return fmt.Errorf("failed to unmarshal key record: %w", err)
				}
				return nil
			})
			if err != nil {
				return err
			}
			// if keyRec.RegistrationTimestamp > int64(cutoffTime) {
			// 	continue
			// }
			rec := &pb.DBRecord{
				Name: publicKeyBase58,
				Payload: &pb.DBRecord_Key{
					Key: keyRec}}
			err = w.Write(rec)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
