package pop

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"

	"github.com/btcsuite/btcutil/base58"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/store"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
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
	Messages   []*pb.DMSMessage
	Deleted    []bool
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
}

func NewBackend(dumpClient pb.DMSDumpServiceClient, lookupClient pb.LookupServiceClient,
	privateKey *easyecc.PrivateKey, user, password string, localStore store.Store) *Backend {
	return &Backend{
		dumpClient:   dumpClient,
		lookupClient: lookupClient,
		privateKey:   privateKey,
		user:         user,
		password:     password,
		sessions:     make(map[string]*Session),
		localStore:   localStore}
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
		salt := base58.Decode(user)
		privateKey := easyecc.NewPrivateKeyFromPassword([]byte(pass), salt)

		// Confirm that this key is registered.
		res, err := b.lookupClient.LookupKey(context.TODO(), &pb.LookupKeyRequest{
			Key: privateKey.PublicKey().SerializeCompressed()})

		if err != nil {
			log.Error().Err(err).Msg("failed to look up key")
			log.Debug().Bool("authorized", false).Msg("[POP] -> LOGIN")
			return false
		}
		if res.GetDisabled() {
			log.Error().Msg("this key is disabled")
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
	// Read all locally stored messages.
	if b.localStore != nil {
		localMessages, err := b.localStore.GetAll(privateKey.PublicKey().SerializeCompressed())
		if err != nil {
			return fmt.Errorf("failed to read local messages: %w", err)
		}

		for _, msg := range localMessages {
			sess.Messages = append(sess.Messages, msg)
			sess.Deleted = append(sess.Deleted, false)
			count++
		}
	}
	log.Debug().Int("count", count).Msg("got local messages")

	// Read all remote messages.
	for {
		res, err := b.dumpClient.Receive(ctx, &pb.ReceiveRequest{
			IdentityProof: protoutil.IdentityProof(privateKey)})
		if util.ErrEqualCode(err, codes.NotFound) {
			if count == 0 {
				log.Debug().Msg("no new messages")
			} else {
				log.Debug().Int("count", count).Msg("got new messages")
			}
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}
		msg := res.GetMessage()

		if b.localStore != nil {
			err = b.localStore.Save(msg, privateKey.PublicKey().SerializeCompressed())
			if err != nil {
				log.Error().Err(err).Msg("error saving message to local store")
			}
		}
		sess.Messages = append(sess.Messages, msg)
		sess.Deleted = append(sess.Deleted, false)
		count++
	}
	log.Debug().Int("count", count).Msg("total messages")
	return nil
}

/* message-number (or message ID)

   After the POP3 server has opened the maildrop, it assigns a message-
   number to each message, and notes the size of each message in octets.
   The first message in the maildrop is assigned a message-number of
   "1", the second is assigned "2", and so on, so that the nth message
   in a maildrop is assigned a message-number of "n".  In POP3 commands
   and responses, all message-numbers and message sizes are expressed in
   base-10 (i.e., decimal).

*/

/*
STAT

Arguments: none

Restrictions:
	may only be given in the TRANSACTION state

Discussion:
	The POP3 server issues a positive response with a line
	containing information for the maildrop.  This line is
	called a "drop listing" for that maildrop.

	In order to simplify parsing, all POP3 servers are
	required to use a certain format for drop listings.  The
	positive response consists of "+OK" followed by a single
	space, the number of messages in the maildrop, a single
	space, and the size of the maildrop in octets.  This memo
	makes no requirement on what follows the maildrop size.
	Minimal implementations should just end that line of the
	response with a CRLF pair.  More advanced implementations
	may include other information.

	   NOTE: This memo STRONGLY discourages implementations
	   from supplying additional information in the drop
	   listing.  Other, optional, facilities are discussed
	   later on which permit the client to parse the messages
	   in the maildrop.

	Note that messages marked as deleted are not counted in
	either total.

Possible Responses:
	+OK nn mm

Examples:
	C: STAT
	S: +OK 2 320
*/

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
	for i, msg := range sess.Messages {
		if sess.Deleted[i] {
			continue
		}
		count++
		totalSize += easyecc.GetPlainTextLength(len(msg.GetContent()))
	}

	log.Debug().Int("count", count).Int("octets", totalSize).Msg("[POP] -> STAT")
	return count, totalSize, nil
}

/*
      LIST [msg]

         Arguments:
             a message-number (optional), which, if present, may NOT
             refer to a message marked as deleted

		 Restrictions:
             may only be given in the TRANSACTION state

         Discussion:
             If an argument was given and the POP3 server issues a
             positive response with a line containing information for
             that message.  This line is called a "scan listing" for
             that message.

             If no argument was given and the POP3 server issues a
             positive response, then the response given is multi-line.
             After the initial +OK, for each message in the maildrop,
             the POP3 server responds with a line containing
             information for that message.  This line is also called a
             "scan listing" for that message.  If there are no
             messages in the maildrop, then the POP3 server responds
             with no scan listings--it issues a positive response
             followed by a line containing a termination octet and a
             CRLF pair.

             In order to simplify parsing, all POP3 servers are
             required to use a certain format for scan listings.  A
             scan listing consists of the message-number of the
             message, followed by a single space and the exact size of
             the message in octets.  Methods for calculating the exact
             size of the message are described in the "Message Format"
             section below.  This memo makes no requirement on what
             follows the message size in the scan listing.  Minimal
             implementations should just end that line of the response
             with a CRLF pair.  More advanced implementations may
             include other information, as parsed from the message.

                NOTE: This memo STRONGLY discourages implementations
                from supplying additional information in the scan
                listing.  Other, optional, facilities are discussed
                later on which permit the client to parse the messages
                in the maildrop.

             Note that messages marked as deleted are not listed.

         Possible Responses:
             +OK scan listing follows
             -ERR no such message

         Examples:
             C: LIST
             S: +OK 2 messages (320 octets)
             S: 1 120
             S: 2 200
             S: .
               ...
             C: LIST 2
             S: +OK 2 200
               ...
             C: LIST 3
             S: -ERR no such message, only 2 messages in maildrop
*/

