package bc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/ubchain/keyregistry"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	deployKeyRegistryCmd.Flags().String("key", "", "key to authorize the transaction")

	deployCmd.AddCommand(deployKeyRegistryCmd)

	BCCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy contract on a blockchain",
	Long:  "Deploy contract on a blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command requried (do 'ubikom-cli bc deploy --help' to see them)")
	},
}

var deployKeyRegistryCmd = &cobra.Command{
	Use:   "key-registry",
	Short: "Deploy key registry",
	Long:  "Deploy key registry",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}

		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
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

		txAddr, tx, _, err := keyregistry.DeployKeyregistry(auth, client)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to deploy")
		}

		fmt.Printf("contract address: %s\n", txAddr.Hex())
		fmt.Printf("tx: %s\n", tx.Hash().Hex())
	},
}
