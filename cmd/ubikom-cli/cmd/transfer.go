package cmd

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	transferCmd.Flags().String("key", "", "key to authorize the transaction")
	transferCmd.Flags().String("keystore-key", "", "keystore key to authorize the transaction")
	transferCmd.Flags().String("keystore-password", "", "password to unlock the keystore key")
	transferCmd.Flags().String("keystore-path", "", "keystore path")
	transferCmd.Flags().String("to", "", "receiver's account")
	transferCmd.Flags().Uint64("gas-limit", 21000, "gas limit")
	transferCmd.Flags().String("value", "", "value to transfer")

	rootCmd.AddCommand(transferCmd)
}

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer funds between accounts",
	Long:  "Transfer funds between accounts",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		var ks *keystore.KeyStore
		var accountFrom accounts.Account
		var accountAddress common.Address
		var key *easyecc.PrivateKey
		keystoreKey, err := cmd.Flags().GetString("keystore-key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get keystore key")
		}

		if keystoreKey != "" {
			// Use the key from the keystore.

			keystoreDir, err := cmd.Flags().GetString("keystore-path")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get keystore path")
			}

			if keystoreDir == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get home directory")
				}
				// TODO: Support this for other OSes.
				keystoreDir = path.Join(homeDir, "Library/Ethereum/keystore")
			}

			keystorePassword, err := cmd.Flags().GetString("keystore-password")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get keystore password")
			}

			ks = keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
			accountFrom, err = ks.Find(accounts.Account{Address: common.HexToAddress(keystoreKey)})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to find the originator's account")
			}
			err = ks.Unlock(accountFrom, keystorePassword)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to unlock the account")
			}
			accountAddress = accountFrom.Address

		} else {
			key, err = cmdutil.LoadKeyFromFlag(cmd, "key")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load key")
			}
			ethereumAddress, _ := key.PublicKey().EthereumAddress()
			accountAddress = common.HexToAddress(ethereumAddress)
		}

		to, err := cmd.Flags().GetString("to")
		if err != nil {
			log.Fatal().Err(err).Msg("--to must be specified")
		}

		// Connect to the node.
		ctx := context.Background()
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to node")
		}
		// Get nonce.
		nonce, err := client.PendingNonceAt(ctx, accountAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get nonce")
		}
		fmt.Printf("got nonce: %d\n", nonce)

		// Recommended gas limit.
		gasLimit, err := cmd.Flags().GetUint64("gas-limit")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas limit")
		}

		// Get gas price.
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get gas price")
		}
		fmt.Printf("gas price: %d\n", gasPrice)

		// Send to address.
		toAddress := common.HexToAddress(to)

		// Parse value.
		value, err := cmd.Flags().GetString("value")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get value")
		}
		valueNum := new(big.Int)
		_, ok := valueNum.SetString(value, 10)
		if !ok {
			log.Fatal().Msg("failed to parse value")
		}

		// Create and sign the transaction.
		tx := types.NewTransaction(nonce, toAddress, valueNum, gasLimit, gasPrice, nil)

		chainID, err := client.NetworkID(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get chain id")
		}
		fmt.Printf("chain ID: %d\n", chainID)

		var signedTx *types.Transaction
		// Depending on what kind of a key we have, sign either using keystore key,
		// or easyecc private key.
		if ks != nil {
			signedTx, err = ks.SignTx(accountFrom, tx, chainID)
		} else {
			signedTx, err = types.SignTx(tx, types.NewEIP155Signer(chainID), key.ToECDSA())
		}

		if err != nil {
			log.Fatal().Err(err).Msg("failed to sign transaction")
		}

		// Send transaction.
		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send transaction")
		}

		fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
	},
}
