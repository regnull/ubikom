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

// NewFromProto creates a mailbox from the database data.
func NewFromProto(protoMailbox pb.ImapMailbox, user string, db *db.Badger) *Mailbox {
	mb := &Mailbox{
		user: user,
		db:   db}
	mb.status.Name = protoMailbox.GetName()
	mb.status.UidValidity = protoMailbox.GetUid()
	return mb
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
	return nil, fmt.Errorf("not implemented")
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
