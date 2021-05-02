package server

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/protoutil"
	"teralyt.com/ubikom/store"
)

type DumpServer struct {
	pb.UnimplementedDMSDumpServiceServer

	baseDir      string
	lookupClient pb.LookupServiceClient
	store        store.Store
}

func NewDumpServer(baseDir string) *DumpServer {
	return &DumpServer{baseDir: baseDir}
}

func (s *DumpServer) Send(ctx context.Context, req *pb.DMSMessage) (*pb.Result, error) {
	// Get the public key associated with the sender's and receiver's name.
	senderKey, resErr := getKeyByName(ctx, s.lookupClient, req.Sender)
	if resErr != nil {
		return resErr, nil
	}

	receiverKey, resErr := getKeyByName(ctx, s.lookupClient, req.Receiver)
	if resErr != nil {
		return resErr, nil
	}

	// Verify signature.
	if !protoutil.VerifySignature(req.Signature, senderKey, req.Content) {
		log.Warn().Msg("signature verification failed")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	err := s.store.Save(req, receiverKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to save message")
		return &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}

	// Save message.
	return &pb.Result{Result: pb.ResultCode_RC_OK}, nil
}

func (s *DumpServer) Receive(ctx context.Context, req *pb.Signed) (*pb.ResultWithContent, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Receive not implemented")
}

func getKeyByName(ctx context.Context, lookupClient pb.LookupServiceClient, name string) ([]byte, *pb.Result) {
	res, err := lookupClient.LookupName(ctx, &pb.LookupNameRequest{
		Name: name})
	if err != nil {
		log.Error().Err(err).Msg("failed to lookup name")
		return nil, &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}
	}
	if res.Result == pb.ResultCode_RC_RECORD_NOT_FOUND {
		log.Error().Str("name", name).Msg("name record not found")
		return nil, &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Error().Str("result", res.Result.String()).Msg("unexpected error when retrieving name")
		return nil, &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}
	}
	return res.Key, nil
}
