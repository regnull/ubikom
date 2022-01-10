package bc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	lookupKeyCmd.Flags().String("key", "", "key to authorize the transaction")
	lookupKeyCmd.Flags().String("contract-address", globals.KeyRegistryContractAddress, "contract address")

	lookupNameCmd.Flags().String("name", "", "name to look up")
	lookupNameCmd.Flags().String("contract-address", globals.NameRegistryContractAddress, "contract address")

	lookupConnectorCmd.Flags().String("name", "", "name to look up")
	lookupConnectorCmd.Flags().String("protocol", "PL_DMS", "protocol to look up")
	lookupConnectorCmd.Flags().String("contract-address", globals.ConnectorRegistryContractAddress, "contract address")

	lookupCmd.AddCommand(lookupKeyCmd)
	lookupCmd.AddCommand(lookupNameCmd)

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

		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
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

		registered, err := instance.Registered(nil, key.PublicKey().SerializeCompressed())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		disabled, err := instance.Disabled(nil, key.PublicKey().SerializeCompressed())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		owner, err := instance.Owner(nil, key.PublicKey().SerializeCompressed())
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

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load name")
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

		instance, err := gocontract.NewNameRegistryCaller(common.HexToAddress(contractAddress), client)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get contract instance")
		}

		key, err := instance.GetKey(nil, name)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		fmt.Printf("key: %x\n", key)
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

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load name")
		}

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

		location, err := instance.GetLocation(nil, name, protocol)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		fmt.Printf("location: %x\n", location)
	},
}
