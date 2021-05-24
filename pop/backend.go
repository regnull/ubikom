package pop

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/btcsuite/btcutil/base58"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

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
	Messages   []string
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
}

func NewBackend(dumpClient pb.DMSDumpServiceClient, lookupClient pb.LookupServiceClient,
	privateKey *easyecc.PrivateKey, user, password string) *Backend {
	return &Backend{
		dumpClient:   dumpClient,
		lookupClient: lookupClient,
		privateKey:   privateKey,
		user:         user,
		password:     password,
		sessions:     make(map[string]*Session)}
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
		if res.GetResult().GetResult() != pb.ResultCode_RC_OK {
			log.Error().Interface("result", res.GetResult()).Msg("failed to look up key")
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
	content := "we will need a bigger boat"
	hash := util.Hash256([]byte(content))

	// Get private key for this user.
	var privateKey *easyecc.PrivateKey
	sess := b.getSession(user)

	if sess == nil {
		log.Error().Str("user", user).Msg("invalid session")
		return fmt.Errorf("invalid session")
	}
	privateKey = sess.PrivateKey

	sig, err := privateKey.Sign(hash)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to sign message")
	}

	req := &pb.Signed{
		Content: []byte(content),
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		Key: privateKey.PublicKey().SerializeCompressed(),
	}

	count := 0
	for {
		res, err := b.dumpClient.Receive(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}
		if res.GetResult().GetResult() == pb.ResultCode_RC_RECORD_NOT_FOUND {
			if count == 0 {
				log.Debug().Msg("no new messages")
			} else {
				log.Debug().Int("count", count).Msg("got new messages")
			}
			break
		}
		if res.Result.Result != pb.ResultCode_RC_OK {
			return fmt.Errorf("server returned error: %s", res.GetResult().GetResult().Enum().String())
		}
		msg := &pb.DMSMessage{}
		err = proto.Unmarshal(res.GetContent(), msg)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}

		lookupRes, err := b.lookupClient.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
		if err != nil {
			return fmt.Errorf("failed to get receiver public key: %w", err)
		}
		if lookupRes.GetResult().GetResult() != pb.ResultCode_RC_OK {
			return fmt.Errorf("failed to get receiver public key: %s", lookupRes.GetResult().String())
		}
		senderKey, err := easyecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
		if err != nil {
			return fmt.Errorf("invalid receiver public key: %w", err)
		}

		if !protoutil.VerifySignature(msg.GetSignature(), lookupRes.GetKey(), msg.GetContent()) {
			return fmt.Errorf("signature verification failed")
		}

		content, err := privateKey.Decrypt(msg.Content, senderKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt message")
		}

		b.lock.Lock()
		sess.Messages = append(sess.Messages, string(content))
		sess.Deleted = append(sess.Deleted, false)
		b.lock.Unlock()
		count++
	}
	return nil
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
		count++
		totalSize += len(msg)
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
	for i, msg := range sess.Messages {
		if sess.Deleted[i] {
			continue
		}
		sizes = append(sizes, len(msg))
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

	var size int
	if msgId > len(sess.Messages) {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> LIST-MESSAGE, no such message")
		return false, 0, nil
	}
	if sess.Deleted[msgId] {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> LIST-MESSAGE, message is deleted")
		return false, 0, nil
	}
	size = len(sess.Messages[msgId])

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

	var msg string
	if msgId > len(sess.Messages) {
		log.Debug().Msg("[POP] -> RETR, no such message")
		return "", fmt.Errorf("no such message")
	}
	if sess.Deleted[msgId] {
		log.Debug().Msg("[POP] -> RETR, message is deleted")
		return "", fmt.Errorf("message is deleted")
	}
	msg = sess.Messages[msgId]
	log.Debug().Str("message", getFirst(msg, 16)).Msg("[POP] -> RETR")
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
		return nil, fmt.Errorf("invalid session")
	}
	var ids []string
	for i, msg := range sess.Messages {
		if sess.Deleted[i] {
			continue
		}
		id := fmt.Sprintf("%x", sha256.Sum256([]byte(msg)))
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

	if msgId > len(sess.Messages) {
		log.Error().Msg("[POP] -> UIDL-MESSAGE, no such message")
		return false, "", nil
	}
	if sess.Deleted[msgId] {
		log.Error().Msg("[POP] -> UIDL-MESSAGE, message is deleted")
		return false, "", nil
	}
	id := fmt.Sprintf("%x", sha256.Sum256([]byte(sess.Messages[msgId])))
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

	var newMessages []string
	count := 0
	for i, msg := range sess.Messages {
		if sess.Deleted[i] {
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

func getFirst(s string, i int) string {
	if i > len(s) {
		return s
	}
	return fmt.Sprintf("%s...", s[:i])
}
