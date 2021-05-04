package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/emersion/go-smtp"
	gosmpt "github.com/emersion/go-smtp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pop"
	"teralyt.com/ubikom/util"
)

// The Backend implements SMTP server methods.
type Backend struct{}

// Login handles a login command with username and password.
func (bkd *Backend) Login(state *gosmpt.ConnectionState, username, password string) (gosmpt.Session, error) {
	log.Debug().Str("user", username).Str("password", password).Msg("login")
	if username != "lgx" || password != "pumpkin123" {
		return nil, errors.New("Invalid username or password")
	}
	return &Session{}, nil
}

// AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (bkd *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	log.Debug().Msg("anonymous login")
	return nil, smtp.ErrAuthRequired
}

// A Session is returned after successful login.
type Session struct{}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	log.Debug().Str("from", from).Msg("mail")
	return nil
}

func (s *Session) Rcpt(to string) error {
	log.Debug().Str("to", to).Msg("rcpt")
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		log.Debug().Str("data", string(b)).Msg("data")
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

type CmdArgs struct {
	DumpURL     string
	LookupURL   string
	KeyLocation string
}

func main() {
	var args CmdArgs
	flag.StringVar(&args.DumpURL, "dump-url", "localhost:8826", "dump service URL")
	flag.StringVar(&args.LookupURL, "lookup-url", "localhost:8825", "lookup service URL")
	flag.StringVar(&args.KeyLocation, "key", "", "private key location")
	flag.Parse()

	if args.KeyLocation == "" {
		var err error
		args.KeyLocation, err = util.GetDefaultKeyLocation()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to determine key location")
		}
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}

	dumpConn, err := grpc.Dial(args.DumpURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the dump server")
	}
	defer dumpConn.Close()

	dumpClient := pb.NewDMSDumpServiceClient(dumpConn)

	lookupConn, err := grpc.Dial(args.LookupURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	key, err := ecc.LoadPrivateKey(args.KeyLocation)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load private key")
	}

	popOpts := &pop.ServerOptions{
		Ctx:          context.Background(),
		Domain:       "localhost",
		Port:         1100,
		User:         "lgx",
		Password:     "pumpkin123",
		DumpClient:   dumpClient,
		LookupClient: lookupClient,
		Key:          key,
		PollInterval: time.Minute,
	}

	backendServer := pop.NewServer(popOpts)
	go func() {
		backendServer.ListenAndServe()
	}()

	be := &Backend{}
	s := gosmpt.NewServer(be)
	s.Addr = ":1025"
	s.Domain = "localhost"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true
	if err := s.ListenAndServe(); err != nil {
		log.Fatal().Err(err)
	}
}
