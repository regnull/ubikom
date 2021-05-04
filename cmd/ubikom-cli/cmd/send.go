package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/util"
)

func init() {
	sendCmd.PersistentFlags().String("dump-service-url", "localhost:8826", "dump service URL")
	sendCmd.PersistentFlags().String("lookup-service-url", "localhost:8825", "lookup service URL")
	sendMessageCmd.Flags().String("receiver", "", "receiver's address")
	sendMessageCmd.Flags().String("sender", "", "sender's address")
	sendMessageCmd.Flags().String("key", "", "Location for the private key file")
	sendCmd.AddCommand(sendMessageCmd)
	rootCmd.AddCommand(sendCmd)
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send stuff",
	Long:  "Send stuff",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var sendMessageCmd = &cobra.Command{
	Use:   "message",
	Short: "Send message",
	Long:  "Send message",
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

		lookupConn, err := grpc.Dial(lookupServiceURL, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the lookup server")
		}
		defer lookupConn.Close()

		sender, err := cmd.Flags().GetString("sender")
		if err != nil || sender == "" {
			log.Fatal().Err(err).Msg("sender's address must be specified")
		}

		receiver, err := cmd.Flags().GetString("receiver")
		if err != nil || receiver == "" {
			log.Fatal().Err(err).Msg("receiver's address must be specified")
		}

		ctx := context.Background()

		lookupService := pb.NewLookupServiceClient(lookupConn)
		lookupRes, err := lookupService.LookupName(ctx, &pb.LookupNameRequest{Name: receiver})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get receiver public key")
		}
		if lookupRes.Result != pb.ResultCode_RC_OK {
			log.Fatal().Str("result", lookupRes.Result.String()).Msg("failed to get receiver public key")
		}
		receiverKey, err := ecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
		if err != nil {
			log.Fatal().Err(err).Msg("invalid receiver public key")
		}

		dumpConn, err := grpc.Dial(dumpURL, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the dump server")
		}
		defer dumpConn.Close()

		var lines []string
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter message: ")
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if text == ".\n" {
				break
			}
			lines = append(lines, text)
		}
		body := strings.Join(lines, "\n")

		encryptedBody, err := privateKey.Encrypt([]byte(body), receiverKey)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to encrypt message")
		}

		hash := util.Hash256(encryptedBody)
		sig, err := privateKey.Sign(hash)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to sign message")
		}

		msg := &pb.DMSMessage{
			Sender:   sender,
			Receiver: receiver,
			Content:  encryptedBody,
			Signature: &pb.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
		}

		client := pb.NewDMSDumpServiceClient(dumpConn)
		res, err := client.Send(ctx, msg)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send message")
		}
		if res.Result != pb.ResultCode_RC_OK {
			log.Fatal().Str("code", res.GetResult().Enum().String()).Msg("server returned error")
		}
		log.Info().Msg("sent message successfully")
	},
}
