package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dgraph-io/badger/v3"
	"github.com/golang/protobuf/proto"
	"github.com/regnull/easyecc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/regnull/ubikom/db"
	"github.com/regnull/ubikom/event"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/pow"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
)

var ErrKeyExists = fmt.Errorf("key already exists")
var ErrNotAuthorized = fmt.Errorf("not authorized")
var ErrNotFound = fmt.Errorf("not found")

type Server struct {
	pb.UnimplementedIdentityServiceServer
	pb.UnimplementedLookupServiceServer

	dbi         *db.BadgerDB
	powStrength int
	privateKey  *easyecc.PrivateKey
	ubikomName  string
	eventSender *event.Sender
}

func NewServer(d *badger.DB, powStrength int, privateKey *easyecc.PrivateKey,
	ubikomName string, eventTarget string) *Server {

	s := &Server{
		dbi:         db.NewBadgerDB(d),
		powStrength: powStrength,
		privateKey:  privateKey,
		ubikomName:  ubikomName,
	}
	var eventSender *event.Sender
	if privateKey != nil && ubikomName != "" && eventTarget != "" {
		log.Debug().Msg("creating event sender")
		eventSender = event.NewSender(eventTarget, &loopbackLookupClient{s})
	} else {
		log.Warn().Msg("cannot create event sender")
	}
	s.eventSender = eventSender
	return s
}

func (s *Server) RegisterKey(ctx context.Context, req *pb.SignedWithPow) (*pb.KeyRegistrationResponse, error) {
	if !verifyPowAndSignature(req, s.powStrength) {
		log.Warn().Msg("insufficient POW or invalid signature")
		return nil, status.Error(codes.InvalidArgument, "invalid pow or signature")
	}

	keyRegistrationReq := &pb.KeyRegistrationRequest{}
	err := proto.Unmarshal(req.GetContent(), keyRegistrationReq)
	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal key registration request")
		return nil, status.Error(codes.InvalidArgument, "failed to marshal request")
	}

	if !bytes.Equal(keyRegistrationReq.GetKey(), req.GetKey()) {
		log.Warn().Hex("content", req.GetContent()).Hex("key", req.GetKey()).Msg("key and content do not match (key registration request)")
		return nil, status.Error(codes.InvalidArgument, "keys do not match")
	}

	key := keyRegistrationReq.GetKey()
	publicKey, err := easyecc.NewPublicFromSerializedCompressed(key)
	if err != nil {
		log.Warn().Err(err).Msg("failed to create public key from serialized compressed")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}
	publicKeyBase58 := base58.Encode(publicKey.SerializeCompressed())

	err = s.dbi.RegisterKey(publicKey)
	if err != nil {
		if err == db.ErrRecordExists {
			log.Warn().Str("key", publicKeyBase58).Msg("this key is already registered")
			return nil, status.Error(codes.AlreadyExists, "key is already registered")
		}
		log.Error().Err(err).Str("key", publicKeyBase58).Msg("error writing public key")
		return nil, status.Error(codes.Internal, "failed to register key")
	}

	if s.eventSender != nil {
		err = s.eventSender.KeyRegistered(ctx, s.privateKey, s.ubikomName,
			util.SerializedCompressedToAddress(keyRegistrationReq.GetKey()))
		if err != nil {
			log.Error().Err(err).Msg("error sending event")
		} else {
			log.Debug().Msg("key registered event sent")
		}
	}

	log.Info().Str("key", util.SerializedCompressedToAddress(keyRegistrationReq.GetKey())).Msg("key is registered successfully")
	return &pb.KeyRegistrationResponse{}, nil
}

func (s *Server) RegisterKeyRelationship(ctx context.Context, req *pb.SignedWithPow) (*pb.KeyRelationshipRegistrationResponse, error) {
	if !verifyPowAndSignature(req, s.powStrength) {
		log.Warn().Msg("insufficient POW or invalid signature")
		return nil, status.Error(codes.InvalidArgument, "invalid pow or signature")
	}

	keyRelReq := &pb.KeyRelationshipRegistrationRequest{}
	err := proto.Unmarshal(req.GetContent(), keyRelReq)
	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal key relationship registration request")
		return nil, status.Error(codes.InvalidArgument, "failed to unmarshal request")
	}

	if keyRelReq.GetRelationship() != pb.KeyRelationship_KR_PARENT {
		return nil, status.Error(codes.InvalidArgument, "invalid relationship")
	}

	parentKey, err := easyecc.NewPublicFromSerializedCompressed(req.GetKey())
	if err != nil {
		log.Warn().Err(err).Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}

	childKey, err := easyecc.NewPublicFromSerializedCompressed(keyRelReq.GetTargetKey())
	if err != nil {
		log.Warn().Err(err).Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}

	err = s.dbi.RegisterKeyParent(childKey, parentKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to register parent key")
		return nil, status.Error(codes.Internal, "failed to register key relationship")
	}
	log.Info().Str("key", util.SerializedCompressedToAddress(req.GetKey())).Msg("parent key registered successfully")
	return &pb.KeyRelationshipRegistrationResponse{}, nil
}

