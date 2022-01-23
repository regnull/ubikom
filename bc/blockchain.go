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
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	registerGasLimit = uint64(300000)
)

var (
	ErrTxNotFound       = errors.New("transaction not found")
	ErrKeyNotConfigured = errors.New("private key is not configured")
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
func (b *Blockchain) RegisterKey(ctx context.Context, key *easyecc.PublicKey) (*types.Receipt, error) {
	if b.privateKey == nil {
		return nil, ErrKeyNotConfigured
	}
	nonce, err := b.client.PendingNonceAt(ctx,
		common.HexToAddress(b.privateKey.PublicKey().EthereumAddress()))
	if err != nil {
		return nil, err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := registerGasLimit

	// Get gas price.
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := b.client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(b.privateKey.ToECDSA(), chainID)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	instance, err := gocontract.NewKeyRegistry(b.keyRegistryContractAddress, b.client)
	if err != nil {
		return nil, err
	}

	tx, err := instance.Register(auth, key.SerializeCompressed())
	if err != nil {
		return nil, err
	}

	return bind.WaitMined(ctx, b.client, tx)
}

func (b *Blockchain) ChangeKeyOwner(ctx context.Context, key *easyecc.PublicKey,
	owner common.Address) (*types.Receipt, error) {
	if b.privateKey == nil {
		return nil, ErrKeyNotConfigured
	}
	nonce, err := b.client.PendingNonceAt(ctx,
		common.HexToAddress(b.privateKey.PublicKey().EthereumAddress()))
	if err != nil {
		return nil, err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := registerGasLimit

	// Get gas price.
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := b.client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(b.privateKey.ToECDSA(), chainID)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	instance, err := gocontract.NewKeyRegistry(b.keyRegistryContractAddress, b.client)
	if err != nil {
		return nil, err
	}

	tx, err := instance.ChangeOwner(auth, key.SerializeCompressed(), owner)
	if err != nil {
		return nil, err
	}

	return bind.WaitMined(ctx, b.client, tx)
}

func (b *Blockchain) RegisterName(ctx context.Context, key *easyecc.PublicKey, name string) (*types.Receipt, error) {
	if b.privateKey == nil {
		return nil, ErrKeyNotConfigured
	}
	nonce, err := b.client.PendingNonceAt(ctx,
		common.HexToAddress(b.privateKey.PublicKey().EthereumAddress()))
	if err != nil {
		return nil, err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := registerGasLimit

	// Get gas price.
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := b.client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(b.privateKey.ToECDSA(), chainID)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	instance, err := gocontract.NewNameRegistry(b.nameRegistryContractAddress, b.client)
	if err != nil {
		return nil, err
	}

	tx, err := instance.Register(auth, name, key.SerializeCompressed())
	if err != nil {
		return nil, err
	}

	return bind.WaitMined(ctx, b.client, tx)
}

func (b *Blockchain) RegisterConnector(ctx context.Context, name string, protocol string, location string) (*types.Receipt, error) {
	if b.privateKey == nil {
		return nil, ErrKeyNotConfigured
	}
	nonce, err := b.client.PendingNonceAt(ctx,
		common.HexToAddress(b.privateKey.PublicKey().EthereumAddress()))
	if err != nil {
		return nil, err
	}
	log.Debug().Uint64("nonce", nonce).Msg("got nonce")

	// Recommended gas limit.
	gasLimit := registerGasLimit

	// Get gas price.
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("gas-price", gasPrice.String()).Msg("got gas price")

	chainID, err := b.client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("chain-id", chainID.String()).Msg("got chain ID")

	auth, err := bind.NewKeyedTransactorWithChainID(b.privateKey.ToECDSA(), chainID)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	instance, err := gocontract.NewConnectorRegistry(b.connectorRegistryContractAddress, b.client)
	if err != nil {
		return nil, err
	}

	tx, err := instance.Register(auth, name, protocol, location)
	if err != nil {
		return nil, err
	}

	return bind.WaitMined(ctx, b.client, tx)
}

func (b *Blockchain) WaitForConfirmation(ctx context.Context, tx string) (uint64, error) {
	// TODO: Remove this? With WaitMined working, this one is not necessary.
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

func (b *Blockchain) MaybeRegisterUser(ctx context.Context, name, regName, password string) error {
	if b.privateKey == nil {
		return ErrKeyNotConfigured
	}
	log.Info().Str("user", name).Msg("checking user blockchain registration")

	name = strings.ToLower(util.StripDomainName(name))
	regName = strings.ToLower(util.StripDomainName(regName))

	privateKey := util.GenerateCanonicalKeyFromNamePassword(name, password)

	// Check key registration.
	keyRegCaller, err := gocontract.NewKeyRegistryCaller(b.keyRegistryContractAddress, b.client)
	if err != nil {
		return err
	}

	keyRegistered, err := keyRegCaller.Registered(nil, privateKey.PublicKey().SerializeCompressed())
	if err != nil {
		return err
	}

	iAmTheOwner := true
	ubikomIsTheOwner := false
	if !keyRegistered {
		log.Info().Str("name", regName).Msg("registering key")

		receipt, err := b.RegisterKey(ctx, privateKey.PublicKey())
		if err != nil {
			return fmt.Errorf("error registering key on blockchain: %w", err)
		}
		log.Info().Interface("receipt", receipt).Str("name", regName).Msg("key registered")
	} else {
		owner, err := keyRegCaller.Owner(nil, privateKey.PublicKey().SerializeCompressed())
		if err != nil {
			return err
		}
		iAmTheOwner = util.EqualHexStrings(owner.Hex(), b.privateKey.PublicKey().EthereumAddress())
		ubikomIsTheOwner = util.EqualHexStrings(owner.Hex(), globals.UbikomEthereumAddress)
		log.Info().Str("name", regName).Bool("i-am-the-owner", iAmTheOwner).Str("owner", owner.Hex()).Msg("key is already registered")
	}

	// Check name registration.

	nameRegCaller, err := gocontract.NewNameRegistryCaller(b.nameRegistryContractAddress, b.client)
	if err != nil {
		return err
	}
	key, err := nameRegCaller.GetKey(nil, regName)
	if err != nil {
		return err
	}

	nameRegistered := len(key) == 33
	if !nameRegistered {
		if iAmTheOwner {
			log.Info().Str("name", regName).Msg("registering name")

			receipt, err := b.RegisterName(ctx, privateKey.PublicKey(), regName)
			if err != nil {
				return err
			}
			log.Info().Interface("receipt", receipt).Str("name", regName).Msg("name registered")
		} else {
			log.Warn().Str("name", regName).Msg("cannot register name, I'm not the key owner")
		}
	} else {
		log.Info().Str("name", regName).Msg("name is already registered")
	}

	// Check connector registration.

	connectorRegCaller, err := gocontract.NewConnectorRegistryCaller(b.connectorRegistryContractAddress, b.client)
	if err != nil {
		return err
	}

	location, err := connectorRegCaller.GetLocation(nil, regName, "PL_DMS")
	if err != nil {
		return err
	}

	connectorRegistered := location != ""
	if !connectorRegistered {
		if iAmTheOwner {
			log.Info().Str("name", regName).Msg("registering connector")

			receipt, err := b.RegisterConnector(ctx, regName, "PL_DMS", globals.PublicDumpServiceURL)
			if err != nil {
				return err
			}
			log.Info().Interface("receipt", receipt).Str("name", name).Msg("connector registered")
		} else {
			log.Warn().Str("name", regName).Msg("cannot register connector, I'm not the key owner")
		}
	} else {
		log.Info().Str("name", regName).Msg("connector is already registered")
	}

	// Change key ownership, if needed.

	if ubikomIsTheOwner {
		log.Info().Str("name", regName).Msg("key ownership is already correct")
	} else {
		if iAmTheOwner {
			log.Info().Str("name", regName).Msg("changing key ownership")

			receipt, err := b.ChangeKeyOwner(ctx, privateKey.PublicKey(), common.HexToAddress(globals.UbikomEthereumAddress))
			if err != nil {
				return err
			}
			log.Info().Interface("receipt", receipt).Str("name", regName).Msg("key ownership changed")
		} else {
			log.Warn().Str("name", regName).Msg("key ownership is incorrect, but I can't change it")
		}
	}
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

func (b *Blockchain) LookupKey(ctx context.Context, in *pb.LookupKeyRequest, opts ...grpc.CallOption) (*pb.LookupKeyResponse, error) {
	caller, err := gocontract.NewKeyRegistryCaller(b.keyRegistryContractAddress, b.client)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: ctx}

	registered, err := caller.Registered(callOpts, in.Key)
	if err != nil {
		return nil, err
	}

	if !registered {
		return nil, status.Error(codes.NotFound, "key not found")
	}

	disabled, err := caller.Disabled(callOpts, in.Key)
	if err != nil {
		return nil, err
	}
	return &pb.LookupKeyResponse{
		Disabled: disabled,
	}, nil
}

func (b *Blockchain) LookupName(ctx context.Context, in *pb.LookupNameRequest, opts ...grpc.CallOption) (*pb.LookupNameResponse, error) {
	caller, err := gocontract.NewNameRegistryCaller(b.nameRegistryContractAddress, b.client)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: ctx}

	key, err := caller.GetKey(callOpts, strings.ToLower(in.GetName()))
	if err != nil {
		return nil, err
	}
	if len(key) != 33 {
		return nil, status.Error(codes.NotFound, "name was not found")
	}
	return &pb.LookupNameResponse{
		Key: key,
	}, nil
}

func (b *Blockchain) LookupAddress(ctx context.Context, in *pb.LookupAddressRequest, opts ...grpc.CallOption) (*pb.LookupAddressResponse, error) {
	caller, err := gocontract.NewConnectorRegistryCaller(b.connectorRegistryContractAddress, b.client)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: ctx}

	location, err := caller.GetLocation(callOpts, strings.ToLower(in.GetName()), in.GetProtocol().String())
	if err != nil {
		return nil, err
	}

	if location == "" {
		return nil, status.Error(codes.NotFound, "address was not found")
	}

	return &pb.LookupAddressResponse{
		Address: location,
	}, nil
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
