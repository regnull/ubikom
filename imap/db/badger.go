package db

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/golang/protobuf/proto"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
)

var ErrNotFound = fmt.Errorf("not found")

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

func getMailboxes(txn *badger.Txn, user string, privateKey *easyecc.PrivateKey) (*pb.ImapMailboxes, error) {
	item, err := txn.Get(mailboxKey(user))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			mailboxes := initialMailboxes()
			bbe, err := marshalAndEncrypt(mailboxes, privateKey)
			if err != nil {
				return nil, err
			}
			err = txn.Set(mailboxKey(user), bbe)
			if err != nil {
				return nil, fmt.Errorf("failed to save mailboxes: %w", err)
			}
			return mailboxes, nil
		}
		return nil, fmt.Errorf("error getting mailbox: %w", err)
	}

	mailboxes := &pb.ImapMailboxes{}
	err = unmarhalItemAndDecrypt(item, mailboxes, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox: %w", err)
	}
	log.Debug().Interface("mailboxes", mailboxes).Msg("got mailboxes")
	return mailboxes, nil
}

func saveMailboxes(txn *badger.Txn, mailboxes *pb.ImapMailboxes, user string, privateKey *easyecc.PrivateKey) error {
	bbe, err := marshalAndEncrypt(mailboxes, privateKey)
	if err != nil {
		return err
	}
	err = txn.Set(mailboxKey(user), bbe)
	if err != nil {
		return fmt.Errorf("failed to save mailboxes: %w", err)
	}
	return nil
}

