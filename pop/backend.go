package pop

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap/db"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/store"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
)

// https://datatracker.ietf.org/doc/html/rfc1939

/*
Example session

13:17:41 DBG [POP] <- LOGIN user=ubikom-user
13:17:41 DBG [POP] -> LOGIN authorized=true
13:17:41 DBG [POP] <- LOCK user=ubikom-user
13:17:41 DBG [POP] -> LOCK, ignored (not implemented)
13:17:41 DBG [POP] <- STAT user=ubikom-user
13:17:41 DBG [POP] -> STAT count=0 octets=0
13:17:41 DBG [POP] <- UPDATE user=ubikom-user
13:17:41 DBG [POP] -> UPDATE deleted=0
13:17:41 DBG [POP] <- UNLOCK user=ubikom-user
13:17:41 DBG [POP] -> UNLOCK, ignored (not implemented)
*/

type Session struct {
	PrivateKey *easyecc.PrivateKey
	Messages   []*pb.ImapMessage
}

// Backend is a fake backend interface implementation used for test
type Backend struct {
	dumpClient   pb.DMSDumpServiceClient
	lookupClient pb.LookupServiceClient
	// If private key is nil, we expect to get key from the user.
	privateKey *easyecc.PrivateKey
	lock       sync.Mutex
	user       string
	password   string
	sessions   map[string]*Session
	localStore store.Store
	imapDB     *db.Badger
}

func NewBackend(dumpClient pb.DMSDumpServiceClient, lookupClient pb.LookupServiceClient,
	privateKey *easyecc.PrivateKey, user, password string, localStore store.Store,
	imapDB *db.Badger) *Backend {
	return &Backend{
		dumpClient:   dumpClient,
		lookupClient: lookupClient,
		privateKey:   privateKey,
		user:         user,
		password:     password,
		sessions:     make(map[string]*Session),
		localStore:   localStore,
		imapDB:       imapDB}
}

func (b *Backend) Authorize(user, pass string) bool {
	log.Debug().Str("user", user).Msg("[POP] <- LOGIN")

	ok := false
	if b.privateKey != nil {
		ok = user == b.user && pass == b.password
		b.lock.Lock()
		b.sessions[user] = &Session{PrivateKey: b.privateKey}
		b.lock.Unlock()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		privateKey, err := util.GetKeyFromNamePassword(ctx, user, pass, b.lookupClient)
		if err != nil {
			log.Error().Err(err).Msg("failed to get private key")
			log.Debug().Bool("authorized", false).Msg("[POP] -> LOGIN")
			return false
		}
		log.Debug().Msg("confirmed key with lookup service")

		b.lock.Lock()
		b.sessions[user] = &Session{
			PrivateKey: privateKey}
		b.lock.Unlock()
		ok = true
	}

	log.Debug().Bool("authorized", ok).Msg("[POP] -> LOGIN")
	return ok
}