// List of sizes of all messages in bytes (octets)
func (b *Backend) List(user string) (octets []int, err error) {
	log.Debug().Str("user", user).Msg("[POP] <- LIST")
	sess := b.getSession(user)
	if sess == nil {
		return nil, fmt.Errorf("invalid session")
	}
	var sizes []int
	for i, msg := range sess.Messages {
		if sess.Deleted[i] {
			continue
		}
		sizes = append(sizes, easyecc.GetPlainTextLength(len(msg.GetContent())))
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

	if msgId <= 0 || msgId > len(sess.Messages) {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> LIST-MESSAGE, no such message")
		return false, 0, nil
	}
	if sess.Deleted[msgId-1] {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> LIST-MESSAGE, message is deleted")
		return false, 0, nil
	}

	msg, err := b.getMessageByID(sess, msgId)
	if err != nil {
		b.lock.Unlock()
		log.Debug().Err(err).Msg("[POP] -> LIST-MESSAGE failed to locate message")
	}

	size := easyecc.GetPlainTextLength(len(msg.GetContent()))

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

	if msgId > len(sess.Messages) {
		log.Debug().Msg("[POP] -> RETR, no such message")
		return "", fmt.Errorf("no such message")
	}
	if sess.Deleted[msgId] {
		log.Debug().Msg("[POP] -> RETR, message is deleted")
		return "", fmt.Errorf("message is deleted")
	}

	msg := sess.Messages[msgId]
	content, err := b.decryptMessage(context.TODO(), sess.PrivateKey, msg)
	if err != nil {
		log.Error().Err(err).Msg("error decrypting message")
		return "", fmt.Errorf("error decrypting message")
	}
	log.Debug().Str("message", getFirst(content, 16)).Msg("[POP] -> RETR")
	return content, nil
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

	if msgId > len(sess.Messages) {
		log.Debug().Msg("[POP] -> DELE, no such message")
		return fmt.Errorf("no such message")
	}
	sess.Deleted[msgId] = true
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
	for i := range sess.Deleted {
		sess.Deleted[i] = false
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
	for i, msg := range sess.Messages {
		if sess.Deleted[i] {
			continue
		}
		id, err := getMessageID(msg)
		if err != nil {
			log.Error().Err(err).Msg("error computing message id")
			return nil, fmt.Errorf("error computing message id")
		}
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

	// Although it's not quite clear from RFC 1939, msgId ranges from 1 to len(sess.Messages).

	if msgId <= 0 || msgId > len(sess.Messages) {
		log.Error().Msg("[POP] -> UIDL-MESSAGE, no such message")
		return false, "", nil
	}
	if sess.Deleted[msgId-1] {
		log.Error().Msg("[POP] -> UIDL-MESSAGE, message is deleted")
		return false, "", nil
	}
	id, err := getMessageID(sess.Messages[msgId-1])
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

	var newMessages []*pb.DMSMessage
	count := 0
	for i, msg := range sess.Messages {
		if sess.Deleted[i] {
			if b.localStore != nil {
				b.localStore.Remove(msg, sess.PrivateKey.PublicKey().SerializeCompressed())
			}
			count++
			continue
		}
		newMessages = append(newMessages, msg)
	}
	sess.Messages = newMessages
	sess.Deleted = make([]bool, len(newMessages))
	log.Debug().Int("deleted", count).Msg("[POP] -> UPDATE")
	return nil
}

func (b *Backend) Top(user string, msgId int, n int) (lines []string, err error) {
	log.Debug().Str("user", user).Int("msgId", msgId).Int("n", n).Msg("[POP] <- TOP")
	sess := b.getSession(user)
	if sess == nil {
		return nil, fmt.Errorf("invalid session")
	}

	if msgId > len(sess.Messages) {
		log.Debug().Msg("[POP] -> TOP, no such message")
		return nil, fmt.Errorf("no such message")
	}
	if sess.Deleted[msgId] {
		log.Debug().Msg("[POP] -> TOP, message is deleted")
		return nil, fmt.Errorf("message is deleted")
	}

	msg := sess.Messages[msgId]
	content, err := b.decryptMessage(context.TODO(), sess.PrivateKey, msg)
	if err != nil {
		log.Error().Err(err).Msg("error decrypting message")
		log.Debug().Msg("[POP] -> TOP, error decrypting message")
		return nil, fmt.Errorf("error decrypting message")
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

	content, err := privateKey.Decrypt(msg.Content, senderKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message")
	}
	return string(content), nil
}

func getFirst(s string, i int) string {
	if i > len(s) {
		return s
	}
	return fmt.Sprintf("%s...", s[:i])
}

func getMessageID(msg *pb.DMSMessage) (string, error) {
	b, err := proto.Marshal(msg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func (b *Backend) getMessageByID(sess *Session, id int) (*pb.DMSMessage, error) {
	// Message id ranges from 1 to len(sess.Messages).
	if id <= 0 || id > len(sess.Messages) {
		return nil, fmt.Errorf("invalid message id")
	}
	if sess.Deleted[id-1] {
		return nil, fmt.Errorf("message is deleted")
	}
	return sess.Messages[id-1], nil
}
