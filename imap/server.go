package imap

import (
	"crypto/tls"
	"fmt"

	"github.com/emersion/go-imap/server"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap/db"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
)

type ServerOptions struct {
	Domain       string
	Port         int
	User         string
	Password     string
	PrivateKey   *easyecc.PrivateKey
	CertFile     string
	KeyFile      string
	LookupClient pb.LookupServiceClient
	DumpClient   pb.DMSDumpServiceClient
	Badger       *db.Badger
}

type Server struct {
	opts   *ServerOptions
	server *server.Server
}

func NewServer(opts *ServerOptions) *Server {
	s := server.New(NewBackend(opts.DumpClient, opts.LookupClient, opts.PrivateKey, opts.User, opts.Password, opts.Badger))
	s.Addr = fmt.Sprintf("%s:%d", opts.Domain, opts.Port)
	s.AllowInsecureAuth = true
	return &Server{opts: opts, server: s}
}

func (s *Server) ListenAndServe() error {
	log.Info().Str("domain", s.opts.Domain).Int("port", s.opts.Port).Msg("IMAP server starting")
	if s.opts.CertFile != "" && s.opts.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(s.opts.CertFile, s.opts.KeyFile)
		if err != nil {
			log.Error().Err(err).Msg("could not load certificate/key")
			return err
		}
		s.server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		return s.server.ListenAndServeTLS()
	}
	s.server.TLSConfig = nil
	s.server.AllowInsecureAuth = true
	return s.server.ListenAndServe()
}
