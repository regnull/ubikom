package smtp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/btcsuite/btcutil/base58"
	gosmtp "github.com/emersion/go-smtp"
	"github.com/regnull/easyecc"
	umail "github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/rs/zerolog/log"
)

type Backend struct {
	user         string
	password     string
	lookupClient pb.LookupServiceClient
	dumpClient   pb.DMSDumpServiceClient
	privateKey   *easyecc.PrivateKey
}

func NewBackend(user, password string, lookupClient pb.LookupServiceClient,
	dumpClient pb.DMSDumpServiceClient, privateKey *easyecc.PrivateKey) *Backend {
	return &Backend{
		user:         user,
		password:     password,
		lookupClient: lookupClient,
		dumpClient:   dumpClient,
		privateKey:   privateKey}
}

func (b *Backend) Login(state *gosmtp.ConnectionState, username, password string) (gosmtp.Session, error) {
	log.Debug().Str("user", username).Msg("[SMTP] <- LOGIN")
	var privateKey *easyecc.PrivateKey
	ok := false
	if b.privateKey != nil {
		ok = username == b.user && password == b.password
		privateKey = b.privateKey
	} else {
		// TODO: Validate username/password (make sure this key is registered).
		salt := base58.Decode(username)
		ok = len(salt) >= 4 && len(password) >= 8
		privateKey = easyecc.NewPrivateKeyFromPassword([]byte(password), salt)
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
	from         string
	to           []string
	lookupClient pb.LookupServiceClient
	dumpClient   pb.DMSDumpServiceClient
	privateKey   *easyecc.PrivateKey
}

func (s *Session) Mail(from string, opts gosmtp.MailOptions) error {
	log.Debug().Str("from", from).Msg("[SMTP] <- MAIL")
	s.from = from
	log.Debug().Msg("[SMTP] -> MAIL")
	return nil
}

func (s *Session) Rcpt(to string) error {
	log.Debug().Str("to", to).Msg("[SMTP] <- RCPT")
	s.to = append(s.to, to)
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
	sender := umail.StripDomain(s.from)

	var internalAddresses []string
	var externalAddresses []string

	for _, to := range s.to {
		if umail.IsInternal(to) {
			internalAddresses = append(internalAddresses, umail.StripDomain(to))
		} else {
			externalAddresses = append(externalAddresses, umail.StripDomain(to))
		}
	}

	// Send to internal addresses one by one.
	for _, addr := range internalAddresses {
		log.Debug().Str("sender", sender).Str("receiver", addr).Msg("about to send message")

		err = protoutil.SendMessage(context.Background(), s.privateKey, body, sender, addr, s.lookupClient)
		if err != nil {
			log.Error().Err(err).Msg("failed to send message")
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	// If we have any external addresses, send the message to the gateway, and it will deal with it.
	if len(externalAddresses) > 0 {
		log.Debug().Str("sender", sender).Str("receiver", "gateway").Msg("about to send message to the gateway")

		err = protoutil.SendMessage(context.Background(), s.privateKey, body, sender, "gateway", s.lookupClient)
		if err != nil {
			log.Error().Err(err).Msg("failed to send message")
			return fmt.Errorf("failed to send message: %w", err)
		}
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
