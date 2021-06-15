package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/gateway"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type CmdArgs struct {
	KeyLocation            string
	DumpURL                string
	LookupURL              string
	ConnectionTimeoutMsec  int
	GlobalRateLimitPerHour int
	PollInterval           time.Duration
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.KeyLocation, "key", "", "key location")
	flag.StringVar(&args.DumpURL, "dump-url", globals.PublicDumpServiceURL, "dump service URL")
	flag.StringVar(&args.LookupURL, "lookup-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.IntVar(&args.ConnectionTimeoutMsec, "connection-timeout-msec", 5000, "connection timeout, milliseconds")
	flag.IntVar(&args.GlobalRateLimitPerHour, "global-rate-limit-per-hour", 100, "global rate limit, per hour")
	flag.DurationVar(&args.PollInterval, "poll-interval", time.Minute, "poll interval")
	flag.Parse()

	if args.KeyLocation == "" {
		log.Fatal().Msg("--key argument must be specified")
	}
	privateKey, err := easyecc.NewPrivateKeyFromFile(args.KeyLocation, "")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load private key")
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Millisecond * time.Duration(args.ConnectionTimeoutMsec)),
	}

	log.Debug().Str("url", args.DumpURL).Msg("connecting to dump service")
	dumpConn, err := grpc.Dial(args.DumpURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the dump server")
	}
	defer dumpConn.Close()

	dumpClient := pb.NewDMSDumpServiceClient(dumpConn)

	log.Debug().Str("url", args.LookupURL).Msg("connecting to lookup service")
	lookupConn, err := grpc.Dial(args.LookupURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	senderOpts := &gateway.SenderOptions{
		PrivateKey:             privateKey,
		LookupClient:           lookupClient,
		DumpClient:             dumpClient,
		GlobalRateLimitPerHour: args.GlobalRateLimitPerHour,
		PollInterval:           args.PollInterval,
		ExternalSender:         gateway.NewSendmailSender(),
	}

	sender := gateway.NewSender(senderOpts)
	err = sender.Run(context.Background())
	log.Info().Err(err).Msg("exiting")
}
