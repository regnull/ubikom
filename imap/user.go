package imap

import (
	"errors"
	"strings"

	"github.com/emersion/go-imap/backend"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap/db"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
)

type User struct {
	name         string
	db           *db.Badger
	privateKey   *easyecc.PrivateKey
	lookupClient pb.LookupServiceClient
	dumpClient   pb.DMSDumpServiceClient
	updateChan   chan<- backend.Update
}

func NewUser(name string, db *db.Badger, privateKey *easyecc.PrivateKey,
	lookupClient pb.LookupServiceClient, dumpClient pb.DMSDumpServiceClient,
	updateChan chan<- backend.Update) *User {
	log.Debug().Str("name", name).Msg("[IMAP] creating new user")
	return &User{
		name:         name,
		db:           db,
		privateKey:   privateKey,
		lookupClient: lookupClient,
		dumpClient:   dumpClient,
		updateChan:   updateChan,
	}
}

func (u *User) Username() string {
	return u.name
}

func (u *User) ListMailboxes(subscribed bool) ([]backend.Mailbox, error) {
	u.logEnter("ListMailboxes")
	defer u.logExit("ListMailboxes")

	mailboxes, err := u.db.GetMailboxes(u.name, u.privateKey)
	if err != nil {
		log.Error().Str("user", u.name).Err(err).Msg("ListMailboxes failed")
		return nil, err
	}
	var ret []backend.Mailbox
	for _, mb := range mailboxes {
		log.Debug().Str("user", u.name).Str("mailbox", mb.GetName()).Msg("got mailbox")
		m := NewMailboxFromProto(mb, u.name, u.db, u.lookupClient, u.dumpClient, u.privateKey)
		ret = append(ret, m)
	}

	if subscribed {
		var filteredMailboxes []backend.Mailbox
		for _, mb := range ret {
			sub, err := u.db.Subscribed(u.name, mb.Name(), u.privateKey)
			if err != nil {
				return nil, err
			}
			if !sub {
				continue
			}
			filteredMailboxes = append(filteredMailboxes, mb)
		}
		ret = filteredMailboxes
	}

	return ret, nil
}

func (u *User) GetMailbox(name string) (backend.Mailbox, error) {
	u.logEnter("GetMailbox")
	defer u.logExit("GetMailbox")

	mailboxes, err := u.db.GetMailboxes(u.name, u.privateKey)
	if err != nil {
		return nil, err
	}

	for _, mb := range mailboxes {
		if mb.Name == name {
			mb := NewMailboxFromProto(mb, u.name, u.db, u.lookupClient, u.dumpClient, u.privateKey)
			return mb, nil
		}
	}

	return nil, errors.New("no such mailbox")
}

func (u *User) CreateMailbox(name string) error {
	u.logEnter("CreateMailbox")
	defer u.logExit("CreateMailbox")

	mb, err := NewMailbox(u.name, name, u.db, u.lookupClient, u.dumpClient, u.privateKey, u.updateChan)
	if err != nil {
		return err
	}
	return u.db.CreateMailbox(u.name, mb.ToProto(), u.privateKey)
}

func (u *User) DeleteMailbox(name string) error {
	u.logEnter("DeleteMailbox")
	defer u.logExit("DeleteMailbox")

	if strings.ToUpper(name) == "INBOX" {
		return errors.New("can't delete inbox")
	}

	return u.db.DeleteMailbox(u.name, name, u.privateKey)
}

func (u *User) RenameMailbox(existingName, newName string) error {
	u.logEnter("RenameMailbox")
	defer u.logExit("RenameMailbox")

	return u.db.RenameMailbox(u.name, existingName, newName, u.privateKey)
}

func (u *User) Logout() error {
	u.logEnter("Logout")
	defer u.logExit("Logout")

	// Nothing to do.
	return nil
}

func (u *User) logEnter(name string) {
	log.Debug().Str("user", u.name).Msg("[IMAP] <- " + name)
}

func (u *User) logExit(name string) {
	log.Debug().Str("user", u.name).Msg("[IMAP] -> " + name)
}
