package cmd

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	cnt "github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	updatePublicKeyCmd.Flags().String("key", "", "key to authorize the transaction")
	updatePublicKeyCmd.Flags().String("pub-key", "", "public key to update")

	updateOwnerCmd.Flags().String("key", "", "key to authorize the transaction")
	updateOwnerCmd.Flags().String("new-owner-address", "", "new owner address")

	updatePriceCmd.Flags().String("key", "", "key to authorize the transaction")
	updatePriceCmd.Flags().Int64("price", 0, "new price")

	updateConfigCmd.Flags().String("key", "", "key to authorize the transaction")
	updateConfigCmd.Flags().String("config-name", "", "new price")
	updateConfigCmd.Flags().String("config-value", "", "new price")

	updateCmd.AddCommand(updatePublicKeyCmd)
	updateCmd.AddCommand(updateOwnerCmd)
	updateCmd.AddCommand(updatePriceCmd)
	updateCmd.AddCommand(updateConfigCmd)

	rootCmd.AddCommand(updateCmd)
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
		key, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		encKey, err := cmdutil.LoadKeyFromFlag(cmd, "pub-key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load reg key")
		}
		if len(args) < 1 {
			log.Fatal().Msg("name must be specified")
		}

		name := args[0]
		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		log.Debug().Str("node-url", nodeURL).Msg("using node")
		contractAddress, err := cmdutil.GetContractAddress(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		log.Debug().Str("contract-address", contractAddress).Msg("using contract")
		err = interactWithContract(nodeURL, key, contractAddress, 0, 0, 0,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cnt.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.UpdatePublicKey(auth, encKey.PublicKey().SerializeCompressed(), name)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to update public key")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to update public key")
		}
	},
}

var updateOwnerCmd = &cobra.Command{
	Use:   "owner",
	Short: "Update name owner on the blockchain",
	Long:  "Update name owner on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		if len(args) < 1 {
			log.Fatal().Msg("name must be specified")
		}
		name := args[0]

		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		newOwnerAddressHex, err := cmd.Flags().GetString("new-owner-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get new owner address")
		}
		newOwnerAddress := common.HexToAddress(newOwnerAddressHex)
		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		err = interactWithContract(nodeURL, key, contractAddress, 0, 0, 0,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cnt.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.UpdateOwnership(auth, name, newOwnerAddress)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to update owner")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to update owner")
		}
	},
}

var updatePriceCmd = &cobra.Command{
	Use:   "price",
	Short: "Update name price on the blockchain",
	Long:  "Update name price on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		if len(args) < 1 {
			log.Fatal().Msg("name must be specified")
		}
		name := args[0]

		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		log.Debug().Str("node-url", nodeURL).Msg("using node")
		contractAddress, err := cmdutil.GetContractAddress(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		log.Debug().Str("contract-address", contractAddress).Msg("using contract")
		price, err := cmd.Flags().GetInt64("price")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get new owner address")
		}
		err = interactWithContract(nodeURL, key, contractAddress, 0, 0, 0,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cnt.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.UpdatePrice(auth, name, big.NewInt(price))
				if err != nil {
					log.Fatal().Err(err).Msg("failed to update price")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to update price")
		}
	},
}

var updateConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Update name config on the blockchain",
	Long:  "Update name config on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		if len(args) < 1 {
			log.Fatal().Msg("name must be specified")
		}
		name := args[0]

		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		log.Debug().Str("node-url", nodeURL).Msg("using node")
		contractAddress, err := cmdutil.GetContractAddress(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		log.Debug().Str("contract-address", contractAddress).Msg("using contract")
		gasPrice, err := cmd.Flags().GetUint64("gas-price")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas-price flag")
		}
		gasLimit, err := cmd.Flags().GetUint64("gas-limit")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas-limit flag")
		}

		configName, err := cmd.Flags().GetString("config-name")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get config name")
		}
		configValue, err := cmd.Flags().GetString("config-value")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get config value")
		}
		if configName == "" {
			log.Fatal().Msg("--config-name cannot be empty")
		}
		err = interactWithContract(nodeURL, key, contractAddress, 0, gasPrice, gasLimit,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cnt.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.UpdateConfig(auth, name, configName, configValue)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to update config")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to update config")
		}
	},
}
