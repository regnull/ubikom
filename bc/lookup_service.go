package bc

import (
	"bytes"
	"context"
	"sync"

	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type LookupServiceClient struct {
	bcLookup     pb.LookupServiceClient
	legacyLookup pb.LookupServiceClient
}

func NewLookupServiceClient(bcLookup, legacyLookup pb.LookupServiceClient) *LookupServiceClient {
	return &LookupServiceClient{
		bcLookup:     bcLookup,
		legacyLookup: legacyLookup,
	}
}

func (c *LookupServiceClient) LookupKey(ctx context.Context, in *pb.LookupKeyRequest, opts ...grpc.CallOption) (*pb.LookupKeyResponse, error) {
	var legacyRes *pb.LookupKeyResponse
	var legacyErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		legacyRes, legacyErr = c.legacyLookup.LookupKey(ctx, in, opts...)
		wg.Done()
	}()

	var bcRes *pb.LookupKeyResponse
	var bcErr error
	go func() {
		bcRes, bcErr = c.bcLookup.LookupKey(ctx, in, opts...)
		wg.Done()
	}()

	if legacyErr != nil && bcErr != nil {
		return nil, legacyErr
	}

	if legacyErr != nil || bcErr != nil {
		log.Error().Msg("lookup error mismatch")
		if legacyErr != nil {
			return nil, legacyErr
		}
	}

	_ = bcRes

	return legacyRes, nil
}

func (c *LookupServiceClient) LookupName(ctx context.Context, in *pb.LookupNameRequest, opts ...grpc.CallOption) (*pb.LookupNameResponse, error) {
	var legacyRes *pb.LookupNameResponse
	var legacyErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		legacyRes, legacyErr = c.legacyLookup.LookupName(ctx, in, opts...)
		wg.Done()
	}()

	var bcRes *pb.LookupNameResponse
	var bcErr error
	go func() {
		bcRes, bcErr = c.bcLookup.LookupName(ctx, in, opts...)
		wg.Done()
	}()

	if legacyErr != nil && bcErr != nil {
		return nil, legacyErr
	}

	if legacyErr != nil || bcErr != nil {
		log.Error().Msg("lookup error mismatch")
		if legacyErr != nil {
			return nil, legacyErr
		}
	}

	if !bytes.Equal(legacyRes.GetKey(), bcRes.GetKey()) {
		log.Error().Msg("lookup key mismatch")
	}

	return legacyRes, nil
}

func (c *LookupServiceClient) LookupAddress(ctx context.Context, in *pb.LookupAddressRequest, opts ...grpc.CallOption) (*pb.LookupAddressResponse, error) {
	var legacyRes *pb.LookupAddressResponse
	var legacyErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		legacyRes, legacyErr = c.legacyLookup.LookupAddress(ctx, in, opts...)
		wg.Done()
	}()

	var bcRes *pb.LookupAddressResponse
	var bcErr error
	go func() {
		bcRes, bcErr = c.bcLookup.LookupAddress(ctx, in, opts...)
		wg.Done()
	}()

	if legacyErr != nil && bcErr != nil {
		return nil, legacyErr
	}

	if legacyErr != nil || bcErr != nil {
		log.Error().Msg("lookup error mismatch")
		if legacyErr != nil {
			return nil, legacyErr
		}
	}

	if legacyRes.GetAddress() != bcRes.GetAddress() {
		log.Error().Msg("lookup address mismatch")
	}

	return legacyRes, nil
}