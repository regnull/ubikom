package bc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/globals"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	registerKeyCmd.Flags().String("key", "", "key to authorize the transaction")
	registerKeyCmd.Flags().String("reg-key", "", "key to register")
	registerKeyCmd.Flags().String("contract-address", globals.KeyRegistryContractAddress, "contract address")

	registerNameCmd.Flags().String("key", "", "key to authorize the transaction")
	registerKeyCmd.Flags().String("reg-key", "", "key to register")
	registerNameCmd.Flags().String("name", "", "name to register")
	registerNameCmd.Flags().String("contract-address", globals.KeyRegistryContractAddress, "contract address")

	registerConnectorCmd.Flags().String("key", "", "key to authorize the transaction")
	registerConnectorCmd.Flags().String("name", "", "name")
	registerConnectorCmd.Flags().String("protocol", "", "protocol")
	registerConnectorCmd.Flags().String("location", "", "location to register for this name/protocol")

	registerCmd.AddCommand(registerKeyCmd)
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

var registerKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Register key on the blockchain",
	Long:  "Register key on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		regKey, err := LoadKeyFromFlag(cmd, "reg-key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load reg key")
		}
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}

		tx, err := interactWithContract(nodeURL, key, contractAddress,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := gocontract.NewKeyRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.Register(auth, regKey.PublicKey().SerializeCompressed())
				if err != nil {
					log.Fatal().Err(err).Msg("failed to register key")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register key")
		}
		fmt.Printf("tx sent: %s\n", tx.Hash().Hex())
	},
}

var registerNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Register name on the blockchain",
	Long:  "Register name on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		regKey, err := LoadKeyFromFlag(cmd, "reg-key")
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
		tx, err := interactWithContract(nodeURL, key, contractAddress,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := gocontract.NewNameRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.Register(auth, name, regKey.PublicKey().SerializeCompressed())
				if err != nil {
					log.Fatal().Err(err).Msg("failed to register name")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register name")
		}
		fmt.Printf("tx sent: %s\n", tx.Hash().Hex())
	},
}

var registerConnectorCmd = &cobra.Command{
	Use:   "connector",
	Short: "Register connector on the blockchain",
	Long:  "Register connector on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load key")
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get name")
		}
		protocol, err := cmd.Flags().GetString("protocol")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get protocol")
		}
		location, err := cmd.Flags().GetString("location")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get location")
		}
		nodeURL, err := cmd.Flags().GetString("node-url")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		contractAddress, err := cmd.Flags().GetString("contract-address")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		tx, err := interactWithContract(nodeURL, key, contractAddress,
			func(client *ethclient.Client, auth *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
				instance, err := gocontract.NewConnectorRegistry(addr, client)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get contract instance")
				}

				tx, err := instance.Register(auth, name, protocol, location)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to register name")
				}
				return tx, err
			})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register connector")
		}
		fmt.Printf("tx sent: %s\n", tx.Hash().Hex())
	},
}

func interactWithContract(nodeURL string, key *easyecc.PrivateKey,
	contractAddress string, f mutateStateFunc) (*types.Transaction, error) {
	// Connect to the node.
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Get nonce.
	nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(key.PublicKey().EthereumAddress()))
	if err != nil {
		return nil, err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := uint64(300000)

	// Get gas price.
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(key.ToECDSA(), chainID)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	addr := common.HexToAddress(contractAddress)

	return f(client, auth, addr)
}
