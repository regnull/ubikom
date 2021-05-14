package smtp

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	gosmtp "github.com/emersion/go-smtp"
	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/rs/zerolog/log"
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
	var privateKey *ecc.PrivateKey
	ok := false
	if b.privateKey != nil {
		ok = username == b.user && password == b.password
		privateKey = b.privateKey
	} else {
		// TODO: Validate username/password (make sure this key is registered).
		log.Debug().Str("user", username).Str("pass", password).Msg("got credentials")
		salt := base58.Decode(username)
		ok = len(salt) >= 4 && len(password) >= 8
		privateKey = ecc.NewPrivateKeyFromPassword([]byte(password), salt)
	}
	log.Debug().Bool("authorized", ok).Msg("[SMTP] -> LOGIN")
	if !ok {
		return nil, errors.New("Invalid username or password")
	}
	return &Session{
		lookupClient: b.lookupClient,
		dumpClient:   b.dumpClient,
		privateKey:   privateKey,
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
		log.Debug().Int("size", len(body)).Msg("[SMTP] -> DATA")
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

	err = protoutil.SendMessage(context.Background(), s.privateKey, body, sender, receiver, s.lookupClient)
	if err != nil {
		log.Error().Err(err).Msg("failed to send message")
	}
	return nil
}

func (s *Session) Reset() {
	log.Debug().Msg("reset")
}

func (s *Session) Logout() error {
	log.Debug().Msg("logout")
	return nil
}
