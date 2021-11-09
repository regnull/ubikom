package imap

import (
	"crypto/tls"
	"fmt"

	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/server"
	"github.com/rs/zerolog/log"
)

type ServerOptions struct {
	Domain   string
	Port     int
	CertFile string
	KeyFile  string
}

type Server struct {
	opts   *ServerOptions
	server *server.Server
}

func NewServer(opts *ServerOptions) *Server {
	s := server.New(memory.New())
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
