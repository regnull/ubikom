package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func init() {
	receiveCmd.PersistentFlags().String("dump-service-url", globals.PublicDumpServiceURL, "dump service URL")
	receiveCmd.PersistentFlags().String("lookup-service-url", globals.PublicLookupServiceURL, "lookup service URL")
	receiveMessageCmd.Flags().String("key", "", "Location for the private key file")
	receiveCmd.AddCommand(receiveMessageCmd)
	rootCmd.AddCommand(receiveCmd)
}

var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive stuff",
	Long:  "Receive stuff",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var receiveMessageCmd = &cobra.Command{
	Use:   "message",
	Short: "Receive message",
	Long:  "Receive message",
	Run: func(cmd *cobra.Command, args []string) {
		dumpURL, err := cmd.Flags().GetString("dump-service-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get dump server URL")
		}

		lookupServiceURL, err := cmd.Flags().GetString("lookup-service-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get lookup server URL")
		}

		keyFile, err := cmd.Flags().GetString("key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key location")
		}

		if keyFile == "" {
			keyFile, err = util.GetDefaultKeyLocation()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get private key")
			}
		}

		privateKey, err := ecc.LoadPrivateKey(keyFile)
		if err != nil {
			log.Fatal().Err(err).Str("location", keyFile).Msg("cannot load private key")
		}

		opts := []grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(time.Second * 5),
		}

		dumpConn, err := grpc.Dial(dumpURL, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the dump server")
		}
		defer dumpConn.Close()

		hash := util.Hash256([]byte("we need a bigger boat"))
		sig, err := privateKey.Sign(hash)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to sign message")
		}

		req := &pb.Signed{
			Content: []byte("we need a bigger boat"),
			Signature: &pb.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
			Key: privateKey.PublicKey().SerializeCompressed(),
		}

		ctx := context.Background()
		client := pb.NewDMSDumpServiceClient(dumpConn)
		res, err := client.Receive(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send message")
		}
		if res.Result.Result != pb.ResultCode_RC_OK {
			log.Fatal().Str("code", res.GetResult().GetResult().Enum().String()).Msg("server returned error")
		}
		msg := &pb.DMSMessage{}
		err = proto.Unmarshal(res.GetContent(), msg)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to unmarshal message")
		}

		lookupConn, err := grpc.Dial(lookupServiceURL, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the lookup server")
		}
		defer lookupConn.Close()
		lookupService := pb.NewLookupServiceClient(lookupConn)
		lookupRes, err := lookupService.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get receiver public key")
		}
		if lookupRes.Result != pb.ResultCode_RC_OK {
			log.Fatal().Str("result", lookupRes.Result.String()).Msg("failed to get receiver public key")
		}
		senderKey, err := ecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
		if err != nil {
			log.Fatal().Err(err).Msg("invalid receiver public key")
		}

		if !protoutil.VerifySignature(msg.GetSignature(), lookupRes.GetKey(), msg.GetContent()) {
			log.Fatal().Msg("signature verification failed")
		}

		content, err := privateKey.Decrypt(msg.Content, senderKey)
		if err != nil {
			log.Fatal().Msg("failed to decode message")
		}
		fmt.Printf("%s\n", content)
	},
}
