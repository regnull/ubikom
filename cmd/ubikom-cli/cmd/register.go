package cmd

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/protoutil"
	"teralyt.com/ubikom/util"
)

const (
	defaultPowStrength = 23
)

func init() {
	registerKeyCmd.Flags().Int("pow-strength", defaultPowStrength, "POW strength")
	registerKeyCmd.Flags().String("url", "localhost:8825", "server URL")
	registerKeyCmd.Flags().String("key", "", "Location for the private key file")
	registerCmd.AddCommand(registerKeyCmd)
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
