package smtp

import (
	"fmt"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/rs/zerolog/log"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
)

type ServerOptions struct {
	Domain       string
	Port         int
	User         string
	Password     string
	LookupClient pb.LookupServiceClient
	DumpClient   pb.DMSDumpServiceClient
	PrivateKey   *ecc.PrivateKey
}

type Server struct {
	opts    *ServerOptions
	server  *gosmtp.Server
	backend *Backend
}

func NewServer(opts *ServerOptions) *Server {
	backend := NewBackend(opts.User, opts.Password, opts.LookupClient, opts.DumpClient, opts.PrivateKey)
	server := gosmtp.NewServer(backend)
	server.Addr = fmt.Sprintf(":%d", opts.Port)
	server.Domain = opts.Domain
	server.ReadTimeout = 10 * time.Second
	server.WriteTimeout = 10 * time.Second
	server.MaxMessageBytes = 1024 * 1024
	server.MaxRecipients = 50
	server.AllowInsecureAuth = true
	return &Server{
		opts:    opts,
		backend: backend,
		server:  server}
}

func (s *Server) ListenAndServe() error {
	log.Info().Str("domain", s.opts.Domain).Int("port", s.opts.Port).Msg("SMTP server starting")
	return s.server.ListenAndServe()
}
