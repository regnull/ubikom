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

	deployNameRegistryCmd.Flags().String("key", "", "key to authorize the transaction")
	deployNameRegistryCmd.Flags().String("key-registry-address", "", "key registry contract address")

	deployConnectorRegistryCmd.Flags().String("key", "", "key to authorize the transaction")
	deployConnectorRegistryCmd.Flags().String("key-registry-address", "", "key registry contract address")
	deployConnectorRegistryCmd.Flags().String("name-registry-address", "", "name registry contract address")

	deployCmd.PersistentFlags().Uint64("gas-limit", 1000000, "gas limit")

	deployCmd.AddCommand(deployKeyRegistryCmd)
	deployCmd.AddCommand(deployNameRegistryCmd)
	deployCmd.AddCommand(deployConnectorRegistryCmd)

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

		gasLimit, err := cmd.Flags().GetUint64("gas-limit")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas limit")
		}

		txAddr, tx, err := deploy(nodeURL, key, gasLimit, func(auth *bind.TransactOpts,
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
	Use:   "name-registry",
	Short: "Deploy name registry",
	Long:  "Deploy name registry",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}

		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		gasLimit, err := cmd.Flags().GetUint64("gas-limit")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas limit")
		}

		keyRegistryAddress, err := cmd.Flags().GetString("key-registry-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key registry address")
		}

		if keyRegistryAddress == "" {
			log.Fatal().Msg("--key-registry-address must be specified")
		}

		keyRegistryAddr := common.HexToAddress(keyRegistryAddress)

		txAddr, tx, err := deploy(nodeURL, key, gasLimit, func(auth *bind.TransactOpts,
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

var deployConnectorRegistryCmd = &cobra.Command{
	Use:   "connector-registry",
	Short: "Deploy connector registry",
	Long:  "Deploy connector registry",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}

		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		gasLimit, err := cmd.Flags().GetUint64("gas-limit")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas limit")
		}

		keyRegistryAddress, err := cmd.Flags().GetString("key-registry-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key registry address")
		}

		if keyRegistryAddress == "" {
			log.Fatal().Msg("--key-registry-address must be specified")
		}

		nameRegistryAddress, err := cmd.Flags().GetString("name-registry-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get name registry address")
		}

		if nameRegistryAddress == "" {
			log.Fatal().Msg("--name-registry-address must be specified")
		}

		keyRegistryAddr := common.HexToAddress(keyRegistryAddress)
		nameRegistryAddr := common.HexToAddress(nameRegistryAddress)

		txAddr, tx, err := deploy(nodeURL, key, gasLimit, func(auth *bind.TransactOpts,
			client *ethclient.Client) (common.Address, *types.Transaction, error) {
			txAddr, tx, _, err := gocontract.DeployConnectorRegistry(auth, client, keyRegistryAddr, nameRegistryAddr)
			return txAddr, tx, err
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to deploy")
		}

		fmt.Printf("contract address: %s\n", txAddr.Hex())
		fmt.Printf("tx: %s\n", tx.Hash().Hex())
	},
}

func deploy(nodeURL string, key *easyecc.PrivateKey, gasLimit uint64,
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
