package pop

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DevelHell/popgun"
	"github.com/rs/zerolog/log"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
)

type authorizator struct {
	user     string
	password string
}

func (a *authorizator) Authorize(user, pass string) bool {
	log.Debug().Str("user", user).Msg("[POP] <- LOGIN")
	ok := user == a.user && pass == a.password
	log.Debug().Bool("authorized", ok).Msg("[POP] -> LOGIN")
	return ok
}

type ServerOptions struct {
	Ctx          context.Context
	Domain       string
	Port         int
	User         string
	Password     string
	DumpClient   pb.DMSDumpServiceClient
	LookupClient pb.LookupServiceClient
	Key          *ecc.PrivateKey
	PollInterval time.Duration
}

type Server struct {
	opts   *ServerOptions
	server *popgun.Server
}

func NewServer(opts *ServerOptions) *Server {
	auth := &authorizator{
		user:     opts.User,
		password: opts.Password}
	cfg := popgun.Config{
		ListenInterface: fmt.Sprintf("%s:%d", opts.Domain, opts.Port)}
	backend := NewBackend(opts.DumpClient, opts.LookupClient, opts.Key)
	backend.StartPolling(opts.Ctx, opts.PollInterval)
	popServer := popgun.NewServer(cfg, auth, backend)
	return &Server{opts: opts, server: popServer}
}

func (s *Server) ListenAndServe() error {
	log.Info().Str("domain", s.opts.Domain).Int("port", s.opts.Port).Msg("POP server starting")
	err := s.server.Start()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
	return nil
}
