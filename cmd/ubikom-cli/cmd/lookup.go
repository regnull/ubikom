package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	cnt "github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	lookupConfigCmd.Flags().String("config-name", "", "protocol to look up")

	lookupCmd.AddCommand(lookupNameCmd)
	lookupCmd.AddCommand(lookupConfigCmd)

	rootCmd.AddCommand(lookupCmd)
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
		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		log.Debug().Str("node-url", nodeURL).Msg("using node")
		contractAddress, err := cmdutil.GetContractAddress(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		log.Debug().Str("contract-address", contractAddress).Msg("using contract")

		if len(args) < 1 {
			log.Fatal().Msg("name must be specified")
		}

		name := args[0]

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		instance, err := cnt.NewNameRegistryCaller(common.HexToAddress(contractAddress), client)
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

		publicKey, err := easyecc.DeserializeCompressed(easyecc.SECP256K1, res.PublicKey)
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

var lookupConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Get config",
	Long:  "Get config",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		log.Debug().Str("node-url", nodeURL).Msg("using node")
		contractAddress, err := cmdutil.GetContractAddress(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}

		if len(args) < 1 {
			log.Fatal().Err(err).Msg("name must be specified")
		}

		name := args[0]

		configName, err := cmd.Flags().GetString("config-name")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load protocol")
		}

		// Connect to the node.
		client, err := ethclient.Dial(nodeURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to blockchain node")
		}

		instance, err := cnt.NewNameRegistryCaller(common.HexToAddress(contractAddress), client)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get contract instance")
		}

		log.Debug().Str("name", name).Str("config-name", configName).Msg("about to look up config")
		configValue, err := instance.LookupConfig(nil, name, configName)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query the config")
		}

		fmt.Printf("%s\n", configValue)
	},
}
