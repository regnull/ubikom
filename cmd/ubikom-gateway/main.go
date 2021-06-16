package main

import (
	"context"
	"flag"
	"io"
	"os"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/gateway"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	defaultKeyLocation = "/home/ubuntu/ubikom/gateway.key"
	defaultSenderName  = "gateway"
)

type CmdArgs struct {
	KeyLocation            string
	DumpURL                string
	LookupURL              string
	ConnectionTimeoutMsec  int
	GlobalRateLimitPerHour int
	PollInterval           time.Duration
	Receive                bool
	SenderName             string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.KeyLocation, "key", defaultKeyLocation, "key location")
	flag.StringVar(&args.DumpURL, "dump-url", globals.PublicDumpServiceURL, "dump service URL")
	flag.StringVar(&args.LookupURL, "lookup-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.IntVar(&args.ConnectionTimeoutMsec, "connection-timeout-msec", 5000, "connection timeout, milliseconds")
	flag.IntVar(&args.GlobalRateLimitPerHour, "global-rate-limit-per-hour", 100, "global rate limit, per hour")
	flag.DurationVar(&args.PollInterval, "poll-interval", time.Minute, "poll interval")
	flag.BoolVar(&args.Receive, "receive", false, "receive mail (if false, will monitor and send mail)")
	flag.StringVar(&args.SenderName, "sender-name", defaultSenderName, "sender name (must correspond to the key)")
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

	log.Debug().Str("url", args.LookupURL).Msg("connecting to lookup service")
	lookupConn, err := grpc.Dial(args.LookupURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	ctx := context.Background()

	if args.Receive {
		// Receive mail and exit.
		receive(ctx, privateKey, lookupClient, args.SenderName)
		return
	}

	log.Debug().Str("url", args.DumpURL).Msg("connecting to dump service")
	dumpConn, err := grpc.Dial(args.DumpURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the dump server")
	}
	defer dumpConn.Close()

	dumpClient := pb.NewDMSDumpServiceClient(dumpConn)

	// Monitor for new mail, send it out.
	send(ctx, privateKey, lookupClient, dumpClient, &args)
}

func receive(ctx context.Context, privateKey *easyecc.PrivateKey, lookupClient pb.LookupServiceClient, sender string) {
	// Read the email from stdin.
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read stdin")
	}

	// Get the receiver name.
	receiver, err := mail.ExtractReceiverInternalName(string(body))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get receiver name")
	}

	// Send the message.
	log.Debug().Str("receiver", receiver).Msg("sending mail")
	err = protoutil.SendMessage(ctx, privateKey, body, sender, receiver, lookupClient)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to send message")
	}
}

func send(ctx context.Context, privateKey *easyecc.PrivateKey, lookupClient pb.LookupServiceClient,
	dumpClient pb.DMSDumpServiceClient, args *CmdArgs) {
	senderOpts := &gateway.SenderOptions{
		PrivateKey:             privateKey,
		LookupClient:           lookupClient,
		DumpClient:             dumpClient,
		GlobalRateLimitPerHour: args.GlobalRateLimitPerHour,
		PollInterval:           args.PollInterval,
		ExternalSender:         gateway.NewSendmailSender(),
	}

	sender := gateway.NewSender(senderOpts)
	err := sender.Run(ctx)
	log.Info().Err(err).Msg("exiting")
}