func (b *Badger) GetMailboxes(user string, privateKey *easyecc.PrivateKey) ([]*pb.ImapMailbox, error) {
	var mbs []*pb.ImapMailbox
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
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

func (b *Badger) GetMailbox(user string, name string, privateKey *easyecc.PrivateKey) (*pb.ImapMailbox, error) {
	var mailboxes *pb.ImapMailboxes
	err := b.db.Update(func(txn *badger.Txn) error {
		var err error
		mailboxes, err = getMailboxes(txn, user, privateKey)
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

func (b *Badger) CreateMailbox(user string, mb *pb.ImapMailbox, privateKey *easyecc.PrivateKey) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		for _, m := range mailboxes.GetMailbox() {
			log.Debug().Str("name", m.GetName()).Msg("got mailbox")
			if mb.GetName() == m.GetName() {
				return fmt.Errorf("mailbox already exists")
			}
		}
		mailboxes.Mailbox = append(mailboxes.Mailbox, mb)
		err = saveMailboxes(txn, mailboxes, user, privateKey)
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

func (b *Badger) DeleteMailbox(user string, name string, privateKey *easyecc.PrivateKey) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
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
		err = saveMailboxes(txn, mailboxes, user, privateKey)
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

func (b *Badger) RenameMailbox(user string, existingName, newName string, privateKey *easyecc.PrivateKey) error {
	if strings.ToUpper(existingName) == "INBOX" {
		// TODO: Implement this (create a copy of inbox, rename it, leave inbox empty).
		return errors.New("renaming inbox is not allowed")
	}
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		for _, mb := range mailboxes.GetMailbox() {
			if strings.HasPrefix(mb.GetName(), existingName) {
				n := newName + mb.GetName()[len(existingName):]
				mb.Name = n
			}
		}
		err = saveMailboxes(txn, mailboxes, user, privateKey)
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

func (b *Badger) Subscribe(user string, name string, privateKey *easyecc.PrivateKey) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		found := false
		for _, s := range mailboxes.GetSubscribed() {
			if s == name {
				found = true
				break
			}
		}
		if found {
			return fmt.Errorf("already subscribed")
		}
		mailboxes.Subscribed = append(mailboxes.Subscribed, name)
		err = saveMailboxes(txn, mailboxes, user, privateKey)
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

func (b *Badger) Unsubscribe(user string, name string, privateKey *easyecc.PrivateKey) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		var newSubscribed []string
		for _, s := range mailboxes.GetSubscribed() {
			if s == name {
				continue
			}
			newSubscribed = append(newSubscribed, s)
		}
		if len(newSubscribed) == len(mailboxes.GetSubscribed()) {
			return fmt.Errorf("not subscribed")
		}
		mailboxes.Subscribed = newSubscribed
		err = saveMailboxes(txn, mailboxes, user, privateKey)
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

func (b *Badger) Subscribed(user string, name string, privateKey *easyecc.PrivateKey) (bool, error) {
	subscribed := false
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
		if err != nil {
			return fmt.Errorf("failed to get mailboxes: %w", err)
		}
		for _, s := range mailboxes.GetSubscribed() {
			if s == name {
				subscribed = true
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return subscribed, nil
}

func mailboxKey(user string) []byte {
	return []byte("mailbox_" + user)
}

func (b *Badger) SaveMessage(user string, mbid uint32, msg *pb.ImapMessage, privateKey *easyecc.PrivateKey) error {
	bbe, err := marshalAndEncrypt(msg, privateKey)
	if err != nil {
		return err
	}
	err = b.db.Update(func(txn *badger.Txn) error {
		key := messageKey(user, mbid, msg.GetUid())
		log.Debug().Str("key", string(key)).Msg("saving message")
		e := badger.NewEntry(messageKey(user, mbid, msg.GetUid()), bbe)
		if b.ttl > 0 {
			e = e.WithTTL(b.ttl)
		}
		return txn.SetEntry(e)
	})
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

func (b *Badger) GetMessages(user string, mailbox uint32, privateKey *easyecc.PrivateKey) ([]*pb.ImapMessage, error) {
	var messages []*pb.ImapMessage
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := mailboxMessagePrefix(user, mailbox)
		log.Debug().Str("prefix", string(prefix)).Msg("reading messages")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			msg := &pb.ImapMessage{}
			err := unmarhalItemAndDecrypt(item, msg, privateKey)
			if err != nil {
				return err
			}
			messages = append(messages, msg)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (b *Badger) DeleteMessage(user string, mbid uint32, msgid uint32) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(messageKey(user, mbid, msgid))
	})
}

func (b *Badger) mutateMailboxes(user string, f func(mailboxes *pb.ImapMailboxes), privateKey *easyecc.PrivateKey) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes := initialMailboxes()
		item, err := txn.Get(mailboxKey(user))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == nil {
			err = unmarhalItemAndDecrypt(item, mailboxes, privateKey)
			if err != nil {
				return err
			}
		}
		f(mailboxes)
		err = saveMailboxes(txn, mailboxes, user, privateKey)
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

func (b *Badger) IncrementMailboxID(user string, privateKey *easyecc.PrivateKey) (uint32, error) {
	var mbid uint32
	err := b.mutateMailboxes(user, func(mailboxes *pb.ImapMailboxes) {
		mbid = mailboxes.GetNextMailboxUid()
		mailboxes.NextMailboxUid++
	}, privateKey)
	if err != nil {
		return 0, err
	}
	return mbid, nil
}

func (b *Badger) GetNextMailboxID(user string, privateKey *easyecc.PrivateKey) (uint32, error) {
	var mbid uint32
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
		if err != nil {
			return err
		}
		mbid = mailboxes.GetNextMailboxUid()
		return nil
	})
	if err != nil {
		return 0, err
	}
	return mbid, nil
}

func (b *Badger) IncrementMessageID(user string, mb string, privateKey *easyecc.PrivateKey) (uint32, error) {
	var msgid uint32
	err := b.mutateMailboxes(user, func(mailboxes *pb.ImapMailboxes) {
		for _, m := range mailboxes.GetMailbox() {
			if m.GetName() != mb {
				continue
			}
			msgid = m.GetNextMessageUid()
			m.NextMessageUid++
			break
		}
	}, privateKey)
	if err != nil {
		return 0, err
	}
	return msgid, nil
}

func (b *Badger) GetNextMessageID(user string, mb string, privateKey *easyecc.PrivateKey) (uint32, error) {
	var msgid uint32
	err := b.db.Update(func(txn *badger.Txn) error {
		mailboxes, err := getMailboxes(txn, user, privateKey)
		if err != nil {
			return err
		}
		for _, m := range mailboxes.GetMailbox() {
			if m.GetName() != mb {
				continue
			}
			msgid = m.GetNextMessageUid()
			break
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return msgid, nil
}

func messageKey(user string, mbid uint32, msgid uint32) []byte {
	return []byte("message_" + user + "_" + strconv.FormatInt(int64(mbid), 10) +
		"_" + strconv.FormatInt(int64(msgid), 10))
}

func mailboxMessagePrefix(user string, mbid uint32) []byte {
	return []byte("message_" + user + "_" + strconv.FormatInt(int64(mbid), 10) + "_")
}

func marshalAndEncrypt(msg proto.Message, privateKey *easyecc.PrivateKey) ([]byte, error) {
	bb, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error marshaling message: %w", err)
	}
	bbe, err := privateKey.EncryptSymmetric(bb)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt mailboxes: %w", err)
	}
	return bbe, nil
}

func unmarshalAndDecrypt(data []byte, msg proto.Message, privateKey *easyecc.PrivateKey) error {
	bb, err := privateKey.DecryptSymmetric(data)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}
	err = proto.Unmarshal(bb, msg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return nil
}

func unmarhalItemAndDecrypt(item *badger.Item, msg proto.Message, privateKey *easyecc.PrivateKey) error {
	return item.Value(func(val []byte) error {
		return unmarshalAndDecrypt(val, msg, privateKey)
	})
}

func initialMailboxes() *pb.ImapMailboxes {
	var initialMailboxes = &pb.ImapMailboxes{
		Mailbox: []*pb.ImapMailbox{
			{
				Name:           "INBOX",
				Attribute:      nil,
				Uid:            INBOX_UID,
				NextMessageUid: FIRST_MESSAGE_UID,
			},
			{
				Name:           "Sent",
				Attribute:      nil,
				Uid:            SENT_UID,
				NextMessageUid: FIRST_MESSAGE_UID,
			},
			{
				Name:           "Trash",
				Attribute:      nil,
				Uid:            TRASH_UID,
				NextMessageUid: FIRST_MESSAGE_UID,
			},
		},
		NextMailboxUid: FIRST_REGULAR_MAILBOX_UID,
	}
	return initialMailboxes
}
