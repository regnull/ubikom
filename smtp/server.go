package smtp

import (
	"crypto/tls"
	"fmt"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/event"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
)

type ServerOptions struct {
	Domain                string
	Port                  int
	User                  string
	Password              string
	LookupClient          pb.LookupServiceClient
	DumpClient            pb.DMSDumpServiceClient
	PrivateKey            *easyecc.PrivateKey
	CertFile              string
	KeyFile               string
	EventTarget           string
	UbikomName            string
	EventSenderPrivateKey *easyecc.PrivateKey
}

type Server struct {
	opts    *ServerOptions
	server  *gosmtp.Server
	backend *Backend
}

func NewServer(opts *ServerOptions) (*Server, error) {
	var eventSender *event.Sender
	if opts.EventSenderPrivateKey != nil && opts.UbikomName != "" && opts.EventTarget != "" {
		log.Debug().Msg("creating event sender")
		eventSender = event.NewSender(opts.EventTarget, opts.UbikomName, "server",
			opts.EventSenderPrivateKey, opts.LookupClient)
	} else {
		log.Warn().Msg("cannot create event sender")
	}
	backend := NewBackend(opts.User, opts.Password, opts.LookupClient, opts.DumpClient,
		opts.PrivateKey, eventSender)
	server := gosmtp.NewServer(backend)
	server.Addr = fmt.Sprintf(":%d", opts.Port)
	server.Domain = opts.Domain
	server.ReadTimeout = 10 * time.Second
	server.WriteTimeout = 10 * time.Second
	server.MaxMessageBytes = 1024 * 1024
	server.MaxRecipients = 50
	server.AllowInsecureAuth = true

	if opts.CertFile != "" && opts.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, err
		}
		server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	return &Server{
		opts:    opts,
		backend: backend,
		server:  server}, nil
}

func (s *Server) ListenAndServe() error {
	log.Info().Str("domain", s.opts.Domain).Int("port", s.opts.Port).Msg("SMTP server starting")
	if s.server.TLSConfig != nil {
		return s.server.ListenAndServeTLS()
	} else {
		return s.server.ListenAndServe()
	}
}
