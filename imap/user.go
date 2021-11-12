package imap

import (
	"github.com/emersion/go-imap/backend"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap/db"
)

type User struct {
	name       string
	db         *db.Badger
	privateKey *easyecc.PrivateKey
}

func NewUser(name string, db *db.Badger, privateKey *easyecc.PrivateKey) *User {
	return &User{name: name, db: db, privateKey: privateKey}
}

func (u *User) Username() string {
	return u.name
}

func (u *User) ListMailboxes(subscribed bool) ([]backend.Mailbox, error) {
	mailboxes, err := u.db.GetMailboxes(u.name)
	if err != nil {
		return nil, err
	}
	var ret []backend.Mailbox
	for _, mb := range mailboxes {
		m, err := NewMailbox(u.name, mb.Name, u.db)
		if err != nil {
			return nil, err
		}
		ret = append(ret, m)
	}
	return ret, nil
}

func (u *User) GetMailbox(name string) (backend.Mailbox, error) {
	mailboxes, err := u.db.GetMailboxes(u.name)
	if err != nil {
		return nil, err
	}

	for _, mb := range mailboxes {
		if mb.Name == name {
			mb, err := NewMailbox(u.name, name, u.db)
			if err != nil {
				return nil, err
			}
			return mb, nil
		}
	}

	return nil, nil
}

func (u *User) CreateMailbox(name string) error {
	mb, err := NewMailbox(u.name, name, u.db)
	if err != nil {
		return err
	}
	return u.db.CreateMailbox(u.name, mb.ToProto())
}

func (u *User) DeleteMailbox(name string) error {
	return u.db.DeleteMailbox(u.name, name)
}

func (u *User) RenameMailbox(existingName, newName string) error {
	return u.db.RenameMailbox(u.name, existingName, newName)
}

func (u *User) Logout() error {
	// Nothing to do.
	return nil
}
