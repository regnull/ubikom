package cmd

import (
	"fmt"
	"os"

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
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
