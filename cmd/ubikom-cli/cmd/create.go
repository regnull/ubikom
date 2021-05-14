package cmd

import (
	"fmt"
	"os"
	"path"

	"crypto/rand"

	"github.com/btcsuite/btcutil/base58"
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
	createKeyCmd.Flags().String("from-password", "", "Create private key from the given password")
	createKeyCmd.Flags().String("salt", "", "Salt used for private key creation")
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

		fromPassword, err := cmd.Flags().GetString("from-password")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get password")
		}

		var privateKey *ecc.PrivateKey
		if fromPassword != "" {
			if len(fromPassword) < 8 {
				log.Fatal().Err(err).Msg("password must be at least 8 characters long")
			}
			saltStr, err := cmd.Flags().GetString("salt")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get salt")
			}
			var salt []byte
			if saltStr != "" {
				salt = base58.Decode(saltStr)
			} else {
				var saltArr [8]byte
				_, err := rand.Read(saltArr[:])
				if err != nil {
					log.Fatal().Err(err).Msg("failed to generate salt")
				}
				salt = saltArr[:]
			}
			fmt.Printf("salt: %s\n", base58.Encode(salt[:]))
			privateKey = ecc.NewPrivateKeyFromPassword([]byte(fromPassword), salt[:])
		} else {
			privateKey, err = ecc.NewRandomPrivateKey()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate private key")
			}
		}

		err = privateKey.Save(out)
		if err != nil {
			log.Fatal().Err(err).Str("location", out).Msg("failed to save private key")
		}
		log.Info().Str("location", out).Msg("private key saved")
	},
}
