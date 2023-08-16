package bc

import (
	"context"
	"math/big"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/regnull/easyecc/v2"
	"github.com/regnull/ubikom/bc/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Blockchain_PublicKey(t *testing.T) {
	assert := assert.New(t)

	caller := new(mocks.MockNameRegistryCaller)
	bchain := &Blockchain{
		caller: caller,
	}

	privateKey, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)
	publicKey := privateKey.PublicKey()
	address, err := publicKey.BitcoinAddress()
	assert.NoError(err)
	addressBytes := base58.Decode(address)
	addr := common.BytesToAddress(addressBytes)

	caller.EXPECT().LookupName(mock.Anything, "foo").Return(struct {
		Owner     common.Address
		PublicKey []byte
		Price     *big.Int
	}{
		Owner:     addr,
		PublicKey: publicKey.CompressedBytes(),
		Price:     big.NewInt(0),
	}, nil)

	caller.EXPECT().LookupName(mock.Anything, "bar").Return(struct {
		Owner     common.Address
		PublicKey []byte
		Price     *big.Int
	}{
		Owner:     zeroAddress,
		PublicKey: []byte{},
		Price:     big.NewInt(0),
	}, nil)

	ctx := context.Background()
	publicKey1, err := bchain.PublicKey(ctx, "foo")
	assert.NoError(err)
	assert.NotNil(publicKey1)

	_, err = bchain.PublicKey(ctx, "bar")
	assert.Equal(ErrNotFound, err)

	caller.AssertExpectations(t)
}
