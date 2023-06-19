package cmd

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	cnt "github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	registerNameCmd.Flags().String("key", "", "key to authorize the transaction")
	registerNameCmd.Flags().String("enc-key", "", "encryption key")

	registerCmd.AddCommand(registerNameCmd)

	rootCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register various things on the blockchain",
	Long:  "Register various things on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command required (do 'ubikom-cli bc register --help' to see available commands)")
	},
}

var registerNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Register name on the blockchain",
	Long:  "Register name on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
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
			encKey, err = cmdutil.LoadKeyFromFlag(cmd, "enc-key")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load encryption key")
			}
		}
		gasPrice, err := cmd.Flags().GetUint64("gas-price")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas-price flag")
		}
		gasLimit, err := cmd.Flags().GetUint64("gas-limit")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas-limit flag")
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
		err = interactWithContract(nodeURL, key, contractAddress, 0, gasPrice, gasLimit,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cnt.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.RegisterName(auth, encKey.PublicKey().SerializeCompressed(), name)
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
