package bc

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	defaultNodeURL = "http://18.223.40.196:8545"
)

func init() {
	BCCmd.PersistentFlags().String("node-url", defaultNodeURL, "blockchain node location")
}

var BCCmd = &cobra.Command{
	Use:   "bc",
	Short: "Blockchain-related commands",
	Long:  "Blockchain-related commands",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("bc requires sub-command (do 'ubikom-cli bc --help' to see available commands)")
	},
}
