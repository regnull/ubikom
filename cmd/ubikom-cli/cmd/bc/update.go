package bc

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	cntv2 "github.com/regnull/ubchain/gocontract/v2"
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	updatePublicKeyCmd.Flags().String("key", "", "key to authorize the transaction")
	updatePublicKeyCmd.Flags().String("pub-key", "", "public key to update")
	updatePublicKeyCmd.Flags().String("contract-address", globals.NameRegistryContractAddress, "contract address")

	updateCmd.AddCommand(updatePublicKeyCmd)

	BCCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update various things on the blockchain",
	Long:  "Update various things on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command required (do 'ubikom-cli bc update --help' to see available commands)")
	},
}

var updatePublicKeyCmd = &cobra.Command{
	Use:   "public-key",
	Short: "Update public key on the blockchain",
	Long:  "Update public key on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		encKey, err := LoadKeyFromFlag(cmd, "pub-key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load reg key")
		}
		if len(args) < 1 {
			log.Fatal().Msg("name must be specified")
		}

		name := args[0]
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		err = interactWithContract(nodeURL, key, contractAddress,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cntv2.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.UpdatePublicKey(auth, encKey.PublicKey().SerializeCompressed(), name)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to register name")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register name")
		}
	},
}