func (s *Server) DisableKey(ctx context.Context, req *pb.SignedWithPow) (*pb.KeyDisableResponse, error) {
	if !verifyPowAndSignature(req, s.powStrength) {
		log.Warn().Msg("insufficient POW or invalid signature")
		return nil, status.Error(codes.InvalidArgument, "invalid pow or signature")
	}

	keyDisReq := &pb.KeyDisableRequest{}
	err := proto.Unmarshal(req.GetContent(), keyDisReq)
	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal key disable request")
		return nil, status.Error(codes.InvalidArgument, "failed to unmarshal request")
	}

	log.Debug().Str("orig-key", base58.Encode(req.GetKey())).Str("target-key",
		base58.Encode(keyDisReq.GetKey())).Msg("disable key request")

	originator, err := easyecc.NewPublicFromSerializedCompressed(req.GetKey())
	if err != nil {
		log.Warn().Err(err).Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}

	key, err := easyecc.NewPublicFromSerializedCompressed(keyDisReq.GetKey())
	if err != nil {
		log.Warn().Err(err).Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}

	err = s.dbi.DisableKey(originator, key)

	if err != nil {
		log.Error().Err(err).Msg("failed to disable key")
		return nil, status.Error(codes.Internal, "failed to disable key")
	}

	log.Info().Str("key", util.SerializedCompressedToAddress(key.SerializeCompressed())).Msg("key disabled")

	return &pb.KeyDisableResponse{}, nil
}

func (s *Server) RegisterName(ctx context.Context, req *pb.SignedWithPow) (*pb.NameRegistrationResponse, error) {
	if !verifyPowAndSignature(req, s.powStrength) {
		log.Warn().Msg("insufficient POW or invalid signature")
		return nil, status.Error(codes.InvalidArgument, "invalid pow or signature")
	}

	originatorKey, err := easyecc.NewPublicFromSerializedCompressed(req.GetKey())
	if err != nil {
		log.Warn().Err(err).Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}

	nameRegistrationReq := &pb.NameRegistrationRequest{}
	err = proto.Unmarshal(req.GetContent(), nameRegistrationReq)
	if err != nil {
		log.Warn().Msg("invalid name registration request")
		return nil, status.Error(codes.InvalidArgument, "failed to unmarshal request")
	}

	targetKey, err := easyecc.NewPublicFromSerializedCompressed(nameRegistrationReq.GetKey())
	if err != nil {
		log.Warn().Err(err).Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}

	if !util.ValidateName(nameRegistrationReq.GetName()) {
		log.Warn().Str("name", nameRegistrationReq.GetName()).Msg("invalid name")
		return nil, status.Error(codes.InvalidArgument, "invalid name")
	}
	log.Info().Str("name", nameRegistrationReq.GetName()).Msg("registering name")

	err = s.dbi.RegisterName(originatorKey, targetKey, nameRegistrationReq.GetName())
	if err != nil {
		if err == db.ErrRecordExists {
			log.Warn().Str("name", nameRegistrationReq.GetName()).Msg("this name is already registered")
			return nil, status.Error(codes.AlreadyExists, "name is already registered")
		}
		if err == db.ErrNotAuthorized {
			log.Warn().Str("name", nameRegistrationReq.GetName()).Msg("key is not authorized")
			return nil, status.Error(codes.PermissionDenied, "key is not authorized")
		}
		log.Error().Str("name", nameRegistrationReq.GetName()).Err(err).Msg("error writing name")
		return nil, status.Error(codes.Internal, "db error")
	}
	log.Info().Str("name", nameRegistrationReq.GetName()).Msg("name registered successfully")
	return &pb.NameRegistrationResponse{}, nil
}

