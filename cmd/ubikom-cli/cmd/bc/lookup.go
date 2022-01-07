package bc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/ubchain/gocontract"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	lookupKeyCmd.Flags().String("key", "", "key to authorize the transaction")
	lookupKeyCmd.Flags().String("contract-address", "", "contract address")

	lookupCmd.AddCommand(lookupKeyCmd)

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

		instance, err := gocontract.NewKeyRegistry(common.HexToAddress(contractAddress), client)
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

		fmt.Printf("registered: %b\n", registered)
		fmt.Printf("disabled: %b\n", disabled)
		fmt.Printf("owner: %s\n", owner.Hex())
	},
}
