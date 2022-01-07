package bc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	registerKeyCmd.Flags().String("key", "", "key to authorize the transaction")
	registerKeyCmd.Flags().String("reg-key", "", "key to register")
	registerKeyCmd.Flags().String("contract-address", globals.KeyRegistryContractAddress, "contract address")

	registerCmd.AddCommand(registerKeyCmd)

	BCCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register various things on the blockchain",
	Long:  "Register various things on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command requried (do 'ubikom-cli bc register --help' to see available commands)")
	},
}

var registerKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Register key on the blockchain",
	Long:  "Register key on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		regKey, err := LoadKeyFromFlag(cmd, "reg-key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load reg key")
		}
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		ctx := context.Background()

		// Get nonce.
		nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(key.PublicKey().EthereumAddress()))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get nonce")
		}
		log.Debug().Uint64("nonce", nonce).Msg("got nonce")

		// Recommended gas limit.
		gasLimit := uint64(300000)

		// Get gas price.
		gasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get suggested gas price")
		}
		log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

		chainID, err := client.NetworkID(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get chain ID")
		}
		log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

		auth, err := bind.NewKeyedTransactorWithChainID(key.ToECDSA(), chainID)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create transactor")
		}
		auth.Nonce = big.NewInt(int64(nonce))
		auth.Value = big.NewInt(0) // in wei
		auth.GasLimit = gasLimit
		auth.GasPrice = gasPrice

		instance, err := gocontract.NewKeyRegistry(common.HexToAddress(contractAddress), client)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get contract instance")
		}

		tx, err := instance.Register(auth, regKey.PublicKey().SerializeCompressed())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register key")
		}

		fmt.Printf("tx sent: %s\n", tx.Hash().Hex())
	},
}
