package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/util"
)

const (
	defaultHomeSubDir = ".ubikom"
	defaultKeyFile    = "key"
)

func init() {
	createKeyCmd.Flags().String("out", "", "Location for the private key file")
	createKeyCmd.Flags().String("from-password", "", "Create private key from the given password")
	createKeyCmd.Flags().Bool("from-mnemonic", false, "create private key from mnemonic")
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

		fromMnemonic, err := cmd.Flags().GetBool("from-mnemonic")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get --from-mnemonic flag")
		}

		var privateKey *easyecc.PrivateKey
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
				saltStr = strings.ToLower(saltStr)
				salt = util.Hash256([]byte(saltStr))
			}
			privateKey = easyecc.NewPrivateKeyFromPassword([]byte(fromPassword), salt[:])
		} else if fromMnemonic {
			var words []string
			reader := bufio.NewReader(os.Stdin)

			for i := 0; i < 24; i++ {
				fmt.Printf("Enter world # %d: ", i+1)
				word, err := reader.ReadString('\n')
				if err != nil {
					log.Fatal().Err(err).Msg("failed to read word")
				}
				word = strings.TrimSuffix(word, "\n")
				word = strings.TrimSuffix(word, "\r")
				words = append(words, word)
			}
			mnemonic := strings.Join(words, " ")
			privateKey, err = easyecc.NewPrivateKeyFromMnemonic(mnemonic)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to create key from mnemonic")
			}
		} else {
			privateKey, err = easyecc.NewRandomPrivateKey()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate private key")
			}
		}

		passphrase, err := util.EnterPassphrase()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get passphase")
		}

		if passphrase == "" {
			log.Warn().Msg("saving private key without passphrase")
		}
		err = privateKey.Save(out, passphrase)
		if err != nil {
			log.Fatal().Err(err).Str("location", out).Msg("failed to save private key")
		}
		log.Info().Str("location", out).Msg("private key saved")
	},
}
