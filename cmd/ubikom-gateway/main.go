package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/mail"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type CmdArgs struct {
	KeyLocation            string
	DumpURL                string
	LookupURL              string
	ConnectionTimeoutMsec  int
	GlobalRateLimitPerHour int
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
		content := "we will need a bigger boat"
		hash := util.Hash256([]byte(content))

		sig, err := privateKey.Sign(hash)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to sign message")
		}

		req := &pb.Signed{
			Content: []byte(content),
			Signature: &pb.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
			Key: privateKey.PublicKey().SerializeCompressed(),
		}
		for {
			res, err := dumpClient.Receive(ctx, req)
			if err != nil {
				log.Error().Err(err).Msg("failed to receive message")
				break
			}
			if res.GetResult().GetResult() == pb.ResultCode_RC_RECORD_NOT_FOUND {
				log.Info().Msg("no new messages")
				break
			}
			if res.Result.Result != pb.ResultCode_RC_OK {
				log.Error().Str("result", res.GetResult().GetResult().String()).Msg("server returned error")
				break
			}
			msg := &pb.DMSMessage{}
			err = proto.Unmarshal(res.GetContent(), msg)
			if err != nil {
				log.Error().Err(err).Msg("failed to unmarshal message")
				break
			}

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

			contentReader := strings.NewReader(content)
			mailMsg, err := mail.ReadMessage(contentReader)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse email message")
				break
			}

			// Process headers.

			from := mailMsg.Header.Get("From")
			address, err := mail.ParseAddress(from)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse address")
				break
			}

			addr := address.Address
			addr = strings.Replace(addr, "@x", "@ubikom.cc", 1)

			addrStr := ""
			if address.Name != "" {
				addrStr = fmt.Sprintf("%s <%s>", address.Name, addr)
			} else {
				addrStr = addr
			}

			var buf bytes.Buffer

			buf.Write([]byte(fmt.Sprintf("To: %s\n", mailMsg.Header.Get("To"))))
			buf.Write([]byte(fmt.Sprintf("From: %s\n", addrStr)))
			for name, values := range mailMsg.Header {
				if name == "From" {
					continue
				}
				if name == "To" {
					continue
				}
				for _, value := range values {
					buf.Write([]byte(fmt.Sprintf("%s: %s\n", name, value)))
				}
			}
			buf.Write([]byte("\n"))
			io.Copy(&buf, mailMsg.Body)

			// Pipe to sendmail.
			cmd := exec.Command("sendmail", "-t", "-f", addr)
			cmd.Stdin = bytes.NewReader(buf.Bytes())
			err = cmd.Run()
			if err != nil {
				log.Error().Err(err).Msg("error running sendmail")
			}

			log.Debug().Str("to", mailMsg.Header.Get("To")).Msg("external mail sent")
		}
		time.Sleep(60 * time.Second)
	}
}
