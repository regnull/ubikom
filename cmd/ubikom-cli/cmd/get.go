package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/util"
)

func init() {
	getAddressCmd.Flags().String("key", "", "Location for the private key file")
	getCmd.AddCommand(getAddressCmd)
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get various things",
	Long:  "Get various things",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var getAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Get address",
	Long:  "Get address",
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

		fmt.Printf("%s\n", privateKey.PublicKey().Address())
	},
}
