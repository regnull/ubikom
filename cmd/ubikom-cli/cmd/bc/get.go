package bc

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	getCmd.AddCommand(getBalanceCmd)
	getCmd.AddCommand(getBlockCmd)
	getCmd.AddCommand(getReceiptCmd)

	BCCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get various things on the blockchain",
	Long:  "Get various things on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command requried (do 'ubikom-cli bc get --help' to see them)")
	},
}

var getBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Get balance",
	Long:  "Get balance",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := getNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		if len(args) < 1 {
			log.Fatal().Msg("address is required")
		}
		address := args[0]

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		// Get balance.
		ctx := context.Background()
		balance, err := client.BalanceAt(ctx, common.HexToAddress(address), nil)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get balance")
		}
		fmt.Printf("Balance: %d\n", balance)
	},
}

var getBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Get latest block number",
	Long:  "Get latest block number",
	Run: func(cmd *cobra.Command, args []string) {
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
		num, err := client.BlockNumber(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get latest block number")
		}
		fmt.Printf("%d\n", num)
	},
}

var getReceiptCmd = &cobra.Command{
	Use:   "receipt",
	Short: "Get transaction receipt",
	Long:  "Get transaction receipt",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		if len(args) < 1 {
			log.Fatal().Msg("transaction hash required")
		}

		tx := args[0]
		if len(tx) < 10 {
			log.Fatal().Msg("transaction hash required")
		}

		if strings.HasPrefix("0x", tx) {
			tx = tx[2:]
		}

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		ctx := context.Background()

		hash, err := hex.DecodeString(tx)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid transaction")
		}
		receipt, err := client.TransactionReceipt(ctx, common.BytesToHash(hash))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get transaction receipt")
		}

		fmt.Printf("status: %d\n", receipt.Status)
		fmt.Printf("block hash: %s\n", receipt.BlockHash.Hex())
		fmt.Printf("block number: %d\n", receipt.BlockNumber)
		fmt.Printf("gas used: %d\n", receipt.GasUsed)
		fmt.Printf("tx index: %d\n", receipt.TransactionIndex)
		fmt.Printf("tx type: %d\n", receipt.Type)
	},
}
