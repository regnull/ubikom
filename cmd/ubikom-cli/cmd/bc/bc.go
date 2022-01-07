package bc

import (
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	BCCmd.PersistentFlags().String("node-url", globals.BlockchainNodeURL, "blockchain node location")
}

var BCCmd = &cobra.Command{
	Use:   "bc",
	Short: "Blockchain-related commands",
	Long:  "Blockchain-related commands",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("bc requires sub-command (do 'ubikom-cli bc --help' to see available commands)")
	},
}
