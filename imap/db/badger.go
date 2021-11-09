package db

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/ubikom/pb"
)

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
	return nil, fmt.Errorf("not implemented")
}

func (b *Badger) CreateMailbox(user string, name string) error {
	return fmt.Errorf("not implemented")
}

func (b *Badger) DeleteMailbox(user string, name string) error {
	return fmt.Errorf("not implemented")
}

func (b *Badger) RenameMailbox(user string, existingName, newName string) error {
	return fmt.Errorf("not implemented")
}
