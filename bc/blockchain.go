package bc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
)

const (
	registerGasLimit = uint64(300000)
)

var (
	ErrTxNotFound = errors.New("transaction not found")
)

// Blockchain represents Ethereum blockchain.
type Blockchain struct {
	client                           *ethclient.Client
	keyRegistryContractAddress       common.Address
	nameRegistryContractAddres       common.Address
	connectorRegistryContractAddress common.Address
	privateKey                       *easyecc.PrivateKey
}

// NewBlockchain returns a new blockchain.
func NewBlockchain(client *ethclient.Client, keyRegistryContractAddress string,
	nameRegistryContractAddress string, connectorRegistryContractAddress string,
	privateKey *easyecc.PrivateKey) *Blockchain {
	return &Blockchain{
		client:                           client,
		keyRegistryContractAddress:       common.HexToAddress(keyRegistryContractAddress),
		nameRegistryContractAddres:       common.HexToAddress(nameRegistryContractAddress),
		connectorRegistryContractAddress: common.HexToAddress(connectorRegistryContractAddress),
		privateKey:                       privateKey}
}

// RegisterKey registers a new key.
func (b *Blockchain) RegisterKey(ctx context.Context, key *easyecc.PublicKey) (string, error) {
	nonce, err := b.client.PendingNonceAt(ctx,
		common.HexToAddress(b.privateKey.PublicKey().EthereumAddress()))
	if err != nil {
		return "", err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := registerGasLimit

	// Get gas price.
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := b.client.NetworkID(ctx)
	if err != nil {
		return "", err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(b.privateKey.ToECDSA(), chainID)
	if err != nil {
		return "", err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	instance, err := gocontract.NewKeyRegistry(b.keyRegistryContractAddress, b.client)
	if err != nil {
		return "", err
	}

	tx, err := instance.Register(auth, key.SerializeCompressed())
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func (b *Blockchain) RegisterName(ctx context.Context, key *easyecc.PublicKey, name string) (string, error) {
	nonce, err := b.client.PendingNonceAt(ctx,
		common.HexToAddress(b.privateKey.PublicKey().EthereumAddress()))
	if err != nil {
		return "", err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := registerGasLimit

	// Get gas price.
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := b.client.NetworkID(ctx)
	if err != nil {
		return "", err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(b.privateKey.ToECDSA(), chainID)
	if err != nil {
		return "", err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	instance, err := gocontract.NewNameRegistry(b.nameRegistryContractAddres, b.client)
	if err != nil {
		return "", err
	}

	tx, err := instance.Register(auth, name, key.SerializeCompressed())
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func (b *Blockchain) RegisterConnector(ctx context.Context, name string, protocol string, location string) (string, error) {
	nonce, err := b.client.PendingNonceAt(ctx,
		common.HexToAddress(b.privateKey.PublicKey().EthereumAddress()))
	if err != nil {
		return "", err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := registerGasLimit

	// Get gas price.
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := b.client.NetworkID(ctx)
	if err != nil {
		return "", err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(b.privateKey.ToECDSA(), chainID)
	if err != nil {
		return "", err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	instance, err := gocontract.NewConnectorRegistry(b.connectorRegistryContractAddress, b.client)
	if err != nil {
		return "", err
	}

	tx, err := instance.Register(auth, name, protocol, location)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func (b *Blockchain) WaitForConfirmation(ctx context.Context, tx string) (uint64, error) {
	ticker := time.NewTicker(time.Second * 10)
	for {
		block, err := b.findTx(ctx, 10, tx)
		if err == nil {
			return block, nil
		}
		if err != nil && err != ErrTxNotFound {
			return 0, err
		}

		select {
		case <-ticker.C:
			block, err := b.findTx(ctx, 10, tx)
			if err == ErrTxNotFound {
				continue
			}
			if err != nil {
				return 0, err
			}
			return block, nil
		case <-ctx.Done():
			return 0, ErrTxNotFound
		}
	}
}

func (b *Blockchain) MaybeRegisterUser(ctx context.Context, name, password string) error {
	log.Info().Str("user", name).Msg("checking user blockchain registration")
	nameRegCaller, err := gocontract.NewNameRegistry(b.nameRegistryContractAddres, b.client)
	if err != nil {
		return fmt.Errorf("error getting name registry on blockchain: %w", err)
	}
	name = strings.ToLower(util.StripDomainName(name))
	key, err := nameRegCaller.GetKey(nil, name)
	if err != nil {
		return fmt.Errorf("error getting key on blockchain: %w", err)
	}
	if len(key) == 33 {
		log.Debug().Str("user", name).Msg("name is already registered")
		return nil
	}

	privateKey := util.GenerateCanonicalKeyFromNamePassword(name, password)
	keyTx, err := b.RegisterKey(ctx, privateKey.PublicKey())
	if err != nil {
		return fmt.Errorf("error registering key on blockchain: %w", err)
	}
	log.Debug().Str("tx", keyTx).Msg("key is registered")
	nameTx, err := b.RegisterName(ctx, privateKey.PublicKey(), name)
	if err != nil {
		return fmt.Errorf("error registering name on blockchain: %w", err)
	}
	log.Debug().Str("tx", nameTx).Msg("name is registered")
	connectorTx, err := b.RegisterConnector(ctx, name, "PL_DMS", globals.PublicDumpServiceURL)
	if err != nil {
		return fmt.Errorf("error registering connector on blockchain: %w", err)
	}
	log.Debug().Str("tx", connectorTx).Msg("connector is registered")

	ctx1, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	block, err := b.WaitForConfirmation(ctx1, keyTx)
	if err != nil {
		return fmt.Errorf("error waiting for blockchain confirmation: %w", err)
	}
	log.Debug().Str("tx", keyTx).Uint64("block", block).Msg("key registration tx confirmed")
	block, err = b.WaitForConfirmation(ctx1, nameTx)
	if err != nil {
		return fmt.Errorf("error waiting for blockchain confirmation: %w", err)
	}
	log.Debug().Str("tx", nameTx).Uint64("block", block).Msg("name registration tx confirmed")
	block, err = b.WaitForConfirmation(ctx1, connectorTx)
	if err != nil {
		return fmt.Errorf("error waiting for blockchain confirmation: %w", err)
	}
	log.Debug().Str("tx", connectorTx).Uint64("block", block).Msg("connector registration tx confirmed")
	return nil
}

func (b *Blockchain) findTx(ctx context.Context, maxBlocks uint, tx string) (uint64, error) {
	head, err := b.client.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	blockNumber := head.Number()

	count := uint(0)
	for {
		count++
		block, err := b.client.BlockByNumber(ctx, blockNumber)
		if err != nil {
			return 0, err
		}

		for _, tx1 := range block.Transactions() {
			if tx1.Hash().Hex() == tx {
				return block.Number().Uint64(), nil
			}
		}
		blockNumber.Sub(blockNumber, big.NewInt(1))
		if blockNumber.Cmp(big.NewInt(0)) <= 0 {
			return 0, ErrTxNotFound
		}
		if count == maxBlocks {
			return 0, ErrTxNotFound
		}
	}
}
