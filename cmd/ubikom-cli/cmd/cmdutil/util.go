package cmdutil

import (
	"fmt"
	"os"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func LoadKeyFromFlag(cmd *cobra.Command, keyFlagName string) (*easyecc.PrivateKey, error) {
	keyFile, err := cmd.Flags().GetString(keyFlagName)
	if err != nil {
		return nil, err
	}

	if keyFile == "" {
		keyFile, err = util.GetDefaultKeyLocation()
		if err != nil {
			return nil, err
		}
	}

	encrypted, err := util.IsKeyEncrypted(keyFile)
	if err != nil {
		return nil, err
	}

	passphrase := ""
	if encrypted {
		passphrase, err = util.ReadPassphase()
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, passphrase)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func GetNodeURL(flags *pflag.FlagSet) (string, error) {
	nodeURL, err := flags.GetString("node-url")
	if err != nil {
		return "", fmt.Errorf("failed to get node URL")
	}
	if nodeURL != "" {
		return nodeURL, nil
	}
	mode, err := flags.GetString("network")
	if err != nil {
		return "", fmt.Errorf("failed to get network")
	}
	infuraId, err := flags.GetString("infura-project-id")
	if err != nil {
		return "", fmt.Errorf("failed to get infura project id")
	}
	if infuraId == "" {
		infuraId = os.Getenv("INFURA_PROJECT_ID")
	}
	if infuraId == "" {
		return "", fmt.Errorf("infura project id must be specified")
	}
	if mode == "main" {
		return fmt.Sprintf(globals.InfuraNodeURL, infuraId), nil
	} else if mode == "sepolia" {
		log.Warn().Msg("using Sepolia testnet")
		return fmt.Sprintf(globals.InfuraSepoliaNodeURL, infuraId), nil
	}
	return "", fmt.Errorf("invalid network, must be main or sepolia")
}

func GetContractAddress(flags *pflag.FlagSet) (string, error) {
	contractAddress, err := flags.GetString("contract-address")
	if err != nil {
		return "", fmt.Errorf("failed to get node contract address")
	}
	if contractAddress != "" {
		return contractAddress, nil
	}
	mode, err := flags.GetString("network")
	if err != nil {
		return "", fmt.Errorf("failed to get network")
	}

	if mode == "main" {
		return globals.MainnetNameRegistryAddress, nil
	} else if mode == "sepolia" {
		log.Warn().Msg("using Sepolia testnet")
		return globals.SepoliaNameRegistryAddress, nil
	}
	return "", fmt.Errorf("invalid network, must be main or sepolia")
}
