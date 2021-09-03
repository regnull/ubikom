package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"

	"crypto/rand"

	"github.com/btcsuite/btcutil/base58"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/regnull/easyecc"
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

		fmt.Print("Passphrase (enter for none): ")
		bytePassphrase, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read passphrase")
		}
		passphrase1 := string(bytePassphrase)

		fmt.Print("\nConfirm passphrase (enter for none): ")
		bytePassphrase, err = term.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read passphrase")
		}
		passphrase2 := string(bytePassphrase)
		if passphrase1 != passphrase2 {
			log.Fatal().Msg("passphrase mismatch")
		}

		if passphrase1 == "" {
			log.Warn().Msg("saving private key without passphrase")
		}
		err = privateKey.Save(out, passphrase1)
		if err != nil {
			log.Fatal().Err(err).Str("location", out).Msg("failed to save private key")
		}
		log.Info().Str("location", out).Msg("private key saved")
	},
}
