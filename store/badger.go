package store

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/ubikom/pb"
	"google.golang.org/protobuf/proto"
)

type Badger struct {
	db  *badger.DB
	ttl time.Duration
}

func NewBadger(dir string, ttl time.Duration) (*Badger, error) {
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}
	return &Badger{db: db, ttl: ttl}, nil
}

func (b *Badger) Save(msg *pb.DMSMessage, receiverKey []byte) error {
	bb, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}
	msgID := fmt.Sprintf("%x", sha256.Sum256(bb))

	dbKey := "msg_" + fmt.Sprintf("%x", receiverKey) + "_" + msgID
	err = b.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(dbKey), bb).WithTTL(b.ttl)
		err := txn.SetEntry(e)
		if err != nil {
			return err
		}
		return nil
	})
	return err

}

func (b *Badger) GetNext(receiverKey []byte) (*pb.DMSMessage, error) {
	prefix := []byte("msg_" + fmt.Sprintf("%x", receiverKey))
	var msg *pb.DMSMessage
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			err := it.Item().Value(func(v []byte) error {
				// TODO: How can this possibly work?
				msg = &pb.DMSMessage{}
				return proto.Unmarshal(v, msg)
			})
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (b *Badger) GetAll(receiverKey []byte) ([]*pb.DMSMessage, error) {
	prefix := []byte("msg_" + fmt.Sprintf("%x", receiverKey))
	var msgs []*pb.DMSMessage
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			msg := &pb.DMSMessage{}
			err := it.Item().Value(func(v []byte) error {
				err := proto.Unmarshal(v, msg)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
			msgs = append(msgs, msg)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return msgs, nil

}

func (b *Badger) Remove(msg *pb.DMSMessage, receiverKey []byte) error {
	bb, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}
	msgID := fmt.Sprintf("%x", sha256.Sum256(bb))

	dbKey := "msg_" + fmt.Sprintf("%x", receiverKey) + "_" + msgID
	err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(dbKey))
	})
	return err
}
