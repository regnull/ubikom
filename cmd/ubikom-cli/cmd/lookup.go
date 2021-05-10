package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	lookupCmd.PersistentFlags().String("url", globals.PublicLookupServiceURL, "server URL")

	lookupAddressCmd.Flags().String("protocol", "PL_DMS", "protocol")
	lookupKeyCmd.Flags().String("key", "", "Location for the private key file")

	lookupCmd.AddCommand(lookupKeyCmd)
	lookupCmd.AddCommand(lookupNameCmd)
	lookupCmd.AddCommand(lookupAddressCmd)
	rootCmd.AddCommand(lookupCmd)
}

var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Look stuff up",
	Long:  "Look stuff up",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var lookupKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Lookup key",
	Long:  "Lookup key",
	Run: func(cmd *cobra.Command, args []string) {
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

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get server URL")
		}

		opts := []grpc.DialOption{
			grpc.WithInsecure(),
		}
		conn, err := grpc.Dial(url, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the server")
		}
		defer conn.Close()

		req := &pb.LookupKeyRequest{
			Key: privateKey.PublicKey().SerializeCompressed()}

		client := pb.NewLookupServiceClient(conn)
		ctx := context.Background()
		res, err := client.LookupKey(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("key lookup request failed")
		}
		if res.GetResult() != pb.ResultCode_RC_OK {
			log.Fatal().Str("result", res.GetResult().String()).Msg("key lookup request failed")
		}

		if res.GetDisabled() {
			fmt.Printf("THIS KEY IS DISABLED\n")
			disabledTs := util.TimeFromMs(res.GetDisabledTimestamp())
			fmt.Printf("disabled timestamp: %s\n", disabledTs.Format(time.RFC822))
			fmt.Printf("disabled by: %x\n", res.GetDisabledBy())
		}
		ts := util.TimeFromMs(res.GetRegistrationTimestamp())
		fmt.Printf("registration timestamp: %s\n", ts.Format(time.RFC822))

		for _, parentKey := range res.GetParentKey() {
			fmt.Printf("parent key: %x\n", parentKey)
		}
	},
}

var lookupNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Lookup name",
	Long:  "Lookup name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get server URL")
		}

		opts := []grpc.DialOption{
			grpc.WithInsecure(),
		}
		conn, err := grpc.Dial(url, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the server")
		}
		defer conn.Close()

		req := &pb.LookupNameRequest{
			Name: args[0]}

		client := pb.NewLookupServiceClient(conn)
		ctx := context.Background()
		res, err := client.LookupName(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("name lookup request failed")
		}
		fmt.Printf("%s\n", base58.Encode(res.GetKey()))
	},
}

var lookupAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Lookup address",
	Long:  "Lookup address",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get server URL")
		}

		protocolStr, err := cmd.Flags().GetString("protocol")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get protocol")
		}
		protocol, ok := pb.Protocol_value[protocolStr]
		if !ok {
			log.Fatal().Str("protocol", protocolStr).Msg("invalid protocol")
		}

		opts := []grpc.DialOption{
			grpc.WithInsecure(),
		}
		conn, err := grpc.Dial(url, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the server")
		}
		defer conn.Close()

		req := &pb.LookupAddressRequest{
			Name:     args[0],
			Protocol: pb.Protocol(protocol)}

		client := pb.NewLookupServiceClient(conn)
		ctx := context.Background()
		res, err := client.LookupAddress(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("address lookup request failed")
		}
		fmt.Printf("%s\n", res.GetAddress())
	},
}
