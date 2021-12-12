package imap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/event"
	"github.com/regnull/ubikom/imap/db"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
)

var ErrAccessDenied = errors.New("access denied")

type Backend struct {
	dumpClient   pb.DMSDumpServiceClient
	lookupClient pb.LookupServiceClient
	privateKey   *easyecc.PrivateKey
	user         string
	password     string
	db           *db.Badger
	eventSender  *event.Sender
}

func NewBackend(dumpClient pb.DMSDumpServiceClient, lookupClient pb.LookupServiceClient,
	privateKey *easyecc.PrivateKey, user, password string, db *db.Badger, eventSender *event.Sender) *Backend {
	log.Debug().Msg("IMAP backend created")
	return &Backend{
		dumpClient:   dumpClient,
		lookupClient: lookupClient,
		privateKey:   privateKey,
		user:         user,
		password:     password,
		db:           db,
		eventSender:  eventSender}
}

func (b *Backend) Login(_ *imap.ConnInfo, user, pass string) (backend.User, error) {
	log.Debug().Str("user", user).Msg("[IMAP] <- LOGIN")
	privateKey := b.privateKey
	if b.privateKey != nil {
		if user != b.user || pass != b.password {
			log.Debug().Bool("authorized", false).Msg("[IMAP] -> LOGIN")
			return nil, ErrAccessDenied
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var err error
		privateKey, err = util.GetKeyFromNamePassword(ctx, user, pass, b.lookupClient)
		if err != nil {
			log.Error().Err(err).Msg("failed to get private key")
			log.Debug().Bool("authorized", false).Msg("[IMAP] -> LOGIN")
			return nil, fmt.Errorf("failed to get private key")
		}
		if b.eventSender != nil {
			b.eventSender.IMAPLogin(context.TODO(), privateKey.PublicKey().Address())
		}
	}
	log.Debug().Bool("authorized", true).Msg("[IMAP] -> LOGIN")
	u := NewUser(privateKey.PublicKey().Address(), b.db, privateKey, b.lookupClient, b.dumpClient)

	// Force polling for new messages (otherwise we will have to wait until client decides to do it).
	inbox, err := u.GetMailbox("INBOX")
	if err != nil {
		log.Error().Err(err).Msg("failed to get inbox")
		return nil, err
	}
	err = inbox.(backend.MailboxPoller).Poll()
	if err != nil {
		log.Error().Err(err).Msg("failed to poll for new messages")
		return nil, err
	}

	return u, nil
}
