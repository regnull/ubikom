package db

import (
	"fmt"
	"strings"

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

func getMailboxes(txn *badger.Txn, user string) (*pb.ImapMailboxes, error) {
	item, err := txn.Get(mailboxKey(user))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return &pb.ImapMailboxes{}, nil
		}
		return nil, fmt.Errorf("error getting mailbox: %w", err)
	}

	mailboxes := &pb.ImapMailboxes{}
	err = item.Value(func(val []byte) error {
		err := proto.Unmarshal(val, mailboxes)
		if err != nil {
			return fmt.Errorf("failed to unmarshal mailboxes record: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox: %w", err)
	}
	return mailboxes, nil
}

func (b *Badger) GetMailboxes(user string) ([]*pb.ImapMailbox, error) {
	var mbs []*pb.ImapMailbox
	err := b.db.View(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		mbs = mailboxes.GetMailbox()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mbs, nil
}

func (b *Badger) GetMailbox(user string, name string) (*pb.ImapMailbox, error) {
	var mailboxes *pb.ImapMailboxes
	err := b.db.View(func(txn *badger.Txn) error {
		var err error
		mailboxes, err = getMailboxes(txn, user)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, mb := range mailboxes.GetMailbox() {
		if mb.GetName() == name {
			return mb, nil
		}
	}
	return nil, ErrNotFound
}

func (b *Badger) CreateMailbox(user string, name string) error {
	mb := &pb.ImapMailbox{
		Name: name}

	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		for _, mb := range mailboxes.GetMailbox() {
			if mb.GetName() == name {
				return fmt.Errorf("mailbox already exists")
			}
		}
		mailboxes.Mailbox = append(mailboxes.Mailbox, mb)
		bb, err := proto.Marshal(mailboxes)
		if err != nil {
			return fmt.Errorf("failed to marshal mailboxes: %w", err)
		}
		err = txn.Set(mailboxKey(user), bb)
		if err != nil {
			return fmt.Errorf("failed to save mailboxes: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (b *Badger) DeleteMailbox(user string, name string) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		var newMailboxes []*pb.ImapMailbox
		for _, mb := range mailboxes.GetMailbox() {
			if mb.GetName() == name {
				continue
			}
			newMailboxes = append(newMailboxes, mb)
		}
		if len(newMailboxes) == len(mailboxes.GetMailbox()) {
			return fmt.Errorf("mailbox not found")
		}
		mailboxes.Mailbox = newMailboxes
		bb, err := proto.Marshal(mailboxes)
		if err != nil {
			return fmt.Errorf("failed to marshal mailboxes: %w", err)
		}
		err = txn.Set(mailboxKey(user), bb)
		if err != nil {
			return fmt.Errorf("failed to save mailboxes: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *Badger) RenameMailbox(user string, existingName, newName string) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		for _, mb := range mailboxes.GetMailbox() {
			if strings.HasPrefix(mb.GetName(), existingName) {
				n := newName + mb.GetName()[len(existingName):]
				mb.Name = n
			}
		}
		bb, err := proto.Marshal(mailboxes)
		if err != nil {
			return fmt.Errorf("failed to marshal mailboxes: %w", err)
		}
		err = txn.Set(mailboxKey(user), bb)
		if err != nil {
			return fmt.Errorf("failed to save mailboxes: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func mailboxKey(user string) []byte {
	return []byte("mailbox_" + user)
}
