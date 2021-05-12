package pop

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

// Backend is a fake backend interface implementation used for test
type Backend struct {
	messages     []string
	deleted      []bool
	dumpClient   pb.DMSDumpServiceClient
	lookupClient pb.LookupServiceClient
	privateKey   *ecc.PrivateKey
	lock         sync.Mutex
}

func NewBackend(dumpClient pb.DMSDumpServiceClient, lookupClient pb.LookupServiceClient,
	privateKey *ecc.PrivateKey) *Backend {
	return &Backend{
		dumpClient:   dumpClient,
		lookupClient: lookupClient,
		privateKey:   privateKey}
}

func (b *Backend) Poll(ctx context.Context) error {
	hash := util.Hash256([]byte("we need a bigger boat"))
	sig, err := b.privateKey.Sign(hash)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to sign message")
	}

	req := &pb.Signed{
		Content: []byte("we need a bigger boat"),
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		Key: b.privateKey.PublicKey().SerializeCompressed(),
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
		senderKey, err := ecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
		if err != nil {
			return fmt.Errorf("invalid receiver public key: %w", err)
		}

		if !protoutil.VerifySignature(msg.GetSignature(), lookupRes.GetKey(), msg.GetContent()) {
			return fmt.Errorf("signature verification failed")
		}

		content, err := b.privateKey.Decrypt(msg.Content, senderKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt message")
		}

		b.lock.Lock()
		b.messages = append(b.messages, string(content))
		b.deleted = append(b.deleted, false)
		b.lock.Unlock()
		count++
	}
	return nil
}

func (b *Backend) StartPolling(ctx context.Context, interval time.Duration) {
	log.Debug().Msg("starting polling for new messages...")
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				log.Debug().Msg("polling for new messages")
				err := b.Poll(ctx)
				if err != nil {
					log.Error().Err(err).Msg("error polling for new messages")
				}
			}
		}
	}()
}

// Returns total message count and total mailbox size in bytes (octets).
// Deleted messages are ignored.
func (b *Backend) Stat(user string) (messages, octets int, err error) {
	log.Debug().Str("user", user).Msg("[POP] <- STAT")
	totalSize := 0
	count := 0
	b.lock.Lock()
	for _, msg := range b.messages {
		count++
		totalSize += len(msg)
	}
	b.lock.Unlock()

	log.Debug().Int("count", count).Int("octets", totalSize).Msg("[POP] -> STAT")
	return count, totalSize, nil
}

// List of sizes of all messages in bytes (octets)
func (b *Backend) List(user string) (octets []int, err error) {
	log.Debug().Str("user", user).Msg("[POP] <- LIST")
	var sizes []int
	b.lock.Lock()
	for i, msg := range b.messages {
		if b.deleted[i] {
			continue
		}
		sizes = append(sizes, len(msg))
	}
	b.lock.Unlock()

	log.Debug().Ints("sizes", sizes).Msg("[POP] -> LIST")
	return sizes, nil
}

// Returns whether message exists and if yes, then return size of the message in bytes (octets)
func (b *Backend) ListMessage(user string, msgId int) (exists bool, octets int, err error) {
	log.Debug().Str("user", user).Int("msg-id", msgId).Msg("[POP] <- LIST-MESSAGE")

	var size int
	b.lock.Lock()
	if msgId > len(b.messages) {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> LIST-MESSAGE, no such message")
		return false, 0, nil
	}
	if b.deleted[msgId] {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> LIST-MESSAGE, message is deleted")
		return false, 0, nil
	}
	size = len(b.messages[msgId])
	b.lock.Unlock()

	log.Debug().Int("size", size).Msg("[POP] -> LIST-MESSAGE")
	return true, size, nil
}

// Retrieve whole message by ID - note that message ID is a message position returned
// by List() function, so be sure to keep that order unchanged while client is connected
// See Lock() function for more details
func (b *Backend) Retr(user string, msgId int) (message string, err error) {
	log.Debug().Str("user", user).Int("msg-id", msgId).Msg("[POP] <- RETR")

	var msg string
	b.lock.Lock()
	if msgId > len(b.messages) {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> RETR, no such message")
		return "", fmt.Errorf("no such message")
	}
	if b.deleted[msgId] {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> RETR, message is deleted")
		return "", fmt.Errorf("message is deleted")
	}
	msg = b.messages[msgId]
	b.lock.Unlock()
	log.Debug().Str("message", getFirst(msg, 16)).Msg("[POP] -> RETR")
	return msg, nil
}

