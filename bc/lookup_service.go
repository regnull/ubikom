package bc

import (
	"bytes"
	"context"
	"sync"
	"time"

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
		defer wg.Done()
		legacyRes, legacyErr = c.legacyLookup.LookupKey(ctx, in, opts...)
	}()

	var bcRes *pb.LookupKeyResponse
	var bcErr error
	go func() {
		defer wg.Done()
		ctx1, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		bcRes, bcErr = c.bcLookup.LookupKey(ctx1, in, opts...)
	}()

	wg.Wait()

	if legacyErr != nil && bcErr != nil {
		return nil, legacyErr
	}

	if legacyErr != nil || bcErr != nil {
		log.Error().Msg("lookup key error mismatch")
		if legacyErr != nil {
			log.Error().Err(legacyErr).Msg("legacy error")
		}
		if bcErr != nil {
			log.Error().Err(bcErr).Msg("bc error")
		}
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
		defer wg.Done()
		legacyRes, legacyErr = c.legacyLookup.LookupName(ctx, in, opts...)
	}()

	wg.Wait()

	var bcRes *pb.LookupNameResponse
	var bcErr error
	go func() {
		defer wg.Done()
		ctx1, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		bcRes, bcErr = c.bcLookup.LookupName(ctx1, in, opts...)
	}()

	if legacyErr != nil && bcErr != nil {
		return nil, legacyErr
	}

	if legacyErr != nil || bcErr != nil {
		log.Error().Str("name", in.GetName()).Msg("lookup name error mismatch")
		if legacyErr != nil {
			log.Error().Err(legacyErr).Msg("legacy error")
		}
		if bcErr != nil {
			log.Error().Err(bcErr).Msg("bc error")
		}
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
		defer wg.Done()
		legacyRes, legacyErr = c.legacyLookup.LookupAddress(ctx, in, opts...)
	}()

	var bcRes *pb.LookupAddressResponse
	var bcErr error
	go func() {
		defer wg.Done()
		ctx1, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		bcRes, bcErr = c.bcLookup.LookupAddress(ctx1, in, opts...)
	}()

	wg.Wait()

	if legacyErr != nil && bcErr != nil {
		return nil, legacyErr
	}

	if legacyErr != nil || bcErr != nil {
		log.Error().Str("name", in.GetName()).Str("protocol", in.GetProtocol().String()).Msg("lookup address error mismatch")
		if legacyErr != nil {
			log.Error().Err(legacyErr).Msg("legacy error")
		}
		if bcErr != nil {
			log.Error().Err(bcErr).Msg("bc error")
		}
		if legacyErr != nil {
			return nil, legacyErr
		}
	}

	if legacyRes.GetAddress() != bcRes.GetAddress() {
		log.Error().Msg("lookup address mismatch")
	}

	return legacyRes, nil
}