func (s *Server) RegisterAddress(ctx context.Context, req *pb.SignedWithPow) (*pb.AddressRegistrationResponse, error) {
	if !verifyPowAndSignature(req, s.powStrength) {
		log.Warn().Msg("insufficient POW or invalid signature")
		return nil, status.Error(codes.InvalidArgument, "invalid pow or signature")
	}

	originatorKey, err := easyecc.NewPublicFromSerializedCompressed(req.GetKey())
	if err != nil {
		log.Warn().Err(err).Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}

	addressRegistrationReq := &pb.AddressRegistrationRequest{}
	err = proto.Unmarshal(req.GetContent(), addressRegistrationReq)
	if err != nil {
		log.Warn().Msg("invalid address registration request")
		return nil, status.Error(codes.InvalidArgument, "failed to unmarshal request")
	}

	log.Info().Str("name", addressRegistrationReq.GetName()).
		Str("address", addressRegistrationReq.GetAddress()).Msg("registering address")

	targetKey, err := easyecc.NewPublicFromSerializedCompressed(addressRegistrationReq.GetKey())
	if err != nil {
		log.Warn().Err(err).Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")
	}

	err = s.dbi.RegisterAddress(originatorKey, targetKey, addressRegistrationReq.GetName(),
		addressRegistrationReq.GetProtocol(), addressRegistrationReq.GetAddress())

	if err != nil {
		log.Error().Str("name", addressRegistrationReq.GetName()).
			Str("address", addressRegistrationReq.GetAddress()).Err(err).Msg("error registering address")
		if err == db.ErrNotAuthorized {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		return nil, status.Error(codes.Internal, "db error")
	}

	log.Info().Str("name", addressRegistrationReq.GetName()).
		Str("address", addressRegistrationReq.GetAddress()).Msg("address registered successfully")
	return &pb.AddressRegistrationResponse{}, nil
}

func (s *Server) LookupKey(ctx context.Context, req *pb.LookupKeyRequest) (*pb.LookupKeyResponse, error) {
	key, err := easyecc.NewPublicFromSerializedCompressed(req.GetKey())
	if err != nil {
		log.Warn().Msg("invalid key")
		return nil, status.Error(codes.InvalidArgument, "invalid key")

	}
	publicKeyBase58 := base58.Encode(req.GetKey())
	log.Info().Str("key", publicKeyBase58).Msg("key lookup request")

	rec, err := s.dbi.GetKey(key)

	if err == db.ErrNotFound {
		return nil, status.Error(codes.NotFound, "key not found")
	}

	if err != nil {
		log.Error().Err(err).Msg("key lookup failed")
		return nil, status.Error(codes.Internal, "db error")
	}

	res := &pb.LookupKeyResponse{
		RegistrationTimestamp: rec.GetRegistrationTimestamp(),
		Disabled:              rec.GetDisabled(),
		ParentKey:             rec.GetParentKey(),
	}
	if rec.GetDisabled() {
		res.DisabledTimestamp = rec.GetDisabledTimestamp()
		res.DisabledBy = rec.GetDisabledBy()
	}
	return res, nil
}

func (s *Server) LookupName(ctx context.Context, req *pb.LookupNameRequest) (*pb.LookupNameResponse, error) {
	log.Info().Str("name", req.GetName()).Msg("name lookup request")
	if !util.ValidateName(req.GetName()) {
		log.Warn().Str("name", req.GetName()).Msg("invalid name")
		return nil, status.Error(codes.InvalidArgument, "invalid name")
	}
	key, err := s.dbi.GetName(req.GetName())
	if err == db.ErrNotFound {
		log.Debug().Str("name", req.GetName()).Msg("name not found")
		return nil, status.Error(codes.NotFound, "name was not found")
	}
	if err != nil {
		log.Error().Str("name", req.GetName()).Err(err).Msg("error getting name")
		return nil, status.Error(codes.Internal, "db error")
	}

	return &pb.LookupNameResponse{
		Key: key.SerializeCompressed(),
	}, nil
}

func (s *Server) LookupAddress(ctx context.Context, req *pb.LookupAddressRequest) (*pb.LookupAddressResponse, error) {
	log.Info().Str("name", req.GetName()).Str("protocol", req.GetProtocol().String()).Msg("address lookup request")
	address, err := s.dbi.GetAddress(req.GetName(), req.GetProtocol())
	if err == ErrNotFound {
		log.Info().Str("name", req.GetName()).Str("protocol", req.GetProtocol().String()).Msg("address not found")
		return nil, status.Error(codes.NotFound, "address not found")
	}
	if err != nil {
		log.Error().Str("name", req.GetName()).Str("protocol", req.GetProtocol().String()).Msg("error getting address")
		return nil, status.Error(codes.Internal, "db error")
	}
	return &pb.LookupAddressResponse{
		Address: address,
	}, nil
}

// loopbackLookupClient allows us to use lookup server as client.
type loopbackLookupClient struct {
	s *Server
}

func (c *loopbackLookupClient) LookupKey(ctx context.Context, in *pb.LookupKeyRequest, opts ...grpc.CallOption) (*pb.LookupKeyResponse, error) {
	return c.s.LookupKey(ctx, in)
}

func (c *loopbackLookupClient) LookupName(ctx context.Context, in *pb.LookupNameRequest, opts ...grpc.CallOption) (*pb.LookupNameResponse, error) {
	return c.s.LookupName(ctx, in)
}

func (c *loopbackLookupClient) LookupAddress(ctx context.Context, in *pb.LookupAddressRequest, opts ...grpc.CallOption) (*pb.LookupAddressResponse, error) {
	return c.s.LookupAddress(ctx, in)
}

func Int64ToBytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func verifyPow(req *pb.SignedWithPow, powStrength int) bool {
	if len(req.Pow) > 16 {
		// POW is too long.
		log.Printf("invalid POW")
		return false
	}

	if !pow.Verify(req.GetContent(), req.Pow, powStrength) {
		// POW does not check out.
		log.Printf("POW verification failed")
		return false
	}

	return true
}

func verifyPowAndSignature(req *pb.SignedWithPow, powStrength int) bool {
	if !verifyPow(req, powStrength) {
		return false
	}

	if !protoutil.VerifySignature(req.Signature, req.Key, req.Content) {
		return false
	}

	return true
}
