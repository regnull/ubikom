package bc

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	lookupKeyCmd.Flags().String("key", "", "key to look up")
	lookupKeyCmd.Flags().String("key-hex", "", "key to look up (overrides --key)")
	lookupKeyCmd.Flags().String("contract-address", globals.KeyRegistryContractAddress, "contract address")

	lookupNameCmd.Flags().String("contract-address", globals.NameRegistryContractAddress, "contract address")

	lookupConnectorCmd.Flags().String("protocol", "PL_DMS", "protocol to look up")
	lookupConnectorCmd.Flags().String("contract-address", globals.ConnectorRegistryContractAddress, "contract address")

	lookupCmd.AddCommand(lookupKeyCmd)
	lookupCmd.AddCommand(lookupNameCmd)
	lookupCmd.AddCommand(lookupConnectorCmd)

	BCCmd.AddCommand(lookupCmd)
}

var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup stuff on the blockchain",
	Long:  "Lookup stuff on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var lookupKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Get key",
	Long:  "Get key",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		var publicKey []byte

		keyHex, err := cmd.Flags().GetString("key-hex")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get hex key")
		}

		if keyHex != "" {
			publicKey, err = hex.DecodeString(keyHex)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to decode hex key")
			}
		} else {
			key, err := LoadKeyFromFlag(cmd, "key")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load key")
			}
			publicKey = key.PublicKey().SerializeCompressed()
		}

		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		instance, err := gocontract.NewKeyRegistryCaller(common.HexToAddress(contractAddress), client)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get contract instance")
		}

		registered, err := instance.Registered(nil, publicKey)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		disabled, err := instance.Disabled(nil, publicKey)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		owner, err := instance.Owner(nil, publicKey)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		fmt.Printf("registered: %t\n", registered)
		fmt.Printf("disabled: %t\n", disabled)
		fmt.Printf("owner: %s\n", owner.Hex())
	},
}

var lookupNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Get name",
	Long:  "Get name",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		if len(args) < 1 {
			log.Fatal().Msg("name must be specified")
		}

		name := args[0]

		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		instance, err := gocontract.NewNameRegistryCaller(common.HexToAddress(contractAddress), client)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get contract instance")
		}

		key, err := instance.GetKey(nil, name)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		if len(key) == 0 {
			fmt.Printf("name is not registered\n")
			return
		}

		publicKey, err := easyecc.NewPublicFromSerializedCompressed(key)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid key returned")
		}

		fmt.Printf("hex: %x\n", key)
		fmt.Printf("base58: %s\n", base58.Encode(key))
		fmt.Printf("btc addr: %s\n", publicKey.Address())
		fmt.Printf("eth addr: %s\n", publicKey.EthereumAddress())
	},
}

var lookupConnectorCmd = &cobra.Command{
	Use:   "connector",
	Short: "Get connector",
	Long:  "Get connector",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}

		if len(args) < 1 {
			log.Fatal().Err(err).Msg("name must be specified")
		}

		name := args[0]

		protocol, err := cmd.Flags().GetString("protocol")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load protocol")
		}
		if protocol != "PL_DMS" {
			log.Fatal().Msg("invalid protocol")
		}

		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		instance, err := gocontract.NewConnectorRegistryCaller(common.HexToAddress(contractAddress), client)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get contract instance")
		}

		log.Debug().Str("name", name).Str("protocol", protocol).Msg("about to look up")
		location, err := instance.GetLocation(nil, name, protocol)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		fmt.Printf("location: %s\n", location)
	},
}
