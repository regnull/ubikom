package cmd

import (
	"fmt"
	"strings"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	getAddressCmd.Flags().String("key", "", "Location for the private key file")

	getEthereumAddressCmd.Flags().String("key", "", "Location for the private key file")

	getBitcoinAddressCmd.Flags().String("key", "", "Location for the private key file")

	getPublicKeyCmd.Flags().String("key", "", "Location for the private key file")

	getMnemonicCmd.Flags().String("key", "", "Location for the private key file")

	getCmd.AddCommand(getAddressCmd)
	getCmd.AddCommand(getEthereumAddressCmd)
	getCmd.AddCommand(getBitcoinAddressCmd)
	getCmd.AddCommand(getPublicKeyCmd)
	getCmd.AddCommand(getMnemonicCmd)

	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get various things",
	Long:  "Get various things",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var getEthereumAddressCmd = &cobra.Command{
	Use:   "ethereum-address",
	Short: "Get Ethereum address",
	Long:  "Get Ethereum address for the specified key",
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

		fmt.Printf("%s\n", privateKey.PublicKey().EthereumAddress())
	},
}

var getAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Get Ethereum address",
	Long:  "Get Ethereum address for the specified key",
	Run: func(cmd *cobra.Command, args []string) {
		getEthereumAddressCmd.Run(cmd, args)
	},
}

var getBitcoinAddressCmd = &cobra.Command{
	Use:   "bitcoin-address",
	Short: "Get Bitcoin address",
	Long:  "Get Bitcoin address for the specified key",
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

		fmt.Printf("%s\n", privateKey.PublicKey().Address())
	},
}

var getPublicKeyCmd = &cobra.Command{
	Use:   "public-key",
	Short: "Get public key",
	Long:  "Get public key",
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

		fmt.Printf("0x%0x\n", privateKey.PublicKey().SerializeCompressed())
	},
}

var getMnemonicCmd = &cobra.Command{
	Use:   "mnemonic",
	Short: "Get key mnemonic",
	Long:  "Get private key mnemonic",
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

		mnemonic, err := privateKey.Mnemonic()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key mnemonic")
		}
		words := strings.Split(mnemonic, " ")
		for i, word := range words {
			fmt.Printf("%d: \t%s\n", i+1, word)
		}
	},
}
