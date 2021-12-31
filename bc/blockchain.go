package bc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubchain/keyregistry"
	"github.com/rs/zerolog/log"
)

// Blockchain represents Ethereum blockchain.
type Blockchain struct {
	client          *ethclient.Client
	contractAddress common.Address
	privateKey      *easyecc.PrivateKey
}

// NewBlockchain returns a new blockchain.
func NewBlockchain(client *ethclient.Client, contractAddress string, privateKey *easyecc.PrivateKey) *Blockchain {
	return &Blockchain{
		client:          client,
		contractAddress: common.HexToAddress(contractAddress),
		privateKey:      privateKey}
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
	gasLimit := uint64(300000)

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

	instance, err := keyregistry.NewKeyregistry(b.contractAddress, b.client)
	if err != nil {
		return "", err
	}

	tx, err := instance.Register(auth, key.SerializeCompressed())
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}
