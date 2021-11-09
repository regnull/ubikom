package db

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/golang/protobuf/proto"
	"github.com/regnull/ubikom/pb"
)

var ErrNotFound = fmt.Errorf("not found")

type Badger struct {
	db *badger.DB
}

func NewBadger(dir string) (*Badger, error) {
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}
	return &Badger{db: db}, nil
}

func (b *Badger) GetMailboxes(user string) ([]*pb.ImapMailbox, error) {
	prefix := []byte(mailboxPrefix(user))
	var mbs []*pb.ImapMailbox
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			mb := &pb.ImapMailbox{}
			err := it.Item().Value(func(v []byte) error {
				return proto.Unmarshal(v, mb)
			})
			if err != nil {
				return err
			}
			mbs = append(mbs, mb)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mbs, nil
}

func (b *Badger) GetMailbox(user string, name string) (*pb.ImapMailbox, error) {
	key := mailboxKey(user, name)
	var mb pb.ImapMailbox
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotFound
			}
			return fmt.Errorf("error getting mailbox: %w", err)
		}
		err = item.Value(func(val []byte) error {
			err := proto.Unmarshal(val, &mb)
			if err != nil {
				return fmt.Errorf("failed to unmarshal mailbox record: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to get mailbox: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &mb, nil
}

func (b *Badger) CreateMailbox(user string, name string) error {
	key := mailboxKey(user, name)
	mb := &pb.ImapMailbox{
		Name: name}
	bb, err := proto.Marshal(mb)
	if err != nil {
		return err
	}

	err = b.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), bb)
		err := txn.SetEntry(e)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (b *Badger) DeleteMailbox(user string, name string) error {
	key := []byte(mailboxKey(user, name))
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *Badger) RenameMailbox(user string, existingName, newName string) error {
	oldKey := []byte(mailboxKey(user, existingName))
	newKey := []byte(mailboxKey(user, newName))

	err := b.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(oldKey)
		if err != nil {
			return err
		}
		mb := &pb.ImapMailbox{}
		err = item.Value(func(val []byte) error {
			return proto.Unmarshal(val, mb)
		})
		if err != nil {
			return err
		}
		mb.Name = newName
		bb, err := proto.Marshal(mb)
		if err != nil {
			return err
		}
		err = txn.SetEntry(badger.NewEntry(newKey, bb))
		if err != nil {
			return err
		}
		err = txn.Delete(oldKey)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func mailboxKey(user, name string) string {
	return "mailbox_" + user + "_" + name
}

func mailboxPrefix(user string) string {
	return "mailbox_" + user + "_"
}
