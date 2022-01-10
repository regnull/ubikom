package bc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	findTxCmd.Flags().Uint("max-blocks", 100, "maximum blocks to scan")

	findCmd.AddCommand(findTxCmd)

	BCCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find various things on the blockchain",
	Long:  "Find various things on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command requried (do 'ubikom-cli bc find --help' to see them)")
	},
}

var findTxCmd = &cobra.Command{
	Use:   "tx",
	Short: "Find transaction",
	Long:  "Find transaction",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		if len(args) < 1 {
			log.Fatal().Msg("transaction must be specified")
		}

		tx := args[0]

		if tx == "" {
			log.Fatal().Msg("transaction must be specified")
		}

		maxBlocks, err := cmd.Flags().GetUint("max-blocks")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get max blocks")
		}

		// Connect to the node.
		ctx := context.Background()
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to node")
		}

		// Get chain height.
		head, err := client.BlockByNumber(ctx, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get chain height")
		}
		blockNumber := head.Number()

		count := uint(0)
		for {
			count++
			block, err := client.BlockByNumber(ctx, blockNumber)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get block")
			}

			found := false
			for _, tx1 := range block.Transactions() {
				if tx1.Hash().Hex() == tx {
					fmt.Printf("block: %d\n", block.Number())
					fmt.Printf("cost: %d\n", tx1.Cost())
					fmt.Printf("data: %x\n", tx1.Data())
					fmt.Printf("gas: %d\n", tx1.Gas())
					fmt.Printf("gas price: %d\n", tx1.GasPrice())
					fmt.Printf("nonce: %d\n", tx1.Nonce())
					fmt.Printf("to: %s\n", tx1.To().Hash().Hex())
					fmt.Printf("value: %d\n", tx1.Value())
					found = true
				}
			}
			if found {
				break
			}

			blockNumber.Sub(blockNumber, big.NewInt(1))
			if blockNumber.Cmp(big.NewInt(0)) <= 0 {
				fmt.Println("genesis block reached")
			}
			if count == maxBlocks {
				fmt.Printf("transaction not found after searching %d blocks\n", maxBlocks)
				break
			}
		}
	},
}
