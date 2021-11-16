package imap

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/backendutil"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap/db"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
)

const (
	DELIMITER = "."
)

type Mailbox struct {
	status       imap.MailboxStatus
	user         string
	db           *db.Badger
	lookupClient pb.LookupServiceClient
	dumpClient   pb.DMSDumpServiceClient
	privateKey   *easyecc.PrivateKey
}

// NewMailbox creates a brand new mailbox.
func NewMailbox(user string, name string, db *db.Badger, lookupClient pb.LookupServiceClient,
	dumpClient pb.DMSDumpServiceClient, privateKey *easyecc.PrivateKey) (*Mailbox, error) {
	mb := &Mailbox{user: user, db: db, lookupClient: lookupClient,
		dumpClient: dumpClient, privateKey: privateKey}
	mb.status.Name = name

	uid, err := db.GetNextMailboxID(user, privateKey)
	if err != nil {
		return nil, err
	}

	mb.status.UidValidity = uid
	return mb, nil
}

// NewMailboxFromProto creates a mailbox from the database data.
func NewMailboxFromProto(protoMailbox *pb.ImapMailbox, user string, db *db.Badger,
	lookupClient pb.LookupServiceClient, dumpClient pb.DMSDumpServiceClient,
	privateKey *easyecc.PrivateKey) *Mailbox {
	mb := &Mailbox{
		user:         user,
		db:           db,
		lookupClient: lookupClient,
		dumpClient:   dumpClient,
		privateKey:   privateKey}
	mb.status.Name = protoMailbox.GetName()
	mb.status.UidValidity = protoMailbox.GetUid()
	log.Debug().Interface("mb", mb).Msg("created mailbox from proto")
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
	m.logEnter("Info")
	defer m.logExit("Info")
	return &imap.MailboxInfo{
		Attributes: nil,
		Delimiter:  DELIMITER,
		Name:       m.status.Name}, nil
}

func (m *Mailbox) flags() ([]string, error) {
	flagsMap := make(map[string]bool)
	messages, err := m.db.GetMessages(m.user, m.status.UidValidity, m.privateKey)
	if err != nil {
		m.logError(err).Msg("failed to get messages")
		return nil, err
	}
	m.logDebug().Int("count", len(messages)).Msg("got messages")
	for _, msg := range messages {
		for _, f := range msg.Flag {
			if !flagsMap[f] {
				flagsMap[f] = true
			}
		}
	}

	var flags []string
	for f := range flagsMap {
		flags = append(flags, f)
	}
	m.logDebug().Interface("flags", flags).Msg("got flags")
	return flags, nil
}

func (m *Mailbox) unseenSeqNum() (uint32, uint32, error) {
	messages, err := m.db.GetMessages(m.user, m.status.UidValidity, m.privateKey)
	if err != nil {
		m.logError(err).Msg("failed to get messages")
		return 0, 0, err
	}
	for i, msg := range messages {
		seqNum := uint32(i + 1)

		seen := false
		for _, flag := range msg.Flag {
			if flag == imap.SeenFlag {
				seen = true
				break
			}
		}

		if !seen {
			log.Debug().Int("count", len(messages)).Uint32("seqNum", seqNum).Msg("unseenSeqNum result")
			return uint32(len(messages)), seqNum, nil
		}
	}
	return uint32(len(messages)), 0, nil
}

