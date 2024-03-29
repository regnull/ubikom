package cmd

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc/v2"
	cnt "github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	buyNameCmd.Flags().String("key", "", "key to authorize the transaction")
	buyNameCmd.Flags().String("enc-key", "", "encryption key")
	buyNameCmd.Flags().Int64("value", 0, "value")

	buyCmd.AddCommand(buyNameCmd)

	rootCmd.AddCommand(buyCmd)
}

var buyCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy",
	Long:  "Buy",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command required (do 'ubikom-cli bc update --help' to see available commands)")
	},
}

var buyNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Buy name",
	Long:  "Buy name",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		log.Debug().Str("node-url", nodeURL).Msg("using node")
		contractAddress, err := cmdutil.GetContractAddress(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		log.Debug().Str("contract-address", contractAddress).Msg("using contract address")

		key, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		if len(args) < 1 {
			log.Fatal().Msg("name must be specified")
		}
		name := args[0]
		encKeyPath, err := cmd.Flags().GetString("enc-key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load public key")
		}
		var encKey *easyecc.PrivateKey
		if encKeyPath == "" {
			encKey, err = easyecc.NewPrivateKey(easyecc.SECP256K1)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate public key")
			}
			entheriumAddress, _ := encKey.PublicKey().EthereumAddress()
			log.Info().Str("key", entheriumAddress).Msg("generated new encryption key")
		} else {
			encKey, err = cmdutil.LoadKeyFromFlag(cmd, "enc-key")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load encryption key")
			}
		}
		value, err := cmd.Flags().GetInt64("value")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get value")
		}
		err = interactWithContract(nodeURL, key, contractAddress, value, 0, 0,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cnt.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.BuyName(auth, name, encKey.PublicKey().CompressedBytes())
				if err != nil {
					log.Fatal().Err(err).Msg("failed to buy name")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to buy name")
		}
	},
}
