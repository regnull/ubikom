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
	return nil, fmt.Errorf("not implemented")
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
	return fmt.Errorf("not implemented")
}

func (b *Badger) RenameMailbox(user string, existingName, newName string) error {
	return fmt.Errorf("not implemented")
}

func mailboxKey(user, name string) string {
	return "mailbox_" + user + "_" + name
}
