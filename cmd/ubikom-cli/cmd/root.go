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
		// Do Stuff Here
	},
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
