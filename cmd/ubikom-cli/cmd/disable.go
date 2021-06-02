package cmd

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
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
	disableCmd.PersistentFlags().String("url", globals.PublicLookupServiceURL, "server URL")
	disableCmd.PersistentFlags().Int("pow-strength", defaultPowStrength, "POW strength")
	disableKeyCmd.Flags().String("key", "", "Location for the private key file")
	disableKeyCmd.Flags().Bool("confirm", false, "confirm operation")
	disableCmd.AddCommand(disableKeyCmd)
	rootCmd.AddCommand(disableCmd)
}

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable something",
	Long:  "Disable something",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var disableKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Disable key",
	Long:  "Disable key",
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

		confirm, err := cmd.Flags().GetBool("confirm")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get confirm flag")
		}
		if !confirm {
			log.Fatal().Msg("!!! DISABLING KEY IS NOT REVERSIBLE !!! Once this key is disabled, it's gone forever. " +
				"If you really want to do that, re-issue the command with --confirm flag.")
		}

		privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, "")
		if err != nil {
			log.Fatal().Err(err).Str("location", keyFile).Msg("cannot load private key")
		}

		powStrength, err := cmd.Flags().GetInt("pow-strength")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get POW strength")
		}

		registerKeyReq := &pb.KeyDisableRequest{
			Key: privateKey.PublicKey().SerializeCompressed()}
		reqBytes, err := proto.Marshal(registerKeyReq)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to marshal request")
		}

		req, err := protoutil.CreateSignedWithPOW(privateKey, reqBytes, powStrength)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create signed request")
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
		client := pb.NewIdentityServiceClient(conn)
		ctx := context.Background()
		_, err = client.DisableKey(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to disable key")
		}
		log.Info().Msg("key is disabled")
	},
}
