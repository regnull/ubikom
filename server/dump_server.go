package server

import (
	"context"
	"time"

	"github.com/regnull/easyecc/v2"
	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/store"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxAllowedIdentitySignatureDifferenceSeconds = 10000000000.0

type DumpServer struct {
	pb.UnimplementedDMSDumpServiceServer

	baseDir string
	bchain  bc.Blockchain
	store   store.Store
}

func NewDumpServer(baseDir string, bchain bc.Blockchain,
	maxMessageAgeHours int) (*DumpServer, error) {
	store, err := store.NewBadger(baseDir, time.Duration(maxMessageAgeHours)*time.Hour)
	if err != nil {
		return nil, err
	}
	return &DumpServer{
		baseDir: baseDir,
		bchain:  bchain,
		store:   store}, nil
}

func (s *DumpServer) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	log.Debug().Msg("got send request")
	protoCurve := req.GetMessage().GetCryptoContext().GetEllipticCurve()
	curve := protoutil.CurveFromProto(protoCurve)
	if curve == easyecc.INVALID_CURVE {
		return nil, status.Error(codes.Internal, "invalid curve")
	}
	// Get the public key associated with the sender's and receiver's name.
	senderKey, resErr := s.bchain.PublicKeyByCurve(ctx, req.GetMessage().GetSender(), curve)
	if resErr != nil {
		return nil, status.Error(codes.Internal, "failed to lookup name")
	}

	receiverKey, resErr := s.bchain.PublicKeyByCurve(ctx, req.GetMessage().GetReceiver(), curve)
	if resErr != nil {
		return nil, status.Error(codes.Internal, "failed to lookup name")
	}

	// Verify signature.
	if !protoutil.VerifySignature(req.GetMessage().GetSignature(),
		senderKey, req.GetMessage().GetContent()) {
		log.Warn().Msg("signature verification failed")
		return nil, status.Error(codes.InvalidArgument, "bad signature")
	}

	err := s.store.Save(req.GetMessage(), receiverKey.CompressedBytes())
	if err != nil {
		log.Error().Err(err).Msg("failed to save message")
		return nil, status.Error(codes.Internal, "message store error")
	}

	// Save message.
	return &pb.SendResponse{}, nil
}

func (s *DumpServer) Receive(ctx context.Context, req *pb.ReceiveRequest) (*pb.ReceiveResponse, error) {
	log.Debug().Msg("got receive request")
	curve := easyecc.SECP256K1
	if req.GetCryptoContext() != nil {
		protoCurve := req.GetCryptoContext().GetEllipticCurve()
		curve = protoutil.CurveFromProto(protoCurve)
		if curve == easyecc.INVALID_CURVE {
			return nil, status.Error(codes.Internal, "invalid curve")
		}
	}

	key, err := easyecc.NewPublicKeyFromCompressedBytes(curve, req.GetIdentityProof().GetKey())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}
	err = protoutil.VerifyIdentity(req.GetIdentityProof(), time.Now(), 10.0, curve)
	if err != nil {
		log.Debug().Err(err).Msg("identity verification failed, using fallback")
		// return nil, status.Error(codes.InvalidArgument, "bad signature")
		// For now, we fallback to the old verification algorithm. To be removed later.
		// TODO: remove this once all the clients are migrated.
		if !protoutil.VerifySignature(req.GetIdentityProof().GetSignature(), key,
			req.GetIdentityProof().GetContent()) {
			log.Warn().Msg("signature verification failed")
			return nil, status.Error(codes.InvalidArgument, "bad signature")
		}
		log.Debug().Msg("signature verification succeeded")
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
