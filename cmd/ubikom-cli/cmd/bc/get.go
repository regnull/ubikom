package bc

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	getBalanceCmd.Flags().String("address", "", "get balance for this address")

	getCmd.AddCommand(getBalanceCmd)
	getCmd.AddCommand(getBlockCmd)

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
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		address, err := cmd.Flags().GetString("address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get address")
		}
		if address == "" {
			log.Fatal().Msg("address is required")
		}

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
