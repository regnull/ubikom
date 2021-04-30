package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	defaultURL = "https://one.washmi.net"
)

func init() {
	rootCmd.PersistentFlags().String("url", defaultURL, "server URL")
}

var rootCmd = &cobra.Command{
	Use:   "ubikom-cli",
	Short: "ubikom-cli is a command line client for Ubikom",
	Long:  `ubikom-cli allows you to run local and remote Ubikom commands`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("error executing command")
		os.Exit(1)
	}
}
