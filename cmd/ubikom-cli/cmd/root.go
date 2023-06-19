package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "ubikom-cli",
	Short: "ubikom-cli is a command line client for Ubikom",
	Long:  `ubikom-cli allows you to run local and remote Ubikom commands`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.PersistentFlags().String("network", "main", "mode, either live or prod")
	rootCmd.PersistentFlags().String("node-url", "", "blockchain node location")
	rootCmd.PersistentFlags().String("infura-project-id", "", "infura project id")
	rootCmd.PersistentFlags().String("contract-address", "", "registry contract address")
	rootCmd.PersistentFlags().Uint64("gas-price", 0, "gas price")
	rootCmd.PersistentFlags().Uint64("gas-limit", 0, "gas limit")
}

func Execute() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	cmd, _, err := rootCmd.Find(os.Args[1:])
	// If no command is given, show help.
	if err == nil && cmd.Use == rootCmd.Use && cmd.Flags().Parse(os.Args[1:]) != pflag.ErrHelp {
		args := append([]string{"help"}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("error executing command")
		os.Exit(1)
	}
}