func (b *Backend) Poll(ctx context.Context, user string) error {
	// Get private key for this user.
	var privateKey *easyecc.PrivateKey
	sess := b.getSession(user)

	if sess == nil {
		log.Error().Str("user", user).Msg("invalid session")
		return fmt.Errorf("invalid session")
	}
	privateKey = sess.PrivateKey

	count := 0
	// Move all local messages to IMAP inbox.
	if b.localStore != nil && b.imapDB != nil {
		localMessages, err := b.localStore.GetAll(privateKey.PublicKey().SerializeCompressed())
		if err != nil {
			return fmt.Errorf("failed to read local messages: %w", err)
		}
		log.Debug().Int("count", len(localMessages)).Msg("loaded messages from legacy store")

		for _, msg := range localMessages {
			content, err := b.decryptMessage(context.TODO(), sess.PrivateKey, msg)
			if err != nil {
				log.Error().Err(err).Str("user", user).Msg("failed to decrypt message")
				continue
			}
			msgid, err := b.imapDB.IncrementMessageID(user, "INBOX", privateKey)
			if err != nil {
				return fmt.Errorf("failed to get message ID: %w", err)
			}
			log.Debug().Uint32("msgid", msgid).Msg("moving message to IMAP mailbox")
			imapMessage := &pb.ImapMessage{
				Content:           []byte(content),
				Flag:              nil,
				ReceivedTimestamp: uint64(util.NowMs()),
				Size:              uint64(len(content)),
				Uid:               msgid,
			}
			err = b.imapDB.SaveMessage(user, db.INBOX_UID, imapMessage, privateKey)
			if err != nil {
				return fmt.Errorf("failed to move message to IMAP inbox")
			}
			err = b.localStore.Remove(msg, privateKey.PublicKey().SerializeCompressed())
			if err != nil {
				return fmt.Errorf("failed to move message to IMAP inbox")
			}
		}
	}

	// Read all IMAP messages.
	if b.imapDB != nil {
		messages, err := b.imapDB.GetMessages(user, db.INBOX_UID, privateKey)
		if err != nil {
			return fmt.Errorf("failed to read messages from IMAP DB: %w", err)
		}
		sess.Messages = append(sess.Messages, messages...)
	}
	log.Debug().Int("count", count).Msg("got local messages")

	// Read all remote messages.
	count = 0
	for {
		res, err := b.dumpClient.Receive(ctx, &pb.ReceiveRequest{
			IdentityProof: protoutil.IdentityProof(privateKey)})
		if util.ErrEqualCode(err, codes.NotFound) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}
		msg := res.GetMessage()

		msgid, err := b.imapDB.IncrementMessageID(user, "INBOX", privateKey)
		if err != nil {
			return fmt.Errorf("failed to get message ID: %w", err)
		}

		imapMsg := &pb.ImapMessage{
			Content:           msg.GetContent(),
			Flag:              []string{imap.RecentFlag},
			ReceivedTimestamp: uint64(util.NowMs()),
			Size:              uint64(len(msg.Content)),
			Uid:               msgid,
		}

		if b.imapDB != nil {
			err = b.imapDB.SaveMessage(user, db.INBOX_UID, imapMsg, privateKey)
			if err != nil {
				return fmt.Errorf("failed to save message to IMAP DB")
			}
		}
		sess.Messages = append(sess.Messages, imapMsg)
		count++
	}
	log.Debug().Int("count", count).Msg("new messages")
	return nil
}

func isDeleted(msg *pb.ImapMessage) bool {
	for _, flag := range msg.GetFlag() {
		if flag == imap.DeletedFlag {
			return true
		}
	}
	return false
}

func clearDeleted(msg *pb.ImapMessage) {
	var flags []string
	for _, flag := range msg.GetFlag() {
		if flag == imap.DeletedFlag {
			continue
		}
		flags = append(flags, flag)
	}
	msg.Flag = flags
}

func clearRecent(msg *pb.ImapMessage) bool {
	var flags []string
	found := false
	for _, flag := range msg.GetFlag() {
		if flag == imap.RecentFlag {
			found = true
			continue
		}
		flags = append(flags, flag)
	}
	msg.Flag = flags
	return found
}

// Returns total message count and total mailbox size in bytes (octets).
// Deleted messages are ignored.
func (b *Backend) Stat(user string) (messages, octets int, err error) {
	log.Debug().Str("user", user).Msg("[POP] <- STAT")
	sess := b.getSession(user)
	if sess == nil {
		return 0, 0, fmt.Errorf("invalid session")
	}
	totalSize := 0
	count := 0
	for _, msg := range sess.Messages {
		if isDeleted(msg) {
			continue
		}
		count++
		totalSize += int(msg.GetSize())
	}

	log.Debug().Int("count", count).Int("octets", totalSize).Msg("[POP] -> STAT")
	return count, totalSize, nil
}

// List of sizes of all messages in bytes (octets)
func (b *Backend) List(user string) (octets []int, err error) {
	log.Debug().Str("user", user).Msg("[POP] <- LIST")
	sess := b.getSession(user)
	if sess == nil {
		return nil, fmt.Errorf("invalid session")
	}
	var sizes []int
	for _, msg := range sess.Messages {
		if isDeleted(msg) {
			continue
		}
		sizes = append(sizes, int(msg.GetSize()))
	}

	log.Debug().Ints("sizes", sizes).Msg("[POP] -> LIST")
	return sizes, nil
}

// Returns whether message exists and if yes, then return size of the message in bytes (octets)
func (b *Backend) ListMessage(user string, msgId int) (exists bool, octets int, err error) {
	log.Debug().Str("user", user).Int("msg-id", msgId).Msg("[POP] <- LIST-MESSAGE")

	sess := b.getSession(user)
	if sess == nil {
		return false, 0, fmt.Errorf("invalid session")
	}

	msg, err := getMessageByID(sess, msgId)
	if err != nil {
		log.Debug().Err(err).Msg("[POP] -> LIST-MESSAGE failed to locate message")
		return false, 0, nil
	}

	size := len(msg)

	log.Debug().Int("size", size).Msg("[POP] -> LIST-MESSAGE")
	return true, size, nil
}

