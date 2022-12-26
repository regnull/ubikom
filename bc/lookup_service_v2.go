package bc

import (
	"context"
	"fmt"

	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type LookupServiceV2 struct {
	main     pb.LookupServiceClient
	fallback pb.LookupServiceClient
}

func NewLookupServiceV2(main, fallback pb.LookupServiceClient) *LookupServiceV2 {
	return &LookupServiceV2{main: main, fallback: fallback}
}

func (b *LookupServiceV2) LookupKey(ctx context.Context, in *pb.LookupKeyRequest, opts ...grpc.CallOption) (*pb.LookupKeyResponse, error) {
	// Blockchain V2 does not implement this, so we don't allow this.
	return nil, fmt.Errorf("not implemented")
}

func (b *LookupServiceV2) LookupName(ctx context.Context, in *pb.LookupNameRequest, opts ...grpc.CallOption) (*pb.LookupNameResponse, error) {
	res, err := b.main.LookupName(ctx, in, opts...)
	if err != nil && util.StatusCodeFromError(err) == codes.NotFound {
		log.Warn().Str("name", in.GetName()).Msg("using fallback lookup service to lookup name")
		return b.fallback.LookupName(ctx, in, opts...)
	}
	return res, err
}

func (b *LookupServiceV2) LookupAddress(ctx context.Context, in *pb.LookupAddressRequest, opts ...grpc.CallOption) (*pb.LookupAddressResponse, error) {
	res, err := b.main.LookupAddress(ctx, in, opts...)
	if err != nil && util.StatusCodeFromError(err) == codes.NotFound {
		log.Warn().Str("name", in.GetName()).Msg("using fallback lookup service to lookup address")
		return b.fallback.LookupAddress(ctx, in, opts...)
	}
	return res, err
}