// Delete message by message ID - message should be just marked as deleted until
// Update() is called. Be aware that after Dele() is called, functions like List() etc.
// should ignore all these messages even if Update() hasn't been called yet
func (b *Backend) Dele(user string, msgId int) error {
	log.Debug().Str("user", user).Int("msg-id", msgId).Msg("[POP] <- DELE")
	b.lock.Lock()
	if msgId > len(b.messages) {
		b.lock.Unlock()
		log.Debug().Msg("[POP] -> DELE, no such message")
		return fmt.Errorf("no such message")
	}
	b.deleted[msgId] = true
	b.lock.Unlock()
	log.Debug().Msg("[POP] -> DELE, message marked as deleted")
	return nil
}

// Undelete all messages marked as deleted in single connection
func (b *Backend) Rset(user string) error {
	log.Debug().Str("user", user).Msg("[POP] <- RSET")
	b.lock.Lock()
	for i := range b.deleted {
		b.deleted[i] = false
	}
	b.lock.Unlock()
	log.Debug().Msg("[POP] <- RSET")
	return nil
}

// List of unique IDs of all message, similar to List(), but instead of size there
// is a unique ID which persists the same across all connections. Uid (unique id) is
// used to allow client to be able to keep messages on the server.
func (b *Backend) Uidl(user string) (uids []string, err error) {
	log.Debug().Str("user", user).Msg("[POP] <- UIDL")
	var ids []string
	b.lock.Lock()
	for i, msg := range b.messages {
		if b.deleted[i] {
			continue
		}
		id := fmt.Sprintf("%x", sha256.Sum256([]byte(msg)))
		ids = append(ids, id)
	}
	b.lock.Unlock()
	log.Debug().Strs("ids", ids).Msg("[POP] -> UIDL")
	return ids, nil
}

// Similar to ListMessage, but returns unique ID by message ID instead of size.
func (b *Backend) UidlMessage(user string, msgId int) (exists bool, uid string, err error) {
	log.Debug().Str("user", user).Int("msg-id", msgId).Msg("[POP] <- UIDL-MESSAGE")
	b.lock.Lock()
	if msgId > len(b.messages) {
		b.lock.Unlock()
		log.Error().Msg("[POP] -> UIDL-MESSAGE, no such message")
		return false, "", nil
	}
	if b.deleted[msgId] {
		b.lock.Unlock()
		log.Error().Msg("[POP] -> UIDL-MESSAGE, message is deleted")
		return false, "", nil
	}
	id := fmt.Sprintf("%x", sha256.Sum256([]byte(b.messages[msgId])))
	b.lock.Unlock()
	log.Debug().Str("id", id).Msg("[POP] -> UIDL-MESSAGE")
	return true, id, nil
}

// Write all changes to persistent storage, i.e. delete all messages marked as deleted.
func (b *Backend) Update(user string) error {
	log.Debug().Str("user", user).Msg("[POP] <- UPDATE")
	b.lock.Lock()
	var newMessages []string
	count := 0
	for i, msg := range b.messages {
		if b.deleted[i] {
			count++
			continue
		}
		newMessages = append(newMessages, msg)
	}
	b.messages = newMessages
	b.deleted = make([]bool, len(newMessages))
	b.lock.Unlock()
	log.Debug().Int("deleted", count).Msg("[POP] -> UPDATE")
	return nil
}

// Lock is called immediately after client is connected. The best way what to use Lock() for
// is to read all the messages into cache after client is connected. If another user
// tries to lock the storage, you should return an error to avoid data race.
func (b *Backend) Lock(user string) error {
	log.Debug().Str("user", user).Msg("[POP] <- LOCK")
	log.Debug().Msg("[POP] -> LOCK, ignored (not implemented)")
	return nil
}

// Release lock on storage, Unlock() is called after client is disconnected.
func (b *Backend) Unlock(user string) error {
	log.Debug().Str("user", user).Msg("[POP] <- UNLOCK")
	log.Debug().Msg("[POP] -> UNLOCK, ignored (not implemented)")
	return nil
}

func getFirst(s string, i int) string {
	if i > len(s) {
		return s
	}
	return fmt.Sprintf("%s...", s[:i])
}
