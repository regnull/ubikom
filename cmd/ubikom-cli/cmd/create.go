package cmd

import (
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"

	"teralyt.com/ubikom/ecc"
)

const (
	defaultHomeSubDir = ".ubikom"
	defaultKeyFile    = "key"
)

func init() {
	createKeyCmd.Flags().String("out", "", "Location for the private key file")
	createCmd.AddCommand(createKeyCmd)
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create various things",
	Long:  "Create various things",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var createKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Create private key",
	Long:  "Create private key",
	Run: func(cmd *cobra.Command, args []string) {
		out, err := cmd.Flags().GetString("out")
		if err != nil {
			log.Fatal(err)
		}
		if out == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal(err)
			}
			dir := path.Join(homeDir, defaultHomeSubDir)
			_ = os.Mkdir(dir, 0700)
			out = path.Join(dir, defaultKeyFile)
		}
		privateKey, err := ecc.NewRandomPrivateKey()
		if err != nil {
			log.Fatal(err)
		}
		err = privateKey.Save(out)
		if err != nil {
			log.Fatal(err)
		}
	},
}
