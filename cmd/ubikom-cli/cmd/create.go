package cmd

import (
	"os"
	"path"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/regnull/ubikom/ecc"
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
			log.Fatal().Err(err).Msg("failed to get output location")
		}
		if out == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get home directory")
			}
			dir := path.Join(homeDir, defaultHomeSubDir)
			_ = os.Mkdir(dir, 0700)
			out = path.Join(dir, defaultKeyFile)
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			log.Fatal().Str("location", out).Msg("key file already exists, if you want to overwrite it, you must first delete it manually")
		}
		privateKey, err := ecc.NewRandomPrivateKey()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to generate private key")
		}
		err = privateKey.Save(out)
		if err != nil {
			log.Fatal().Err(err).Str("location", out).Msg("failed to save private key")
		}
		log.Info().Str("location", out).Msg("private key saved")
	},
}
