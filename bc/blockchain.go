package bc

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc/v2"
	cnt "github.com/regnull/ubchain/gocontract"
)

var ErrNotFound = fmt.Errorf("not found")

var zeroAddress = common.BigToAddress(big.NewInt(0))

type Blockchain interface {
	PublicKey(ctx context.Context, name string) (*easyecc.PublicKey, error)
	Endpoint(ctx context.Context, name string) (string, error)
	PublicKeyP256(ctx context.Context, name string) (*easyecc.PublicKey, error)
	PublicKeyByCurve(ctx context.Context, name string,
		curve easyecc.EllipticCurve) (*easyecc.PublicKey, error)
}
type blockchainImpl struct {
	caller          NameRegistryCaller
	contractAddress string
}

func NewBlockchain(url string, contractAddress string) (Blockchain, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain node: %w", err)
	}
	caller, err := cnt.NewNameRegistryCaller(common.HexToAddress(contractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract instance")
	}
	return &blockchainImpl{
		caller:          caller,
		contractAddress: contractAddress}, nil
}

func (b *blockchainImpl) PublicKey(ctx context.Context, name string) (*easyecc.PublicKey, error) {
	res, err := b.caller.LookupName(&bind.CallOpts{Context: ctx}, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query the key")
	}

	if bytes.Equal(res.Owner.Bytes(), zeroAddress.Bytes()) {
		return nil, ErrNotFound
	}

	return easyecc.NewPublicKeyFromCompressedBytes(easyecc.SECP256K1, res.PublicKey)
}

func (b *blockchainImpl) getConfig(ctx context.Context, name string, configName string) (string, error) {
	location, err := b.caller.LookupConfig(&bind.CallOpts{Context: ctx}, name, configName)
	if err != nil {
		return "", fmt.Errorf("failed to query config")
	}

	if location == "" {
		return "", ErrNotFound
	}
	return location, nil
}

func (b *blockchainImpl) Endpoint(ctx context.Context, name string) (string, error) {
	return b.getConfig(ctx, name, "dms-endpoint")
}

func (b *blockchainImpl) PublicKeyP256(ctx context.Context, name string) (*easyecc.PublicKey, error) {
	keyStr, err := b.getConfig(ctx, name, "pubkey-p256")
	if err != nil {
		return nil, err
	}

	keyStr = strings.TrimPrefix(keyStr, "0x")
	keyBytes, err := hex.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	key, err := easyecc.NewPublicKeyFromCompressedBytes(easyecc.P256, keyBytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (b *blockchainImpl) PublicKeyByCurve(ctx context.Context, name string,
	curve easyecc.EllipticCurve) (*easyecc.PublicKey, error) {
	if curve == easyecc.SECP256K1 {
		return b.PublicKey(ctx, name)
	} else if curve == easyecc.P256 {
		return b.PublicKeyP256(ctx, name)
	}
	return nil, fmt.Errorf("unsupported curve")
}
