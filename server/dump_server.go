package server

import (
	"context"
	"time"

	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/store"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DumpServer struct {
	pb.UnimplementedDMSDumpServiceServer

	baseDir      string
	lookupClient pb.LookupServiceClient
	store        store.Store
}

func NewDumpServer(baseDir string, lookupClient pb.LookupServiceClient, maxMessageAgeHours int) (*DumpServer, error) {
	store, err := store.NewBadger(baseDir, time.Duration(maxMessageAgeHours)*time.Hour)
	if err != nil {
		return nil, err
	}
	return &DumpServer{
		baseDir:      baseDir,
		lookupClient: lookupClient,
		store:        store}, nil
}

func (s *DumpServer) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	log.Debug().Msg("got send request")
	// Get the public key associated with the sender's and receiver's name.
	senderKey, resErr := getKeyByName(ctx, s.lookupClient, req.GetMessage().GetSender())
	if resErr != nil {
		return nil, status.Error(codes.Internal, "failed to lookup name")
	}

	receiverKey, resErr := getKeyByName(ctx, s.lookupClient, req.GetMessage().GetReceiver())
	if resErr != nil {
		return nil, status.Error(codes.Internal, "failed to lookup name")
	}

	// Verify signature.
	if !protoutil.VerifySignature(req.GetMessage().GetSignature(), senderKey, req.GetMessage().GetContent()) {
		log.Warn().Msg("signature verification failed")
		return nil, status.Error(codes.InvalidArgument, "bad signature")
	}

	err := s.store.Save(req.GetMessage(), receiverKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to save message")
		return nil, status.Error(codes.Internal, "message store error")
	}

	// Save message.
	return &pb.SendResponse{}, nil
}

func (s *DumpServer) Receive(ctx context.Context, req *pb.ReceiveRequest) (*pb.ReceiveResponse, error) {
	log.Debug().Msg("got receive request")
	if !protoutil.VerifySignature(req.GetIdentityProof().GetSignature(), req.GetIdentityProof().GetKey(),
		req.GetIdentityProof().GetContent()) {
		log.Warn().Msg("signature verification failed")
		return nil, status.Error(codes.InvalidArgument, "bad signature")
	}

	msg, err := s.store.GetNext(req.GetIdentityProof().GetKey())
	if err != nil {
		log.Error().Err(err).Msg("failed to get next message")
		return nil, status.Error(codes.Internal, "message store error")
	}

	if msg == nil {
		return nil, status.Error(codes.NotFound, "not found")
	}

	err = s.store.Remove(msg, req.GetIdentityProof().GetKey())
	if err != nil {
		log.Error().Err(err).Msg("failed to remove message")
	}

	return &pb.ReceiveResponse{Message: msg}, nil
}

func getKeyByName(ctx context.Context, lookupClient pb.LookupServiceClient, name string) ([]byte, error) {
	res, err := lookupClient.LookupName(ctx, &pb.LookupNameRequest{
		Name: name})
	if err != nil {
		log.Error().Err(err).Msg("failed to lookup name")
		return nil, err
	}
	return res.Key, nil
}
