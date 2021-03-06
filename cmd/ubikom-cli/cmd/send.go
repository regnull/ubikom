package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	sendCmd.PersistentFlags().String("lookup-service-url", globals.PublicLookupServiceURL, "lookup service URL")

	sendMessageCmd.Flags().String("receiver", "", "receiver's address")
	sendMessageCmd.Flags().String("sender", "", "sender's address")
	sendMessageCmd.Flags().String("key", "", "Location for the private key file")
	sendCmd.AddCommand(sendMessageCmd)

	sendEventCmd.Flags().String("receiver", "ubikom-event-processor", "receiver's address")
	sendEventCmd.Flags().String("sender", "", "sender's address")
	sendEventCmd.Flags().String("key", "", "Location for the private key file")
	sendEventCmd.Flags().String("type", "ET_NONE", "event type")
	sendEventCmd.Flags().String("user1", "", "user1")
	sendEventCmd.Flags().String("user2", "", "user2")
	sendEventCmd.Flags().String("message", "", "message")
	sendEventCmd.Flags().String("data1", "", "data1")
	sendEventCmd.Flags().String("component", "", "component")
	sendCmd.AddCommand(sendEventCmd)
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

		privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, "")
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

		var lines []string
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your message, dot on an empty line to finish: \n")
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

		err = protoutil.SendMessage(ctx, privateKey, []byte(body), sender, receiver, lookupService)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send message")
		}
	},
}

var sendEventCmd = &cobra.Command{
	Use:   "event",
	Short: "Send event",
	Long:  "Send event",
	Run: func(cmd *cobra.Command, args []string) {
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

		privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, "")
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

		typeStr, err := cmd.Flags().GetString("type")
		if err != nil || typeStr == "" {
			log.Fatal().Err(err).Msg("receiver's address must be specified")
		}

		eventType, ok := pb.EventType_value[strings.ToUpper(typeStr)]
		if !ok {
			log.Fatal().Str("type", typeStr).Msg("invalid event type")
		}

		user1, err := cmd.Flags().GetString("user1")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get user1 flag")
		}

		user2, err := cmd.Flags().GetString("user2")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get user2 flag")
		}

		message, err := cmd.Flags().GetString("message")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get message flag")
		}

		data1, err := cmd.Flags().GetString("data1")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get data1 flag")
		}

		component, err := cmd.Flags().GetString("component")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get component flag")
		}

		event := &pb.Event{
			Id:        uuid.New().String(),
			Timestamp: uint64(util.NowMs()),
			EventType: pb.EventType(eventType),
			User1:     user1,
			User2:     user2,
			Message:   message,
			Data1:     data1,
			Component: component,
		}
		body, err := proto.Marshal(event)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to marshal message")
		}

		err = protoutil.SendMessage(ctx, privateKey, []byte(body), sender, receiver, lookupService)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send message")
		}
	},
}
