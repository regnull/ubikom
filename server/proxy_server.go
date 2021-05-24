package server

import (
	"context"
	"io/ioutil"
	"path"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type ProxyServer struct {
	pb.UnimplementedProxyServiceServer

	dataDir      string
	lookupClient pb.LookupServiceClient
}

func NewProxyServer(dataDir string, lookupClient pb.LookupServiceClient) *ProxyServer {
	return &ProxyServer{dataDir: dataDir, lookupClient: lookupClient}
}

func (s *ProxyServer) StoreEncryptedKey(ctx context.Context, req *pb.Signed) (*pb.Result, error) {
	if !protoutil.VerifySignature(req.Signature, req.Key, req.Content) {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	storeReq := &pb.StoreEncryptedKeyRequest{}
	err := proto.Unmarshal(req.GetContent(), storeReq)
	if err != nil {
		log.Warn().Err(err).Msg("error parsing store encrypted key request")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	publicKey, err := easyecc.NewPublicFromSerializedCompressed(req.GetKey())
	if err != nil {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	res, err := s.lookupClient.LookupName(ctx, &pb.LookupNameRequest{Name: storeReq.GetName()})
	if err != nil {
		log.Error().Err(err).Msg("failed to look up name")
		return &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}
	if !publicKey.EqualSerializedCompressed(res.GetKey()) {
		log.Warn().Msg("user not authorized to store key")
		return &pb.Result{Result: pb.ResultCode_RC_UNAUTHORIZED}, nil
	}

	fileName := path.Join(s.dataDir, storeReq.GetName())
	err = ioutil.WriteFile(fileName, storeReq.GetEncryptedPrivateKey(), 0600)
	if err != nil {
		log.Error().Err(err).Msg("failed to write encrypted private key")
	}
	return &pb.Result{Result: pb.ResultCode_RC_OK}, nil
}
