package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	defaultKeyLocation       = "/home/ubuntu/ubikom/gateway.key"
	defaultLogFileLocation   = "/home/ubuntu/ubikom/log/ubikom-receive.log"
	defaultSenderName        = "gateway"
	defaultConnectionTimeout = 5000 * time.Millisecond
)

type CmdArgs struct {
	LogFileLocation   string
	KeyLocation       string
	LookupURL         string
	SenderName        string
	ConnectionTimeout time.Duration
}

func main() {
	var args CmdArgs
	flag.StringVar(&args.LogFileLocation, "log-file-location", defaultLogFileLocation, "log file location")
	flag.StringVar(&args.KeyLocation, "key", defaultKeyLocation, "key location")
	flag.StringVar(&args.LookupURL, "lookup-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.DurationVar(&args.ConnectionTimeout, "connection-timeout", defaultConnectionTimeout, "connection timeout")
	flag.StringVar(&args.SenderName, "sender-name", defaultSenderName, "sender name (must correspond to the key)")
	flag.Parse()

	out, err := os.OpenFile(args.LogFileLocation, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: out, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

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
		grpc.WithTimeout(args.ConnectionTimeout),
	}

	log.Debug().Str("url", args.LookupURL).Msg("connecting to lookup service")
	lookupConn, err := grpc.Dial(args.LookupURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	ctx := context.Background()

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
	err = protoutil.SendMessage(ctx, privateKey, body, args.SenderName, receiver, lookupClient)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to send message")
	}
}