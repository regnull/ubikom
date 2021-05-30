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

type CmdArgs struct {
	KeyLocation           string
	LookupURL             string
	ConnectionTimeoutMsec int
}

func main() {
	out, err := os.OpenFile("/home/ubuntu/ubikom/log/ubikom-receive.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: out, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.KeyLocation, "key", "/home/ubuntu/ubikom/gateway.key", "key location")
	flag.StringVar(&args.LookupURL, "lookup-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.IntVar(&args.ConnectionTimeoutMsec, "connection-timeout-msec", 5000, "connection timeout, milliseconds")
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

	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read stdin")
	}

	receiver, err := mail.ExtractReceiverInternalName(string(body))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get receiver name")
	}

	log.Debug().Str("receiver", receiver).Msg("sending mail")
	err = protoutil.SendMessage(ctx, privateKey, body, "gateway", receiver, lookupClient)

	/*
		out, err := os.OpenFile("/home/ubuntu/ubikom/receive.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: out, TimeFormat: "15:04:05"})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read stdin")
		}
		log.Debug().Str("content", string(bytes)).Msg("read data from stdin")
	*/
}
