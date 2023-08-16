package bc

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type NameRegistryCaller interface {
	LookupName(opts *bind.CallOpts, name string) (struct {
		Owner     common.Address
		PublicKey []byte
		Price     *big.Int
	}, error)
	LookupConfig(opts *bind.CallOpts, name string, configName string) (string, error)
}
