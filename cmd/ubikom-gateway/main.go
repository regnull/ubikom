package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"os/exec"
	"time"

	"golang.org/x/time/rate"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

	// For now, we use a global rate limiter to prevent spam.
	globalRateLimiter := rate.NewLimiter(rate.Every(time.Hour), args.GlobalRateLimitPerHour)

	ctx := context.Background()
	for {
		for {
			res, err := dumpClient.Receive(ctx, &pb.ReceiveRequest{IdentityProof: protoutil.IdentityProof(privateKey)})
			if err != nil && util.StatusCodeFromError(err) == codes.NotFound {
				log.Info().Msg("no new messages")
				break
			}
			if err != nil {
				log.Error().Err(err).Msg("failed to receive message")
				break
			}
			msg := res.GetMessage()
			content, err := protoutil.DecryptMessage(ctx, lookupClient, privateKey, msg)
			if err != nil {
				log.Error().Err(err).Msg("failed to decrypt message")
				break
			}

			// Check rate limit.

			if !globalRateLimiter.Allow() {
				log.Warn().Msg("external send is blocked by the global rate limiter")
				break
			}

			// Parse email.

			rewritten, from, to, err := mail.RewriteFromHeader(content)
			if err != nil {
				log.Error().Err(err).Msg("error re-writting message")
				break
			}

			// Pipe to sendmail.

			cmd := exec.Command("sendmail", "-t", "-f", from)
			cmd.Stdin = bytes.NewReader([]byte(rewritten))
			err = cmd.Run()
			if err != nil {
				log.Error().Err(err).Msg("error running sendmail")
			}

			log.Debug().Str("to", to).Msg("external mail sent")
		}
		time.Sleep(args.PollInterval)
	}
}