// Retrieve whole message by ID - note that message ID is a message position returned
// by List() function, so be sure to keep that order unchanged while client is connected
// See Lock() function for more details
func (b *Backend) Retr(user string, msgId int) (message string, err error) {
	log.Debug().Str("user", user).Int("msg-id", msgId).Msg("[POP] <- RETR")

	sess := b.getSession(user)
	if sess == nil {
		return "", fmt.Errorf("invalid session")
	}

	msg, err := getMessageByID(sess, msgId)
	if err != nil {
		log.Debug().Err(err).Msg("[POP] -> RETR failed to locate message")
		return "", fmt.Errorf("no such message")
	}

	if clearRecent(sess.Messages[msgId-1]) {
		if b.imapDB != nil {
			err := b.imapDB.SaveMessage(user, db.INBOX_UID, sess.Messages[msgId-1], sess.PrivateKey)
			if err != nil {
				return "", fmt.Errorf("failed to save message, %w", err)
			}
		}
	}

	log.Debug().Msg("[POP] -> RETR")
	return msg, nil
}

// Delete message by message ID - message should be just marked as deleted until
// Update() is called. Be aware that after Dele() is called, functions like List() etc.
// should ignore all these messages even if Update() hasn't been called yet
func (b *Backend) Dele(user string, msgId int) error {
	log.Debug().Str("user", user).Int("msg-id", msgId).Msg("[POP] <- DELE")
	sess := b.getSession(user)
	if sess == nil {
		return fmt.Errorf("invalid session")
	}

	err := checkMessageID(sess, msgId)
	if err != nil {
		log.Debug().Err(err).Msg("[POP] -> DELE, failed to locate message")
		return fmt.Errorf("no such message")

	}
	if isDeleted(sess.Messages[msgId-1]) {
		log.Debug().Msg("[POP] -> DELE, message already marked as deleted")
		return fmt.Errorf("already deleted")
	}

	sess.Messages[msgId-1].Flag = append(sess.Messages[msgId-1].Flag, imap.DeletedFlag)
	log.Debug().Msg("[POP] -> DELE, message marked as deleted")
	return nil
}

// Undelete all messages marked as deleted in single connection
func (b *Backend) Rset(user string) error {
	log.Debug().Str("user", user).Msg("[POP] <- RSET")
	sess := b.getSession(user)
	if sess == nil {
		return fmt.Errorf("invalid session")
	}
	for _, msg := range sess.Messages {
		clearDeleted(msg)
	}
	log.Debug().Msg("[POP] <- RSET")
	return nil
}

// List of unique IDs of all message, similar to List(), but instead of size there
// is a unique ID which persists the same across all connections. Uid (unique id) is
// used to allow client to be able to keep messages on the server.
func (b *Backend) Uidl(user string) (uids []string, err error) {
	log.Debug().Str("user", user).Msg("[POP] <- UIDL")
	sess := b.getSession(user)
	if sess == nil {
		log.Error().Str("user", user).Msg("invalid session")
		return nil, fmt.Errorf("invalid session")
	}
	var ids []string
	for _, msg := range sess.Messages {
		if isDeleted(msg) {
			continue
		}
		id := fmt.Sprintf("%d", msg.Uid)
		ids = append(ids, id)
	}
	log.Debug().Strs("ids", ids).Msg("[POP] -> UIDL")
	return ids, nil
}

// Similar to ListMessage, but returns unique ID by message ID instead of size.
func (b *Backend) UidlMessage(user string, msgId int) (exists bool, uid string, err error) {
	log.Debug().Str("user", user).Int("msg-id", msgId).Msg("[POP] <- UIDL-MESSAGE")

	sess := b.getSession(user)
	if sess == nil {
		return false, "", fmt.Errorf("invalid session")
	}

	msg, err := getMessageByID(sess, msgId)
	if err != nil {
		log.Error().Err(err).Msg("[POP] -> UIDL-MESSAGE, failed to locate message")
		return false, "", nil
	}
	id, err := getMessageID(msg)
	if err != nil {
		log.Error().Err(err).Msg("error computing message id")
		return false, "", fmt.Errorf("error computing message id")
	}
	log.Debug().Str("id", id).Msg("[POP] -> UIDL-MESSAGE")
	return true, id, nil
}

