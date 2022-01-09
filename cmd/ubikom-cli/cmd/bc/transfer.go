package bc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	transferCmd.Flags().String("key", "", "key to authorize the transaction")
	transferCmd.Flags().String("to", "", "receiver's account")
	transferCmd.Flags().Uint64("gas-limit", 21000, "gas limit")
	transferCmd.Flags().String("value", "", "value to transfer")

	BCCmd.AddCommand(transferCmd)
}

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer funds between accounts",
	Long:  "Transfer funds between accounts",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}

		to, err := cmd.Flags().GetString("to")
		if err != nil {
			log.Fatal().Err(err).Msg("--to must be specified")
		}

		// Connect to the node.
		ctx := context.Background()
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to node")
		}
		// Get nonce.
		nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(key.PublicKey().EthereumAddress()))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get nonce")
		}
		fmt.Printf("got nonce: %d\n", nonce)

		// Recommended gas limit.
		gasLimit, err := cmd.Flags().GetUint64("gas-limit")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas limit")
		}

		// Get gas price.
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas price")
		}
		fmt.Printf("gas price: %d\n", gasPrice)

		// Send to address.
		toAddress := common.HexToAddress(to)

		// Parse value.
		value, err := cmd.Flags().GetString("value")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get value")
		}
		valueNum := new(big.Int)
		_, ok := valueNum.SetString(value, 10)
		if !ok {
			log.Fatal().Msg("failed to parse value")
		}

		// Create and sign the transaction.
		tx := types.NewTransaction(nonce, toAddress, valueNum, gasLimit, gasPrice, nil)

		chainID, err := client.NetworkID(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get chain id")
		}
		fmt.Printf("chain ID: %d\n", chainID)

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key.ToECDSA())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to sign transaction")
		}

		// Send transaction.
		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send transaction")
		}

		fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
	},
}
