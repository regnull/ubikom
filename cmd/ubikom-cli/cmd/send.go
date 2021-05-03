package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/util"
)

func init() {
	sendCmd.PersistentFlags().String("url", "localhost:8826", "server URL")
	sendMessageCmd.Flags().String("receiver", "", "receiver's address")
	sendMessageCmd.Flags().String("sender", "", "sender's address")
	sendMessageCmd.Flags().String("key", "", "Location for the private key file")
	// TODO: Add key location flag.
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
		url, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get server URL")
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
		}
		conn, err := grpc.Dial(url, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the server")
		}
		defer conn.Close()

		sender, err := cmd.Flags().GetString("sender")
		if err != nil || sender == "" {
			log.Fatal().Err(err).Msg("sender's address must be specified")
		}

		receiver, err := cmd.Flags().GetString("receiver")
		if err != nil || receiver == "" {
			log.Fatal().Err(err).Msg("receiver's address must be specified")
		}

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

		hash := util.Hash256([]byte(body))
		sig, err := privateKey.Sign(hash)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to sign message")
		}

		// TODO: Encrypt the message.

		msg := &pb.DMSMessage{
			Sender:   sender,
			Receiver: receiver,
			Content:  []byte(body),
			Signature: &pb.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
		}

		client := pb.NewDMSDumpServiceClient(conn)
		ctx := context.Background()
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
