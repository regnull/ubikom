package smtp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/event"
	umail "github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
)

type Backend struct {
	user         string
	password     string
	lookupClient pb.LookupServiceClient
	dumpClient   pb.DMSDumpServiceClient
	privateKey   *easyecc.PrivateKey
	eventSender  *event.Sender
}

func NewBackend(user, password string, lookupClient pb.LookupServiceClient,
	dumpClient pb.DMSDumpServiceClient, privateKey *easyecc.PrivateKey,
	eventSender *event.Sender) *Backend {
	return &Backend{
		user:         user,
		password:     password,
		lookupClient: lookupClient,
		dumpClient:   dumpClient,
		privateKey:   privateKey,
		eventSender:  eventSender}
}

func (b *Backend) Login(state *gosmtp.ConnectionState, username, password string) (gosmtp.Session, error) {
	log.Debug().Str("user", username).Msg("[SMTP] <- LOGIN")
	var privateKey *easyecc.PrivateKey
	ok := false
	if b.privateKey != nil {
		ok = username == b.user && password == b.password
		privateKey = b.privateKey
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var err error
		privateKey, err = util.GetKeyFromNamePassword(ctx, username, password, b.lookupClient)
		ok = err == nil
		if ok && b.eventSender != nil {
			err := b.eventSender.SMTPLogin(context.TODO(), privateKey.PublicKey().Address())
			if err != nil {
				log.Error().Err(err).Msg("failed to send SMTP login event")
			}
		}
	}
	log.Debug().Bool("authorized", ok).Msg("[SMTP] -> LOGIN")
	if !ok {
		return nil, errors.New("invalid username or password")
	}
	return &Session{
		lookupClient: b.lookupClient,
		dumpClient:   b.dumpClient,
		privateKey:   privateKey,
		eventSender:  b.eventSender,
	}, nil
}

// AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (bkd *Backend) AnonymousLogin(state *gosmtp.ConnectionState) (gosmtp.Session, error) {
	log.Debug().Msg("[SMTP] <- ANON-LOGIN")
	log.Debug().Msg("[SMTP] -> ANON-LOGIN authorization required")
	return nil, gosmtp.ErrAuthRequired
}

// A Session is returned after successful login.
type Session struct {
	from         string
	to           []string
	lookupClient pb.LookupServiceClient
	dumpClient   pb.DMSDumpServiceClient
	privateKey   *easyecc.PrivateKey
	eventSender  *event.Sender
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
			externalAddresses = append(externalAddresses, to)
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

		err = protoutil.SendEmail(context.Background(), s.privateKey, body, sender, "gateway", s.lookupClient)
		if err != nil {
			log.Error().Err(err).Msg("failed to send message")
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	if s.eventSender != nil {
		var dest []string
		dest = append(dest, internalAddresses...)
		if len(externalAddresses) > 0 {
			dest = append(dest, "gateway")
		}
		err := s.eventSender.SMTPSend(context.TODO(), sender, strings.Join(dest, ", "),
			strings.Join(externalAddresses, ", "))
		if err != nil {
			log.Error().Err(err).Msg("failed to send SMTP login event")
		}
	}

	return nil
}

func (s *Session) Reset() {
	log.Debug().Msg("[SMTP] reset")
}

func (s *Session) Logout() error {
	log.Debug().Msg("[SMTP] logout")
	return nil
}