// Write all changes to persistent storage, i.e. delete all messages marked as deleted.
func (b *Backend) Update(user string) error {
	log.Debug().Str("user", user).Msg("[POP] <- UPDATE")

	sess := b.getSession(user)
	if sess == nil {
		return fmt.Errorf("invalid session")
	}

	var newMessages []*pb.ImapMessage
	count := 0
	for _, msg := range sess.Messages {
		if isDeleted(msg) {
			err := b.imapDB.DeleteMessage(user, db.INBOX_UID, msg.GetUid())
			if err != nil {
				return fmt.Errorf("failed to delete message, %w", err)
			}
			count++
			continue
		}
		newMessages = append(newMessages, msg)
	}
	sess.Messages = newMessages
	log.Debug().Int("deleted", count).Msg("[POP] -> UPDATE")
	return nil
}

func (b *Backend) Top(user string, msgId int, n int) (lines []string, err error) {
	log.Debug().Str("user", user).Int("msgId", msgId).Int("n", n).Msg("[POP] <- TOP")
	sess := b.getSession(user)
	if sess == nil {
		return nil, fmt.Errorf("invalid session")
	}

	content, err := getMessageByID(sess, msgId)

	if err != nil {
		log.Debug().Err(err).Msg("[POP] -> TOP, failed to locate message")
		return nil, fmt.Errorf("no such message")
	}
	allLines := strings.Split(content, "\n")
	bodyIndex := 0
	for i, line := range lines {
		if line == "" {
			// Empty line that separates headers from the content.
			lines = append(lines, "")
			bodyIndex = i + 1
			break
		}
		lines = append(lines, line)
	}
	for i := 0; i < n; i++ {
		j := bodyIndex + i
		if j >= len(allLines) {
			break
		}
		lines = append(lines, allLines[j])
	}
	log.Debug().Int("lines", len(lines)).Msg("[POP] -> TOP")
	return lines, nil
}

// Lock is called immediately after client is connected. The best way what to use Lock() for
// is to read all the messages into cache after client is connected. If another user
// tries to lock the storage, you should return an error to avoid data race.
func (b *Backend) Lock(user string) error {
	log.Debug().Str("user", user).Msg("[POP] <- LOCK")
	// TODO: Add timeout to the context.
	b.Poll(context.Background(), user)
	log.Debug().Msg("[POP] -> LOCK")
	return nil
}

// Release lock on storage, Unlock() is called after client is disconnected.
func (b *Backend) Unlock(user string) error {
	log.Debug().Str("user", user).Msg("[POP] <- UNLOCK")
	b.lock.Lock()
	delete(b.sessions, user)
	b.lock.Unlock()
	log.Debug().Msg("[POP] -> UNLOCK")
	return nil
}

func (b *Backend) getSession(user string) *Session {
	b.lock.Lock()
	sess, ok := b.sessions[user]
	if !ok {
		b.lock.Unlock()
		return nil
	}
	b.lock.Unlock()
	return sess
}

func (b *Backend) decryptMessage(ctx context.Context, privateKey *easyecc.PrivateKey, msg *pb.DMSMessage) (string, error) {
	lookupRes, err := b.lookupClient.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
	if err != nil {
		return "", fmt.Errorf("failed to get sender public key: %w", err)
	}
	senderKey, err := easyecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
	if err != nil {
		return "", fmt.Errorf("invalid sender public key: %w", err)
	}

	if !protoutil.VerifySignature(msg.GetSignature(), lookupRes.GetKey(), msg.GetContent()) {
		return "", fmt.Errorf("signature verification failed")
	}

	content, err := privateKey.Decrypt(msg.GetContent(), senderKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message")
	}
	return string(content), nil
}

func getMessageID(msg string) (string, error) {
	fullID := fmt.Sprintf("%x", sha256.Sum256([]byte(msg)))
	partialID := fullID[:12]
	return partialID, nil
}

func checkMessageID(sess *Session, id int) error {
	// Message id ranges from 1 to len(sess.Messages).
	if id <= 0 || id > len(sess.Messages) {
		return fmt.Errorf("invalid message id")
	}
	if isDeleted(sess.Messages[id-1]) {
		return fmt.Errorf("message is deleted")
	}
	return nil
}

func getMessageByID(sess *Session, id int) (string, error) {
	err := checkMessageID(sess, id)
	if err != nil {
		return "", err
	}
	return string(sess.Messages[id-1].GetContent()), nil
}
