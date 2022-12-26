package bc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const suggestedGasLimit = 3000000

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

func WaitMined(client *ethclient.Client, tx *types.Transaction, waitDuration time.Duration) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

type mutateStateFunc func(client *ethclient.Client, auth *bind.TransactOpts,
	addr common.Address) (*types.Transaction, error)

func interactWithContract(nodeURL string, key *easyecc.PrivateKey,
	contractAddress string, value int64, gasPrice uint64, gasLimit uint64, f mutateStateFunc) error {
	// Connect to the node.
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Get nonce.
	nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(key.PublicKey().EthereumAddress()))
	if err != nil {
		return err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	if gasLimit == 0 {
		gasLimit = uint64(suggestedGasLimit)
	}

	// Get gas price.
	if gasPrice == 0 {
		gasPriceBI, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return err
		}
		gasPrice = gasPriceBI.Uint64()
	}
	log.Debug().Uint64("gas-price", gasPrice).Msg("got gas price")

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(key.ToECDSA(), chainID)
	if err != nil {
		return err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(value) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = big.NewInt(int64(gasPrice))

	addr := common.HexToAddress(contractAddress)

	tx, err := f(client, auth, addr)
	if err != nil {
		return err
	}
	log.Info().Str("tx", tx.Hash().Hex()).Msg("tx sent")

	res, err := WaitMined(client, tx, time.Second*30)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get transaction receipt")
	}
	jsonBytes, err := json.MarshalIndent(res, "", "  ")
	fmt.Printf("%s\n", jsonBytes)

	// It's not entirely clear how to see when a write transaction failed, because the contract is
	// not there, etc. The only way I've found, is to look at logs, which are empty if the
	// contract address is wrong. Status is 1 regardless, but we look at the status as well,
	// just in case.
	if len(res.Logs) == 0 || res.Status == 0 {
		log.Error().Msg("transaction failed")
		return fmt.Errorf("transaction failed")
	}
	return nil
}

func getNodeURL(flags *pflag.FlagSet) (string, error) {
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
	if mode == "main" {
		return globals.InfuraNodeURL, nil
	} else if mode == "sepolia" {
		log.Warn().Msg("using Sepolia testnet")
		return globals.InfuraSepoliaNodeURL, nil
	}
	return "", fmt.Errorf("invalid network, must be main or sepolia")
}

func getContractAddress(flags *pflag.FlagSet) (string, error) {
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
