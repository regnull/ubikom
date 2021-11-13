package imap

import (
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/regnull/ubikom/imap/db"
	"github.com/regnull/ubikom/pb"
)

const (
	DELIMITER = "."
)

type Mailbox struct {
	status imap.MailboxStatus
	user   string
	db     *db.Badger
}

// NewMailbox creates a brand new mailbox.
func NewMailbox(user string, name string, db *db.Badger) (*Mailbox, error) {
	mb := &Mailbox{user: user, db: db}
	mb.status.Name = name

	uid, err := db.GetNextMailboxID(user)
	if err != nil {
		return nil, err
	}

	mb.status.UidValidity = uid
	return mb, nil
}

// NewInbox creates an inbox. Inbox is a special mailbox for incoming mail. It's
// always there and cannot be deleted.
func NewInbox(user string, db *db.Badger) (*Mailbox, error) {
	mb := &Mailbox{user: user, db: db}
	mb.status.Name = "INBOX"
	uid, err := db.GetNextMailboxID(user)
	if err != nil {
		return nil, err
	}

	mb.status.UidValidity = uid
	return mb, nil
}

// NewFromProto creates a mailbox from the database data.
func NewFromProto(protoMailbox pb.ImapMailbox, user string, db *db.Badger) *Mailbox {
	mb := &Mailbox{
		user: user,
		db:   db}
	mb.status.Name = protoMailbox.GetName()
	mb.status.UidValidity = protoMailbox.GetUid()
	return mb
}

func (m *Mailbox) IsInbox() bool {
	return m.status.Name == "INBOX"
}

func (m *Mailbox) ToProto() *pb.ImapMailbox {
	return &pb.ImapMailbox{
		Name: m.status.Name,
		Uid:  m.status.UidValidity}
}

func (m *Mailbox) User() string {
	return m.user
}

func (m *Mailbox) ID() uint32 {
	return m.status.UidValidity
}

func (m *Mailbox) Name() string {
	return m.status.Name
}

func (m *Mailbox) Info() (*imap.MailboxInfo, error) {
	return &imap.MailboxInfo{
		Attributes: nil,
		Delimiter:  DELIMITER,
		Name:       m.status.Name}, nil
}

func (m *Mailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {
	return &m.status, nil
}

func (m *Mailbox) SetSubscribed(subscribed bool) error {
	if subscribed {
		return m.db.Subscribe(m.user, m.status.Name)
	} else {
		return m.db.Unsubscribe(m.user, m.status.Name)
	}
}

func (m *Mailbox) Check() error {
	// Nothing to do.
	return nil
}

func (m *Mailbox) ListMessages(uid bool, seqset *imap.SeqSet, items []imap.FetchItem,
	ch chan<- *imap.Message) error {
	defer close(ch)
	messages, err := m.db.GetMessages(m.user, m.status.UidValidity)
	if err != nil {
		return err
	}
	for i, msg := range messages {
		m := NewMessageFromProto(msg)
		seqNum := uint32(i + 1)

		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = seqNum
		}
		if !seqset.Contains(id) {
			continue
		}

		m1, err := m.Fetch(seqNum, items)
		if err != nil {
			continue
		}

		ch <- m1
	}
	return fmt.Errorf("not implemented")
}

func (m *Mailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *Mailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	return fmt.Errorf("not implemented")
}

func (m *Mailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, operation imap.FlagsOp, flags []string) error {
	return fmt.Errorf("not implemented")
}

func (m *Mailbox) CopyMessages(uid bool, seqset *imap.SeqSet, dest string) error {
	return fmt.Errorf("not implemented")
}

func (m *Mailbox) Expunge() error {
	return fmt.Errorf("not implemented")
}

func (m *Mailbox) Poll() error {
	return fmt.Errorf("not implemented")
}
