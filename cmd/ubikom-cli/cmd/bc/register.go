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
	cntv2 "github.com/regnull/ubchain/gocontract/v2"
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	registerNameCmd.Flags().String("key", "", "key to authorize the transaction")
	registerNameCmd.Flags().String("enc-key", "", "key to register")
	registerNameCmd.Flags().String("name", "", "name to register")
	registerNameCmd.Flags().String("contract-address", globals.NameRegistryContractAddress, "contract address")

	registerCmd.AddCommand(registerNameCmd)

	BCCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register various things on the blockchain",
	Long:  "Register various things on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("sub-command requried (do 'ubikom-cli bc register --help' to see available commands)")
	},
}

type mutateStateFunc func(client *ethclient.Client, auth *bind.TransactOpts,
	addr common.Address) (*types.Transaction, error)

var registerNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Register name on the blockchain",
	Long:  "Register name on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		encKey, err := LoadKeyFromFlag(cmd, "enc-key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load reg key")
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get name")
		}
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		err = interactWithContract(nodeURL, key, contractAddress,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := cntv2.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.RegisterName(auth, encKey.PublicKey().SerializeCompressed(), name)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to register name")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register name")
		}
	},
}

func interactWithContract(nodeURL string, key *easyecc.PrivateKey,
	contractAddress string, f mutateStateFunc) error {
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
	gasLimit := uint64(300000)

	// Get gas price.
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

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
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	addr := common.HexToAddress(contractAddress)

	tx, err := f(client, auth, addr)
	if err != nil {
		return err
	}
	fmt.Printf("tx sent: %s\n", tx.Hash().Hex())

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
