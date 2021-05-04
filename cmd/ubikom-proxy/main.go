package main

import (
	"context"
	"flag"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pop"
	"teralyt.com/ubikom/smtp"
	"teralyt.com/ubikom/util"
)

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

	var wg sync.WaitGroup
	wg.Add(2)

	popServer := pop.NewServer(popOpts)
	go func() {
		err := popServer.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("POP server failed to initialize")
		}
		wg.Done()
	}()

	smtpOpts := &smtp.ServerOptions{
		Domain:       "localhost",
		Port:         1025,
		User:         "lgx",
		Password:     "pumpkin123",
		LookupClient: lookupClient,
		DumpClient:   dumpClient,
		PrivateKey:   key,
	}
	smtpServer := smtp.NewServer(smtpOpts)
	go func() {
		err := smtpServer.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("SMTP server failed to initialize")
		}
		wg.Done()
	}()

	wg.Wait()
}
