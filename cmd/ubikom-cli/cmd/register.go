package cmd

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/globals"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/protoutil"
	"teralyt.com/ubikom/util"
)

const (
	defaultPowStrength = 23
)

func init() {
	registerCmd.PersistentFlags().Int("pow-strength", defaultPowStrength, "POW strength")
	registerCmd.PersistentFlags().String("url", globals.PublicIdentityServiceURL, "server URL")
	registerCmd.PersistentFlags().String("key", "", "Location for the private key file")

	registerAddressCmd.Flags().String("protocol", "PL_DMS", "protocol")

	registerCmd.AddCommand(registerKeyCmd)
	registerCmd.AddCommand(registerNameCmd)
	registerCmd.AddCommand(registerAddressCmd)
	rootCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register various things",
	Long:  "Register various things",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var registerKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Register public key",
	Long:  "Register public key",
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
		opts := []grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(time.Second * 5),
		}

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get server URL")
		}

		conn, err := grpc.Dial(url, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the server")
		}
		defer conn.Close()

		powStrength, err := cmd.Flags().GetInt("pow-strength")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get POW strength")
		}

		registerKeyReq := &pb.KeyRegistrationRequest{
			Key: privateKey.PublicKey().SerializeCompressed()}
		reqBytes, err := proto.Marshal(registerKeyReq)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to marshal request")
		}

		req, err := protoutil.CreateSignedWithPOW(privateKey, reqBytes, powStrength)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create signed request")
		}

		client := pb.NewIdentityServiceClient(conn)
		ctx := context.Background()
		res, err := client.RegisterKey(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register key")
		}
		if res.Result != pb.ResultCode_RC_OK {
			log.Fatal().Str("code", res.GetResult().Enum().String()).Msg("server returned error")
		}
		log.Info().Msg("key registered successfully")
	},
}

var registerNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Register name",
	Long:  "Register name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal().Msg("name must be specified")
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

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get server URL")
		}

		conn, err := grpc.Dial(url, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the server")
		}
		defer conn.Close()

		powStrength, err := cmd.Flags().GetInt("pow-strength")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get POW strength")
		}

		registerKeyReq := &pb.NameRegistrationRequest{
			Name: args[0]}
		reqBytes, err := proto.Marshal(registerKeyReq)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to marshal request")
		}

		req, err := protoutil.CreateSignedWithPOW(privateKey, reqBytes, powStrength)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create signed request")
		}

		client := pb.NewIdentityServiceClient(conn)
		ctx := context.Background()
		res, err := client.RegisterName(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register key")
		}
		if res.Result != pb.ResultCode_RC_OK {
			log.Fatal().Str("code", res.GetResult().Enum().String()).Msg("server returned error")
		}
		log.Info().Msg("name registered successfully")
	},
}

var registerAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Register address",
	Long:  "Register address",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Fatal().Msg("address must be specified")
		}

		name := args[0]
		address := args[1]

		protocolStr, err := cmd.Flags().GetString("protocol")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get protocol")
		}
		protocol, ok := pb.Protocol_value[protocolStr]
		if !ok {
			log.Fatal().Str("protocol", protocolStr).Msg("invalid protocol")
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

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get server URL")
		}

		conn, err := grpc.Dial(url, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the server")
		}
		defer conn.Close()

		powStrength, err := cmd.Flags().GetInt("pow-strength")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get POW strength")
		}

		registerAddressReq := &pb.AddressRegistrationRequest{
			Name:     name,
			Protocol: pb.Protocol(protocol),
			Address:  address}

		reqBytes, err := proto.Marshal(registerAddressReq)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to marshal request")
		}

		req, err := protoutil.CreateSignedWithPOW(privateKey, reqBytes, powStrength)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create signed request")
		}

		client := pb.NewIdentityServiceClient(conn)
		ctx := context.Background()
		res, err := client.RegisterAddress(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register address")
		}
		if res.Result != pb.ResultCode_RC_OK {
			log.Fatal().Str("code", res.GetResult().Enum().String()).Msg("server returned error")
		}
		log.Info().Msg("address registered successfully")
	},
}
