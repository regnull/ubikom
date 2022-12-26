package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func init() {
	receiveCmd.PersistentFlags().String("network", "main", "mode, either live or prod")
	receiveCmd.PersistentFlags().String("node-url", "", "blockchain node location")
	receiveCmd.PersistentFlags().String("contract-address", "", "registry contract address")
	receiveCmd.PersistentFlags().String("dump-service-url", "", "dump service url")

	receiveMessageCmd.Flags().String("key", "", "Location of the private key file")
	receiveCmd.AddCommand(receiveMessageCmd)

	receiveEventCmd.Flags().String("key", "", "Location of the private key file")
	receiveCmd.AddCommand(receiveEventCmd)

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
		if dumpURL == "" {
			log.Fatal().Msg("--dump-service-url must be specified")
		}

		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		log.Debug().Str("node-url", nodeURL).Msg("using node")
		contractAddress, err := cmdutil.GetContractAddress(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		log.Debug().Str("contract-address", contractAddress).Msg("using contract addresss")

		privateKey, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load encryption key")
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

		signed := &pb.Signed{
			Content: []byte("we need a bigger boat"),
			Signature: &pb.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
			Key: privateKey.PublicKey().SerializeCompressed(),
		}

		ctx := context.Background()
		client := pb.NewDMSDumpServiceClient(dumpConn)
		res, err := client.Receive(ctx, &pb.ReceiveRequest{IdentityProof: signed})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send message")
		}
		msg := res.GetMessage()

		lookupService, err := bc.NewBlockchainV2(nodeURL, contractAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create lookup service")
		}
		lookupRes, err := lookupService.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get receiver public key")
		}
		senderKey, err := easyecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
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

var receiveEventCmd = &cobra.Command{
	Use:   "event",
	Short: "Receive event",
	Long:  "Receive event",
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

		encrypted, err := util.IsKeyEncrypted(keyFile)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot find key file")
		}

		passphrase := ""
		if encrypted {
			passphrase, err = util.ReadPassphase()
			if err != nil {
				log.Fatal().Err(err).Msg("cannot read passphrase")
			}
		}

		privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, passphrase)
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

		signed := &pb.Signed{
			Content: []byte("we need a bigger boat"),
			Signature: &pb.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
			Key: privateKey.PublicKey().SerializeCompressed(),
		}

		ctx := context.Background()
		client := pb.NewDMSDumpServiceClient(dumpConn)
		lookupConn, err := grpc.Dial(lookupServiceURL, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the lookup server")
		}
		defer lookupConn.Close()
		lookupService := pb.NewLookupServiceClient(lookupConn)

		for {
			res, err := client.Receive(ctx, &pb.ReceiveRequest{IdentityProof: signed})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to receive message")
			}
			msg := res.GetMessage()

			lookupRes, err := lookupService.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get sender public key")
			}
			senderKey, err := easyecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
			if err != nil {
				log.Fatal().Err(err).Msg("invalid sender public key")
			}

			if !protoutil.VerifySignature(msg.GetSignature(), lookupRes.GetKey(), msg.GetContent()) {
				log.Fatal().Msg("signature verification failed")
			}

			content, err := privateKey.Decrypt(msg.Content, senderKey)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to decrypt message")
			}

			event := &pb.Event{}
			err = proto.Unmarshal(content, event)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to unmarshal event")
			}

			marshalOpts := protojson.MarshalOptions{
				Multiline: true,
				Indent:    "  ",
			}
			json, err := marshalOpts.Marshal(event)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to marshal to JSON")
			}

			fmt.Printf("%s\n", json)
		}
	},
}
