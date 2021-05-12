package server

import (
	"context"

	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/store"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type DumpServer struct {
	pb.UnimplementedDMSDumpServiceServer

	baseDir      string
	lookupClient pb.LookupServiceClient
	store        store.Store
}

func NewDumpServer(baseDir string, lookupClient pb.LookupServiceClient) *DumpServer {
	return &DumpServer{
		baseDir:      baseDir,
		lookupClient: lookupClient,
		store:        store.NewFile(baseDir)}
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
	if !protoutil.VerifySignature(req.Signature, req.Key, req.Content) {
		log.Warn().Msg("signature verification failed")
		return &pb.ResultWithContent{Result: &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}}, nil
	}

	msg, err := s.store.GetNext(req.GetKey())
	if err != nil {
		log.Error().Err(err).Msg("failed to get next message")
		return &pb.ResultWithContent{Result: &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}}, nil
	}

	if msg == nil {
		return &pb.ResultWithContent{Result: &pb.Result{Result: pb.ResultCode_RC_RECORD_NOT_FOUND}}, nil
	}

	b, err := proto.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to serialize the message")
		return &pb.ResultWithContent{Result: &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}}, nil
	}

	err = s.store.Remove(msg, req.GetKey())
	if err != nil {
		log.Error().Err(err).Msg("failed to remove message")
	}

	return &pb.ResultWithContent{
		Result:  &pb.Result{Result: pb.ResultCode_RC_OK},
		Content: b,
	}, nil
}

func getKeyByName(ctx context.Context, lookupClient pb.LookupServiceClient, name string) ([]byte, *pb.Result) {
	res, err := lookupClient.LookupName(ctx, &pb.LookupNameRequest{
		Name: name})
	if err != nil {
		log.Error().Err(err).Msg("failed to lookup name")
		return nil, &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}
	}
	if res.GetResult().GetResult() == pb.ResultCode_RC_RECORD_NOT_FOUND {
		log.Error().Str("name", name).Msg("name record not found")
		return nil, &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}
	}
	if res.GetResult().GetResult() != pb.ResultCode_RC_OK {
		log.Error().Str("result", res.Result.String()).Msg("unexpected error when retrieving name")
		return nil, &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}
	}
	return res.Key, nil
}
