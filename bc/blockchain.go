package bc

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	cntv2 "github.com/regnull/ubchain/gocontract"
	"github.com/regnull/ubikom/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Blockchain struct {
	client          *ethclient.Client
	contractAddress string
}

func NewBlockchain(url string, contractAddress string) (*Blockchain, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain node: %w", err)
	}
	return &Blockchain{client: client, contractAddress: contractAddress}, nil
}

func (b *Blockchain) LookupKey(ctx context.Context, in *pb.LookupKeyRequest, opts ...grpc.CallOption) (*pb.LookupKeyResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (b *Blockchain) LookupName(ctx context.Context, in *pb.LookupNameRequest, opts ...grpc.CallOption) (*pb.LookupNameResponse, error) {
	instance, err := cntv2.NewNameRegistryCaller(common.HexToAddress(b.contractAddress), b.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract instance")
	}

	res, err := instance.LookupName(nil, in.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to query the key")
	}

	zeroAddress := common.BigToAddress(big.NewInt(0))
	if bytes.Equal(res.Owner.Bytes(), zeroAddress.Bytes()) {
		return nil, status.Error(codes.NotFound, "name was not found")
	}

	return &pb.LookupNameResponse{
		Key: res.PublicKey,
	}, nil
}

func (b *Blockchain) LookupAddress(ctx context.Context, in *pb.LookupAddressRequest, opts ...grpc.CallOption) (*pb.LookupAddressResponse, error) {
	instance, err := cntv2.NewNameRegistryCaller(common.HexToAddress(b.contractAddress), b.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract instance")
	}

	location, err := instance.LookupConfig(nil, in.GetName(), "dms-endpoint")
	if err != nil {
		return nil, fmt.Errorf("failed to query the key")
	}

	if location == "" {
		return nil, status.Error(codes.NotFound, "address was not found")
	}

	return &pb.LookupAddressResponse{
		Address: location,
	}, nil
}
