package bc

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type LookupServiceClient struct {
	bcLookup     pb.LookupServiceClient
	legacyLookup pb.LookupServiceClient
	useLegacy    bool
}

func NewLookupServiceClient(bcLookup, legacyLookup pb.LookupServiceClient, useLegacy bool) *LookupServiceClient {
	return &LookupServiceClient{
		bcLookup:     bcLookup,
		legacyLookup: legacyLookup,
		useLegacy:    useLegacy,
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

	if legacyErr != bcErr {
		log.Error().Str("key", fmt.Sprintf("%0x", in.GetKey())).Msg("lookup key error mismatch")
		if legacyErr != nil {
			log.Error().Err(legacyErr).Msg("legacy error")
		}
		if bcErr != nil {
			log.Error().Err(bcErr).Msg("bc error")
		}
	}

	var primaryRes, secondaryRes *pb.LookupKeyResponse
	var primaryErr, secondaryErr error

	if c.useLegacy {
		primaryRes = legacyRes
		primaryErr = legacyErr
		secondaryRes = bcRes
		secondaryErr = bcErr
	} else {
		secondaryRes = legacyRes
		secondaryErr = legacyErr
		primaryRes = bcRes
		primaryErr = bcErr
	}

	if primaryErr != nil && secondaryErr == nil {
		return secondaryRes, secondaryErr
	}

	return primaryRes, primaryErr
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

	var bcRes *pb.LookupNameResponse
	var bcErr error
	go func() {
		defer wg.Done()
		ctx1, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		bcRes, bcErr = c.bcLookup.LookupName(ctx1, in, opts...)
	}()

	wg.Wait()

	if legacyErr != bcErr {
		log.Error().Str("name", in.GetName()).Msg("lookup name error mismatch")
		if legacyErr != nil {
			log.Error().Err(legacyErr).Msg("legacy error")
		}
		if bcErr != nil {
			log.Error().Err(bcErr).Msg("bc error")
		}
	}

	if !bytes.Equal(legacyRes.GetKey(), bcRes.GetKey()) {
		log.Error().Msg("lookup key mismatch")
	}

	var primaryRes, secondaryRes *pb.LookupNameResponse
	var primaryErr, secondaryErr error

	if c.useLegacy {
		primaryRes = legacyRes
		primaryErr = legacyErr
		secondaryRes = bcRes
		secondaryErr = bcErr
	} else {
		secondaryRes = legacyRes
		secondaryErr = legacyErr
		primaryRes = bcRes
		primaryErr = bcErr
	}

	if primaryErr != nil && secondaryErr == nil {
		return secondaryRes, secondaryErr
	}

	return primaryRes, primaryErr
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

	if legacyErr != bcErr {
		log.Error().Str("name", in.GetName()).Str("protocol", in.GetProtocol().String()).Msg("lookup address error mismatch")
		if legacyErr != nil {
			log.Error().Err(legacyErr).Msg("legacy error")
		}
		if bcErr != nil {
			log.Error().Err(bcErr).Msg("bc error")
		}
	}

	if legacyRes.GetAddress() != bcRes.GetAddress() {
		log.Error().Msg("lookup address mismatch")
	}

	var primaryRes, secondaryRes *pb.LookupAddressResponse
	var primaryErr, secondaryErr error

	if c.useLegacy {
		primaryRes = legacyRes
		primaryErr = legacyErr
		secondaryRes = bcRes
		secondaryErr = bcErr
	} else {
		secondaryRes = legacyRes
		secondaryErr = legacyErr
		primaryRes = bcRes
		primaryErr = bcErr
	}

	if primaryErr != nil && secondaryErr == nil {
		return secondaryRes, secondaryErr
	}

	return primaryRes, primaryErr
}
