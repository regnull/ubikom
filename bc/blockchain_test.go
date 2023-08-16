package bc

import (
	"context"
	"fmt"
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
	bchain := &blockchainImpl{
		caller: caller,
	}

	privateKey, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)
	publicKey := privateKey.PublicKey()
	address, err := publicKey.BitcoinAddress()
	assert.NoError(err)
	addressBytes := base58.Decode(address)
	addr := common.BytesToAddress(addressBytes)

	// First call - return a valid address and key.
	caller.EXPECT().LookupName(mock.Anything, "foo").Return(struct {
		Owner     common.Address
		PublicKey []byte
		Price     *big.Int
	}{
		Owner:     addr,
		PublicKey: publicKey.CompressedBytes(),
		Price:     big.NewInt(0),
	}, nil)

	// Second call - return a zero address.
	caller.EXPECT().LookupName(mock.Anything, "bar").Return(struct {
		Owner     common.Address
		PublicKey []byte
		Price     *big.Int
	}{
		Owner:     zeroAddress,
		PublicKey: []byte{},
		Price:     big.NewInt(0),
	}, nil)

	// Third call - return an error.
	caller.EXPECT().LookupName(mock.Anything, "baz").Return(struct {
		Owner     common.Address
		PublicKey []byte
		Price     *big.Int
	}{}, fmt.Errorf("some error"))

	ctx := context.Background()
	publicKey1, err := bchain.PublicKey(ctx, "foo")
	assert.NoError(err)
	assert.NotNil(publicKey1)

	_, err = bchain.PublicKey(ctx, "bar")
	assert.Equal(ErrNotFound, err)

	_, err = bchain.PublicKey(ctx, "baz")
	assert.Error(err)

	caller.AssertExpectations(t)
}

func Test_Blockchain_Endpoint(t *testing.T) {
	assert := assert.New(t)

	caller := new(mocks.MockNameRegistryCaller)
	bchain := &blockchainImpl{
		caller: caller,
	}

	caller.EXPECT().LookupConfig(mock.Anything, "foo", "dms-endpoint").Return("some-endpoint", nil)

	ctx := context.Background()
	endpoint, err := bchain.Endpoint(ctx, "foo")
	assert.NoError(err)
	assert.Equal("some-endpoint", endpoint)

	caller.EXPECT().LookupConfig(mock.Anything, "bar", "dms-endpoint").Return("", nil)
	_, err = bchain.Endpoint(ctx, "bar")
	assert.Equal(ErrNotFound, err)

	caller.EXPECT().LookupConfig(mock.Anything, "baz", "dms-endpoint").Return("", fmt.Errorf("some error"))
	_, err = bchain.Endpoint(ctx, "baz")
	assert.Error(err)

	caller.AssertExpectations(t)
}

func Test_Blockchain_PublicKeyP256(t *testing.T) {
	assert := assert.New(t)

	caller := new(mocks.MockNameRegistryCaller)
	bchain := &blockchainImpl{
		caller: caller,
	}
	ctx := context.Background()

	privateKey, err := easyecc.NewPrivateKey(easyecc.P256)
	assert.NoError(err)
	publicKey := privateKey.PublicKey()

	caller.EXPECT().LookupConfig(mock.Anything, "foo", "pubkey-p256").
		Return(fmt.Sprintf("%x", publicKey.CompressedBytes()), nil)
	publicKey1, err := bchain.PublicKeyP256(ctx, "foo")
	assert.NoError(err)
	assert.True(publicKey.Equal(publicKey1))

	// Intentionally use Bytes() instead of CompressedBytes() to cause an error.
	caller.EXPECT().LookupConfig(mock.Anything, "bar", "pubkey-p256").
		Return(fmt.Sprintf("%x", publicKey.Bytes()), nil)
	publicKey1, err = bchain.PublicKeyP256(ctx, "bar")
	assert.Error(err)
	assert.Nil(publicKey1)

	// Return an incorrect hex value.
	caller.EXPECT().LookupConfig(mock.Anything, "baz", "pubkey-p256").
		Return("$$$ this is an incorrect hex string", nil)
	publicKey1, err = bchain.PublicKeyP256(ctx, "baz")
	assert.Error(err)
	assert.Nil(publicKey1)

	// Return an error.
	caller.EXPECT().LookupConfig(mock.Anything, "xyz", "pubkey-p256").
		Return("", fmt.Errorf("some error"))
	publicKey1, err = bchain.PublicKeyP256(ctx, "xyz")
	assert.Error(err)
	assert.Nil(publicKey1)

	caller.AssertExpectations(t)
}

func Test_Blockchain_PublicKeyByCurve(t *testing.T) {
	assert := assert.New(t)
	caller := new(mocks.MockNameRegistryCaller)
	bchain := &blockchainImpl{
		caller: caller,
	}
	ctx := context.Background()

	privateKeySecp256k1, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)
	publicKeySecp256k1 := privateKeySecp256k1.PublicKey()

	privateKeyP256, err := easyecc.NewPrivateKey(easyecc.P256)
	assert.NoError(err)
	publicKeyP256 := privateKeyP256.PublicKey()

	address, err := publicKeySecp256k1.BitcoinAddress()
	assert.NoError(err)
	addressBytes := base58.Decode(address)
	addr := common.BytesToAddress(addressBytes)

	caller.EXPECT().LookupName(mock.Anything, "foo").Return(struct {
		Owner     common.Address
		PublicKey []byte
		Price     *big.Int
	}{
		Owner:     addr,
		PublicKey: publicKeySecp256k1.CompressedBytes(),
		Price:     big.NewInt(0),
	}, nil)

	caller.EXPECT().LookupConfig(mock.Anything, "foo", "pubkey-p256").
		Return(fmt.Sprintf("%x", publicKeyP256.CompressedBytes()), nil)

	publicKey1, err := bchain.PublicKeyByCurve(ctx, "foo", easyecc.SECP256K1)
	assert.NoError(err)
	assert.True(publicKeySecp256k1.Equal(publicKey1))

	publicKey2, err := bchain.PublicKeyByCurve(ctx, "foo", easyecc.P256)
	assert.NoError(err)
	assert.True(publicKeyP256.Equal(publicKey2))

	// Try unsupported curve.
	publicKey3, err := bchain.PublicKeyByCurve(ctx, "foo", easyecc.P521)
	assert.Error(err)
	assert.Nil(publicKey3)
}
