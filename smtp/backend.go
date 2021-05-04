package smtp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/rs/zerolog/log"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/util"
)

type Backend struct {
	user         string
	password     string
	lookupClient pb.LookupServiceClient
	dumpClient   pb.DMSDumpServiceClient
	privateKey   *ecc.PrivateKey
}

func NewBackend(user, password string, lookupClient pb.LookupServiceClient,
	dumpClient pb.DMSDumpServiceClient, privateKey *ecc.PrivateKey) *Backend {
	return &Backend{
		user:         user,
		password:     password,
		lookupClient: lookupClient,
		dumpClient:   dumpClient,
		privateKey:   privateKey}
}

func (b *Backend) Login(state *gosmtp.ConnectionState, username, password string) (gosmtp.Session, error) {
	log.Debug().Str("user", username).Msg("[SMTP] <- LOGIN")
	//ok := username != b.user || password != b.password
	ok := true
	log.Debug().Bool("authorized", ok).Msg("[SMTP] -> LOGIN")
	if !ok {
		return nil, errors.New("Invalid username or password")
	}
	return &Session{
		lookupClient: b.lookupClient,
		dumpClient:   b.dumpClient,
		privateKey:   b.privateKey,
	}, nil
}

// AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (bkd *Backend) AnonymousLogin(state *gosmtp.ConnectionState) (gosmtp.Session, error) {
	log.Debug().Msg("[SMTP] <- ANON-LOGIN")
	log.Debug().Msg("[SMTP] -> ANON-LOGIN authorizatoin required")
	return nil, gosmtp.ErrAuthRequired
}

// A Session is returned after successful login.
type Session struct {
	from, to     string
	lookupClient pb.LookupServiceClient
	dumpClient   pb.DMSDumpServiceClient
	privateKey   *ecc.PrivateKey
}

func (s *Session) Mail(from string, opts gosmtp.MailOptions) error {
	log.Debug().Str("from", from).Msg("[SMTP] <- MAIL")
	s.from = from
	log.Debug().Msg("[SMTP] -> MAIL")
	return nil
}

func (s *Session) Rcpt(to string) error {
	log.Debug().Str("to", to).Msg("[SMTP] <- RCPT")
	s.to = to
	log.Debug().Msg("[SMTP] -> RCPT")
	return nil
}

func (s *Session) Data(r io.Reader) error {
	log.Debug().Msg("[SMTP] <- DATA")
	var body []byte
	var err error
	if body, err = ioutil.ReadAll(r); err != nil {
		log.Error().Err(err).Msg("[SMTP] -> DATA")
		return err
	} else {
		log.Debug().Str("data", string(body)).Msg("[SMTP] -> DATA")
	}

	// Send the actual message.
	sender := s.from
	if strings.HasSuffix(sender, "@x") {
		sender = strings.Replace(sender, "@x", "", -1)
	}

	receiver := s.to
	if strings.HasSuffix(receiver, "@x") {
		receiver = strings.Replace(receiver, "@x", "", -1)
	}

	log.Debug().Str("sender", sender).Str("receiver", receiver).Msg("about to send message")

	lookupRes, err := s.lookupClient.LookupName(context.TODO(), &pb.LookupNameRequest{Name: receiver})
	if err != nil {
		log.Error().Err(err).Str("name", s.to).Msg("failed to get receiver public key")
		return fmt.Errorf("failed to send message")
	}
	if lookupRes.Result != pb.ResultCode_RC_OK {
		log.Error().Str("result", lookupRes.Result.String()).Msg("failed to get receiver public key")
		return fmt.Errorf("failed to send message")
	}

	receiverKey, err := ecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
	if err != nil {
		log.Error().Err(err).Msg("invalid receiver public key")
		return fmt.Errorf("failed to send messsage")
	}

	encryptedBody, err := s.privateKey.Encrypt(body, receiverKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to encrypt message")
		return fmt.Errorf("failed to send messsage")
	}

	hash := util.Hash256(encryptedBody)
	sig, err := s.privateKey.Sign(hash)
	if err != nil {
		log.Error().Err(err).Msg("failed to sign message")
		return fmt.Errorf("failed to send messsage")
	}

	msg := &pb.DMSMessage{
		Sender:   sender,
		Receiver: receiver,
		Content:  encryptedBody,
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
	}
	res, err := s.dumpClient.Send(context.TODO(), msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to send message")
		return fmt.Errorf("failed to send messsage")
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Error().Str("code", res.GetResult().Enum().String()).Msg("server returned error")
		return fmt.Errorf("failed to send messsage")
	}
	log.Debug().Msg("message sent")

	return nil
}

func (s *Session) Reset() {
	log.Debug().Msg("reset")
}

func (s *Session) Logout() error {
	log.Debug().Msg("logout")
	return nil
}
