package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pow"
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

		// TODO: Define flag for server URL.
		conn, err := grpc.Dial(url, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the server")
		}
		defer conn.Close()

		powStrength, err := cmd.Flags().GetInt("pow-strength")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get POW strength")
		}

		compressedKey := privateKey.PublicKey().SerializeCompressed()
		log.Info().Msg("generating POW...")
		pow := pow.Compute(compressedKey, powStrength)
		log.Info().Hex("pow", pow).Msg("POW found")

		hash := util.Hash256(compressedKey)
		sig, err := privateKey.Sign(hash)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to generate signature")
		}

		client := pb.NewIdentityServiceClient(conn)

		req := &pb.SignedWithPow{
			Content: compressedKey,
			Pow:     pow,
			Signature: &pb.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
			Key: compressedKey,
		}

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
