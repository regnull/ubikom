package cmd

import (
	"context"
	"fmt"
	"math/big"

	"encoding/json"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	gocontv2 "github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	deployRegistryCmd.Flags().String("key", "", "key to authorize the transaction")

	deployCmd.PersistentFlags().Uint64("gas-limit", 2000000, "gas limit")

	deployCmd.AddCommand(deployRegistryCmd)

	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy contract on a blockchain",
	Long:  "Deploy contract on a blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command requried (do 'ubikom-cli bc deploy --help' to see them)")
	},
}

type registryDeployResult struct {
	Address string
	Tx      string
	Block   string
}

var deployRegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Deploy registry v2",
	Long:  "Deploy registry v2",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := cmdutil.LoadKeyFromFlag(cmd, "key")
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

		txAddr, tx, block, err := deploy(nodeURL, key, gasLimit, func(auth *bind.TransactOpts,
			client *ethclient.Client) (common.Address, *types.Transaction, error) {
			txAddr, tx, _, err := gocontv2.DeployNameRegistry(auth, client)
			return txAddr, tx, err
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to deploy")
		}

		res := &registryDeployResult{
			Address: txAddr.Hex(),
			Tx:      tx.Hash().Hex(),
			Block:   block.String(),
		}

		s, err := json.MarshalIndent(res, "", "  ")

		fmt.Printf("%s\n", s)
	},
}

func deploy(nodeURL string, key *easyecc.PrivateKey, gasLimit uint64,
	deployFunc func(*bind.TransactOpts,
		*ethclient.Client) (common.Address, *types.Transaction, error)) (common.Address,
	*types.Transaction,
	*big.Int,
	error) {
	// Connect to the node.
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	ctx := context.Background()

	// Get nonce.
	etheriumAddress, _ := key.PublicKey().EthereumAddress()
	nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(etheriumAddress))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Get gas price.
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(key.ToECDSA(), chainID)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	txAddr, tx, err := deployFunc(auth, client)
	if err != nil {
		return txAddr, tx, nil, err
	}
	receipt, err := bind.WaitMined(ctx, client, tx)
	if receipt.Status == 0 {
		return txAddr, tx, nil, fmt.Errorf("miner returned error %w", err)
	}

	return txAddr, tx, receipt.BlockNumber, nil
}
