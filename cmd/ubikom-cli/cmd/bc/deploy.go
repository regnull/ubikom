package bc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubchain/gocontract"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	deployKeyRegistryCmd.Flags().String("key", "", "key to authorize the transaction")

	deployNameRegistryCmd.Flags().String("key-registry-address", "", "key registry contract address")

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

		txAddr, tx, err := deploy(nodeURL, key, func(auth *bind.TransactOpts,
			client *ethclient.Client) (common.Address, *types.Transaction, error) {
			txAddr, tx, _, err := gocontract.DeployKeyRegistry(auth, client)
			return txAddr, tx, err
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to deploy")
		}

		fmt.Printf("contract address: %s\n", txAddr.Hex())
		fmt.Printf("tx: %s\n", tx.Hash().Hex())
	},
}

var deployNameRegistryCmd = &cobra.Command{
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

		keyRegistryAddress, err := cmd.Flags().GetString("key-registry-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key registry address")
		}

		if keyRegistryAddress == "" {
			log.Fatal().Msg("--key-registry-address must be specified")
		}

		keyRegistryAddr := common.HexToAddress(keyRegistryAddress)

		txAddr, tx, err := deploy(nodeURL, key, func(auth *bind.TransactOpts,
			client *ethclient.Client) (common.Address, *types.Transaction, error) {
			txAddr, tx, _, err := gocontract.DeployNameRegistry(auth, client, keyRegistryAddr)
			return txAddr, tx, err
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to deploy")
		}

		fmt.Printf("contract address: %s\n", txAddr.Hex())
		fmt.Printf("tx: %s\n", tx.Hash().Hex())
	},
}

func deploy(nodeURL string, key *easyecc.PrivateKey,
	deployFunc func(*bind.TransactOpts,
		*ethclient.Client) (common.Address, *types.Transaction, error)) (common.Address, *types.Transaction, error) {
	// Connect to the node.
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return common.Address{}, nil, err
	}

	ctx := context.Background()

	// Get nonce.
	nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(key.PublicKey().EthereumAddress()))
	if err != nil {
		return common.Address{}, nil, err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := uint64(300000)

	// Get gas price.
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Address{}, nil, err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return common.Address{}, nil, err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(key.ToECDSA(), chainID)
	if err != nil {
		return common.Address{}, nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	return deployFunc(auth, client)
}
