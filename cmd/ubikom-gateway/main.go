package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type CmdArgs struct {
	KeyLocation           string
	DumpURL               string
	LookupURL             string
	ConnectionTimeoutMsec int
}

func main() {
	var args CmdArgs
	flag.StringVar(&args.KeyLocation, "key", "", "key location")
	flag.StringVar(&args.DumpURL, "dump-url", "locahost:8826", "dump service URL")
	flag.StringVar(&args.LookupURL, "lookup-url", "localhost:8825", "lookup service URL")
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
				// No more messages.
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
			}
			fmt.Printf("%s\n\n", content)
		}
		time.Sleep(60 * time.Second)
	}
}
