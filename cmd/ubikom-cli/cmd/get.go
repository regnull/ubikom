package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	getAddressCmd.Flags().String("key", "", "Location for the private key file")

	getEthereumAddressCmd.Flags().String("key", "", "Location for the private key file")

	getBitcoinAddressCmd.Flags().String("key", "", "Location for the private key file")

	getPublicKeyCmd.Flags().String("key", "", "Location for the private key file")

	getUserPublicKey.Flags().String("name", "", "User name")

	getMnemonicCmd.Flags().String("key", "", "Location for the private key file")

	getCmd.AddCommand(getAddressCmd)
	getCmd.AddCommand(getEthereumAddressCmd)
	getCmd.AddCommand(getBitcoinAddressCmd)
	getCmd.AddCommand(getPublicKeyCmd)
	getCmd.AddCommand(getMnemonicCmd)
	getCmd.AddCommand(getUserPublicKey)
	getCmd.AddCommand(getBalanceCmd)
	getCmd.AddCommand(getBlockCmd)
	getCmd.AddCommand(getReceiptCmd)

	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get various things",
	Long:  "Get various things",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var getEthereumAddressCmd = &cobra.Command{
	Use:   "ethereum-address",
	Short: "Get Ethereum address",
	Long:  "Get Ethereum address for the specified key",
	Run: func(cmd *cobra.Command, args []string) {
		keyFile, err := cmd.Flags().GetString("key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key location")
		}

		if keyFile == "" {
			keyFile, err = util.GetDefaultKeyLocation()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get private key")
			}
		}

		encrypted, err := util.IsKeyEncrypted(keyFile)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot find key file")
		}

		passphrase := ""
		if encrypted {
			passphrase, err = util.ReadPassphase()
			if err != nil {
				log.Fatal().Err(err).Msg("cannot read passphrase")
			}
		}

		privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, passphrase)
		if err != nil {
			log.Fatal().Err(err).Str("location", keyFile).Msg("cannot load private key")
		}

		ethereumAddress, _ := privateKey.PublicKey().EthereumAddress()
		fmt.Printf("%s\n", ethereumAddress)
	},
}

var getAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Get Ethereum address",
	Long:  "Get Ethereum address for the specified key",
	Run: func(cmd *cobra.Command, args []string) {
		getEthereumAddressCmd.Run(cmd, args)
	},
}

var getBitcoinAddressCmd = &cobra.Command{
	Use:   "bitcoin-address",
	Short: "Get Bitcoin address",
	Long:  "Get Bitcoin address for the specified key",
	Run: func(cmd *cobra.Command, args []string) {
		keyFile, err := cmd.Flags().GetString("key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key location")
		}

		if keyFile == "" {
			keyFile, err = util.GetDefaultKeyLocation()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get private key")
			}
		}

		encrypted, err := util.IsKeyEncrypted(keyFile)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot find key file")
		}

		passphrase := ""
		if encrypted {
			passphrase, err = util.ReadPassphase()
			if err != nil {
				log.Fatal().Err(err).Msg("cannot read passphrase")
			}
		}

		privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, passphrase)
		if err != nil {
			log.Fatal().Err(err).Str("location", keyFile).Msg("cannot load private key")
		}

		bitcoinAddress, _ := privateKey.PublicKey().BitcoinAddress()
		fmt.Printf("%s\n", bitcoinAddress)
	},
}

var getPublicKeyCmd = &cobra.Command{
	Use:   "public-key",
	Short: "Get public key",
	Long:  "Get public key",
	Run: func(cmd *cobra.Command, args []string) {
		privateKey, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load private key")
		}

		fmt.Printf("0x%0x\n", privateKey.PublicKey().SerializeCompressed())
	},
}

var getUserPublicKey = &cobra.Command{
	Use:   "user-public-key",
	Short: "Get public key by user name and password",
	Long:  "Get public key which is generated from user name and password",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get name")
		}

		if name == "" {
			log.Fatal().Msg("--name must be specified")
		}

		password, err := util.ReadPassphase()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read passphrase")
		}
		n := strings.TrimSpace(name)
		n = util.StripDomainName(n)
		privateKey := easyecc.NewPrivateKeyFromPassword([]byte(password),
			util.Hash256([]byte(strings.ToLower(n))))

		fmt.Printf("0x%0x\n", privateKey.PublicKey().SerializeCompressed())
	},
}

var getMnemonicCmd = &cobra.Command{
	Use:   "mnemonic",
	Short: "Get key mnemonic",
	Long:  "Get private key mnemonic",
	Run: func(cmd *cobra.Command, args []string) {
		keyFile, err := cmd.Flags().GetString("key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key location")
		}

		if keyFile == "" {
			keyFile, err = util.GetDefaultKeyLocation()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get private key")
			}
		}

		encrypted, err := util.IsKeyEncrypted(keyFile)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot find key file")
		}

		passphrase := ""
		if encrypted {
			passphrase, err = util.ReadPassphase()
			if err != nil {
				log.Fatal().Err(err).Msg("cannot read passphrase")
			}
		}

		privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, passphrase)
		if err != nil {
			log.Fatal().Err(err).Str("location", keyFile).Msg("cannot load private key")
		}

		mnemonic, err := privateKey.Mnemonic()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key mnemonic")
		}
		words := strings.Split(mnemonic, " ")
		for i, word := range words {
			fmt.Printf("%d: \t%s\n", i+1, word)
		}
	},
}

var getBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Get balance",
	Long:  "Get balance",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		if len(args) < 1 {
			log.Fatal().Msg("address is required")
		}
		address := args[0]

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		// Get balance.
		ctx := context.Background()
		balance, err := client.BalanceAt(ctx, common.HexToAddress(address), nil)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get balance")
		}
		fmt.Printf("%d\n", balance)
	},
}

var getBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Get latest block number",
	Long:  "Get latest block number",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		ctx := context.Background()
		num, err := client.BlockNumber(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get latest block number")
		}
		fmt.Printf("%d\n", num)
	},
}

var getReceiptCmd = &cobra.Command{
	Use:   "receipt",
	Short: "Get transaction receipt",
	Long:  "Get transaction receipt",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		if len(args) < 1 {
			log.Fatal().Msg("transaction hash required")
		}

		tx := args[0]
		if len(tx) < 10 {
			log.Fatal().Msg("transaction hash required")
		}

		if strings.HasPrefix("0x", tx) {
			tx = tx[2:]
		}

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		ctx := context.Background()

		hash, err := hex.DecodeString(tx)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid transaction")
		}
		receipt, err := client.TransactionReceipt(ctx, common.BytesToHash(hash))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get transaction receipt")
		}

		fmt.Printf("status: %d\n", receipt.Status)
		fmt.Printf("block hash: %s\n", receipt.BlockHash.Hex())
		fmt.Printf("block number: %d\n", receipt.BlockNumber)
		fmt.Printf("gas used: %d\n", receipt.GasUsed)
		fmt.Printf("tx index: %d\n", receipt.TransactionIndex)
		fmt.Printf("tx type: %d\n", receipt.Type)
	},
}
