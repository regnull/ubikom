package bc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubchain/gocontract"
	cntv2 "github.com/regnull/ubchain/gocontract/v2"
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	lookupNameCmd.Flags().String("contract-address", globals.NameRegistryContractAddress, "contract address")

	lookupConnectorCmd.Flags().String("protocol", "PL_DMS", "protocol to look up")
	lookupConnectorCmd.Flags().String("contract-address", globals.ConnectorRegistryContractAddress, "contract address")

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

type lookupNameRes struct {
	PublicKey string
	Owner     string
	Price     int64
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

		instance, err := cntv2.NewNameRegistryCaller(common.HexToAddress(contractAddress), client)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get contract instance")
		}

		res, err := instance.LookupName(nil, name)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the key")
		}

		zeroAddress := common.BigToAddress(big.NewInt(0))
		if bytes.Equal(res.Owner.Bytes(), zeroAddress.Bytes()) {
			fmt.Printf("name is not registered\n")
			return
		}

		publicKey, err := easyecc.NewPublicFromSerializedCompressed(res.PublicKey)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid key returned")
		}

		cmdRes := &lookupNameRes{
			PublicKey: fmt.Sprintf("0x%x", publicKey.SerializeCompressed()),
			Owner:     fmt.Sprintf("0x%x", res.Owner),
			Price:     res.Price.Int64(),
		}

		s, err := json.MarshalIndent(cmdRes, "", "  ")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to marshal json")
		}

		fmt.Printf("%s\n", s)
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
