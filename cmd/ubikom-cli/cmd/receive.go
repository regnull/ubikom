package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/regnull/easyecc/v2"
	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	receiveCmd.PersistentFlags().String("network", "main", "mode, either live or prod")
	receiveCmd.PersistentFlags().String("node-url", "", "blockchain node location")
	receiveCmd.PersistentFlags().String("contract-address", "", "registry contract address")
	receiveCmd.PersistentFlags().String("dump-service-url", "", "dump service url")

	receiveMessageCmd.Flags().String("key", "", "Location of the private key file")
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

		signed, err := protoutil.IdentityProof(privateKey, time.Now())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create identity proof")
		}

		ctx := context.Background()
		client := pb.NewDMSDumpServiceClient(dumpConn)
		res, err := client.Receive(ctx, &pb.ReceiveRequest{
			IdentityProof: signed,
			CrytoContext: &pb.CryptoContext{
				EllipticCurve: protoutil.CurveToProto(privateKey.Curve()),
				EcdhVersion:   2,
				EcdsaVersion:  1,
			},
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to receive message")
		}
		msg := res.GetMessage()

		bchain, err := bc.NewBlockchain(nodeURL, contractAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create lookup service")
		}
		curve := protoutil.CurveFromProto(msg.CryptoContext.GetEllipticCurve())
		if curve == easyecc.INVALID_CURVE {
			log.Fatal().Msg("invalid curve")
		}
		senderKey, err := bchain.PublicKeyByCurve(ctx, msg.GetSender(), curve)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid sender public key")
		}

		if !protoutil.VerifySignature(msg.GetSignature(), senderKey, msg.GetContent()) {
			log.Fatal().Msg("signature verification failed")
		}

		content, err := privateKey.Decrypt(msg.Content, senderKey)
		if err != nil {
			log.Fatal().Msg("failed to decode message")
		}
		fmt.Printf("%s\n", content)
	},
}
