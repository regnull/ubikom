package bc

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	cntv2 "github.com/regnull/ubchain/gocontract/v2"
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	buyNameCmd.Flags().String("key", "", "key to authorize the transaction")
	buyNameCmd.Flags().String("enc-key", "", "encryption key")
	buyNameCmd.Flags().Int64("value", 0, "value")
	buyNameCmd.Flags().String("contract-address", globals.NameRegistryContractAddress, "contract address")

	buyCmd.AddCommand(buyNameCmd)

	BCCmd.AddCommand(buyCmd)
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
		key, err := LoadKeyFromFlag(cmd, "key")
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
			encKey, err = easyecc.NewRandomPrivateKey()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate public key")
			}
			log.Info().Str("key", string(encKey.PublicKey().EthereumAddress())).Msg("generated new encryption key")
		} else {
			encKey, err = LoadKeyFromFlag(cmd, "enc-key")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load encryption key")
			}
		}
		value, err := cmd.Flags().GetInt64("value")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get value")
		}
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		err = interactWithContract(nodeURL, key, contractAddress, value,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cntv2.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.BuyName(auth, name, encKey.PublicKey().SerializeCompressed())
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
