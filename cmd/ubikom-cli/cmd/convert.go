package cmd

import "github.com/spf13/cobra"

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert various things",
	Long:  "Convert various things",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var convertKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Convert private key",
	Long:  "Convert private key",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
