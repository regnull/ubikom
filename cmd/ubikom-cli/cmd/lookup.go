package cmd

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/pb"
)

func init() {
	lookupCmd.PersistentFlags().String("url", "localhost:8825", "server URL")

	lookupAddressCmd.Flags().String("protocol", "PL_DMS", "protocol")

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
