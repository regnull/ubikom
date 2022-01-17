package bc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	nameRegistryContractAddress      common.Address
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
		nameRegistryContractAddress:      common.HexToAddress(nameRegistryContractAddress),
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

func (b *Blockchain) ChangeKeyOwner(ctx context.Context, key *easyecc.PublicKey,
	owner common.Address) (string, error) {
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

	tx, err := instance.ChangeOwner(auth, key.SerializeCompressed(), owner)
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

	instance, err := gocontract.NewNameRegistry(b.nameRegistryContractAddress, b.client)
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

func (b *Blockchain) GetReceipt(ctx context.Context, tx string) (*types.Receipt, error) {
	if len(tx) < 10 {
		return nil, fmt.Errorf("invalid transaction")
	}
	hash, err := hex.DecodeString(tx[2:])
	if err != nil {
		return nil, err
	}
	receipt, err := b.client.TransactionReceipt(ctx, common.BytesToHash(hash))
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (b *Blockchain) MaybeRegisterUser(ctx context.Context, name, password string) error {
	log.Info().Str("user", name).Msg("checking user blockchain registration")
	nameRegCaller, err := gocontract.NewNameRegistry(b.nameRegistryContractAddress, b.client)
	if err != nil {
		return fmt.Errorf("error getting name registry on blockchain: %w", err)
	}
	name = strings.ToLower(util.StripDomainName(name))
	key, err := nameRegCaller.GetKey(nil, name)
	if err != nil {
		return fmt.Errorf("error getting key on blockchain: %w", err)
	}
	if len(key) == 33 {
		log.Debug().Str("user", name).Msg("name is already registered on blockchain")
		return nil
	}
	log.Debug().Str("user", name).Msg("registering user on blockchain")

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

func (b *Blockchain) IsKeyRegistered(ctx context.Context, key *easyecc.PublicKey) (bool, error) {
	keyRegCaller, err := gocontract.NewKeyRegistry(b.keyRegistryContractAddress, b.client)
	if err != nil {
		return false, fmt.Errorf("error getting key registry on blockchain: %w", err)
	}

	return keyRegCaller.Registered(nil, key.SerializeCompressed())
}

func (b *Blockchain) GetKeyByName(ctx context.Context, name string) (*easyecc.PublicKey, error) {
	nameRegCaller, err := gocontract.NewNameRegistry(b.nameRegistryContractAddress, b.client)
	if err != nil {
		return nil, fmt.Errorf("error getting name registry on blockchain: %w", err)
	}
	bb, err := nameRegCaller.GetKey(nil, name)
	if err != nil {
		return nil, fmt.Errorf("error getting key by name: %w", err)
	}
	if len(bb) != 33 {
		return nil, nil
	}
	return easyecc.NewPublicFromSerializedCompressed(bb)
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
