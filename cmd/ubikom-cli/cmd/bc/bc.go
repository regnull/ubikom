package bc

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	BCCmd.PersistentFlags().String("network", "main", "mode, either live or prod")
	BCCmd.PersistentFlags().String("node-url", "", "blockchain node location")
	BCCmd.PersistentFlags().String("contract-address", "", "registry contract address")
	BCCmd.PersistentFlags().Uint64("gas-price", 0, "gas price")
	BCCmd.PersistentFlags().Uint64("gas-limit", 0, "gas limit")
}

var BCCmd = &cobra.Command{
	Use:   "bc",
	Short: "Blockchain-related commands",
	Long:  "Blockchain-related commands",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("bc requires sub-command (do 'ubikom-cli bc --help' to see available commands)")
	},
}
