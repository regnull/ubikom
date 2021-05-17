package pop

import (
	"context"
	"fmt"
	"sync"

	"github.com/regnull/popgun"
	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
)

type ServerOptions struct {
	Ctx          context.Context
	Domain       string
	Port         int
	User         string
	Password     string
	DumpClient   pb.DMSDumpServiceClient
	LookupClient pb.LookupServiceClient
	Key          *ecc.PrivateKey
	CertFile     string
	KeyFile      string
}

type Server struct {
	opts   *ServerOptions
	server *popgun.Server
}

func NewServer(opts *ServerOptions) *Server {
	cfg := popgun.Config{
		ListenInterface: fmt.Sprintf("%s:%d", opts.Domain, opts.Port)}
	backend := NewBackend(opts.DumpClient, opts.LookupClient, opts.Key, opts.User, opts.Password)
	popServer := popgun.NewServer(cfg, backend, backend)
	return &Server{opts: opts, server: popServer}
}

func (s *Server) ListenAndServe() error {
	log.Info().Str("domain", s.opts.Domain).Int("port", s.opts.Port).Msg("POP server starting")
	var err error
	if s.opts.CertFile != "" && s.opts.KeyFile != "" {
		err = s.server.StartTLS(s.opts.CertFile, s.opts.KeyFile)
	} else {
		err = s.server.Start()
	}
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
	return nil
}
