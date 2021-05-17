package smtp

import (
	"crypto/tls"
	"fmt"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
)

type ServerOptions struct {
	Domain       string
	Port         int
	User         string
	Password     string
	LookupClient pb.LookupServiceClient
	DumpClient   pb.DMSDumpServiceClient
	PrivateKey   *ecc.PrivateKey
	CertFile     string
	KeyFile      string
}

type Server struct {
	opts    *ServerOptions
	server  *gosmtp.Server
	backend *Backend
}

func NewServer(opts *ServerOptions) (*Server, error) {
	backend := NewBackend(opts.User, opts.Password, opts.LookupClient, opts.DumpClient, opts.PrivateKey)
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