func (m *Mailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {
	m.logEnter("Status")
	defer m.logExit("Status")
	status := imap.NewMailboxStatus(m.status.Name, items)
	flags, err := m.flags()
	if err != nil {
		m.logError(err).Msg("failed to get flags")
		return nil, err
	}
	status.Flags = flags
	status.PermanentFlags = []string{"\\*"}
	total, unseenSeqNum, err := m.unseenSeqNum()
	if err != nil {
		m.logError(err).Msg("failed to get unseenSeqNum")
		return nil, err
	}
	status.UnseenSeqNum = unseenSeqNum

	msgid, err := m.db.GetNextMessageID(m.user, m.privateKey)
	if err != nil {
		m.logError(err).Msg("failed to get next message id")
		return nil, err
	}

	for _, name := range items {
		switch name {
		case imap.StatusMessages:
			status.Messages = total
		case imap.StatusUidNext:
			status.UidNext = msgid
		case imap.StatusUidValidity:
			status.UidValidity = 1
		case imap.StatusRecent:
			status.Recent = 0 // TODO
		case imap.StatusUnseen:
			status.Unseen = 0 // TODO
		}
	}

	m.logDebug().Interface("status", status).Msg("got status")
	return status, nil
}

func (m *Mailbox) SetSubscribed(subscribed bool) error {
	m.logEnter("SetSubscribed")
	defer m.logExit("SetSubscribed")
	if subscribed {
		return m.db.Subscribe(m.user, m.status.Name, m.privateKey)
	} else {
		return m.db.Unsubscribe(m.user, m.status.Name, m.privateKey)
	}
}

func (m *Mailbox) Check() error {
	m.logEnter("Check")
	defer m.logExit("Check")
	// Nothing to do.
	return nil
}

func (m *Mailbox) ListMessages(uid bool, seqset *imap.SeqSet, items []imap.FetchItem,
	ch chan<- *imap.Message) error {
	m.logEnter("ListMessages")
	defer m.logExit("ListMessages")
	defer close(ch)
	messages, err := m.db.GetMessages(m.user, m.status.UidValidity, m.privateKey)
	if err != nil {
		m.logError(err)
		log.Error().Str("user", m.user).Str("mailbox", m.status.Name).Err(err).Msg("failed to read messages from the database")
		return err
	}
	for i, msg := range messages {
		log.Debug().Uint32("id", msg.Uid).Msg("got message")
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
	return nil
}

func (m *Mailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {
	m.logEnter("SearchMessages")
	defer m.logExit("SearchMessages")
	return nil, fmt.Errorf("not implemented")
}

func (m *Mailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	m.logEnter("CreateMessage")
	defer m.logExit("CreateMessage")
	if date.IsZero() {
		date = time.Now()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		m.logError(err).Msg("failed to read message body")
		return err
	}

	msgid, err := m.db.GetNextMessageID(m.user, m.privateKey)
	if err != nil {
		m.logError(err).Msg("failed to get next message id")
		return err
	}
	message := &Message{
		Uid:   msgid,
		Date:  date,
		Size:  uint32(len(b)),
		Flags: flags,
		Body:  b,
	}

	err = m.db.SaveMessage(m.user, m.status.UidValidity, message.ToProto(), m.privateKey)
	if err != nil {
		m.logError(err).Msg("failed to save message")
	}
	return nil
}

func (m *Mailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, op imap.FlagsOp, flags []string) error {
	m.logEnter("UpdateMessagesFlags")
	defer m.logExit("UpdateMessagesFlags")

	messages, err := m.db.GetMessages(m.user, m.status.UidValidity, m.privateKey)
	if err != nil {
		m.logError(err).Msg("failed to get messages")
		return err
	}
	for i, msg := range messages {
		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = uint32(i + 1)
		}
		if !seqset.Contains(id) {
			continue
		}

		msg.Flag = backendutil.UpdateFlags(msg.Flag, op, flags)
	}

	return nil
}

func (m *Mailbox) CopyMessages(uid bool, seqset *imap.SeqSet, dest string) error {
	m.logEnter("CopyMessages")
	defer m.logExit("CopyMessages")
	return fmt.Errorf("not implemented")
}

func (m *Mailbox) Expunge() error {
	m.logEnter("Expunge")
	defer m.logExit("Expunge")

	messages, err := m.db.GetMessages(m.user, m.status.UidValidity, m.privateKey)
	if err != nil {
		m.logError(err).Msg("failed to get messages")
		return err
	}
	for _, msg := range messages {
		deleted := false
		for _, flag := range msg.Flag {
			if flag == imap.DeletedFlag {
				deleted = true
				break
			}
		}

		if deleted {
			err := m.db.DeleteMessage(m.user, m.status.UidValidity, msg.Uid)
			if err != nil {
				m.logError(err).Msg("failed to delete message")
				return err
			}
		}
	}

	return nil
}

func (m *Mailbox) Poll() error {
	m.logEnter("Poll")
	defer m.logExit("Poll")

	if m.IsInbox() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := m.getMessageFromDumpServer(ctx)
		if err != nil {
			m.logError(err).Msg("failed to get messages from dump server")
			return err
		}
	}
	return nil
}

func (m *Mailbox) getMessageFromDumpServer(ctx context.Context) error {
	log.Debug().Str("user", m.user).Str("mailbox", m.status.Name).Msg("getting messages from dump server")
	// Read all remote messages.
	count := 0
	for {
		res, err := m.dumpClient.Receive(ctx, &pb.ReceiveRequest{
			IdentityProof: protoutil.IdentityProof(m.privateKey)})
		if util.ErrEqualCode(err, codes.NotFound) {
			if count == 0 {
				log.Debug().Msg("no new messages")
			} else {
				log.Debug().Int("count", count).Msg("got new messages")
			}
			break
		}
		if err != nil {
			m.logError(err).Msg("failed to receive message")
			return fmt.Errorf("failed to receive message: %w", err)
		}
		count++
		log.Debug().Str("user", m.user).Str("mailbox", m.status.Name).Msg("got new message")
		msg := res.GetMessage()
		msgid, err := m.db.GetNextMessageID(m.user, m.privateKey)
		if err != nil {
			m.logError(err).Msg("failed go get next message id")
			return err
		}

		content, err := m.decryptMessage(ctx, m.privateKey, msg)
		if err != nil {
			m.logError(err).Msg("failed to decrypt message")
			return err
		}

		message := &Message{
			Uid:   msgid,
			Date:  time.Now(),
			Size:  uint32(len(content)),
			Flags: []string{"\\Unseen"},
			Body:  content,
		}
		err = m.db.SaveMessage(m.user, 0, message.ToProto(), m.privateKey)
		if err != nil {
			m.logError(err).Msg("failed to save message")
			return err
		}
	}
	log.Debug().Int("count", count).Msg("total messages")
	return nil
}

func (m *Mailbox) decryptMessage(ctx context.Context, privateKey *easyecc.PrivateKey, msg *pb.DMSMessage) ([]byte, error) {
	lookupRes, err := m.lookupClient.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
	if err != nil {
		return nil, fmt.Errorf("failed to get sender public key: %w", err)
	}
	senderKey, err := easyecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
	if err != nil {
		return nil, fmt.Errorf("invalid sender public key: %w", err)
	}

	if !protoutil.VerifySignature(msg.GetSignature(), lookupRes.GetKey(), msg.GetContent()) {
		return nil, fmt.Errorf("signature verification failed")
	}

	content, err := privateKey.Decrypt(msg.GetContent(), senderKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message")
	}
	return content, nil
}

func (m *Mailbox) logEnter(name string) {
	log.Debug().Str("user", m.user).Str("mailbox", m.status.Name).Msg("[IMAP] <- " + name)
}

func (m *Mailbox) logExit(name string) {
	log.Debug().Str("user", m.user).Str("mailbox", m.status.Name).Msg("[IMAP] -> " + name)
}

func (m *Mailbox) logDebug() *zerolog.Event {
	return log.Debug().Str("user", m.user).Str("mailbox", m.status.Name)
}

func (m *Mailbox) logError(err error) *zerolog.Event {
	return log.Error().Err(err).Str("user", m.user).Str("mailbox", m.status.Name)
}
