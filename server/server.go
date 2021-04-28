package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dgraph-io/badger/v3"

	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pow"
	"teralyt.com/ubikom/util"
)

var ErrKeyExists = fmt.Errorf("key already exists")

type Server struct {
	pb.UnimplementedIdentityServiceServer

	db *badger.DB
}

func NewServer(db *badger.DB) *Server {
	return &Server{db: db}
}

func (s *Server) RegisterKey(ctx context.Context, req *pb.KeyRegistrationRequest) (*pb.KeyRegistrationResponse, error) {
	if len(req.Key) != 33 {
		log.Printf("invalid key length")
		return &pb.KeyRegistrationResponse{
			Result: pb.ResultCode_INVALID_KEY}, nil
	}

	if len(req.Pow) > 16 {
		log.Printf("invalid POW")
		// POW is too long.
		return &pb.KeyRegistrationResponse{
			Result: pb.ResultCode_INVALID_REQUEST}, nil
	}

	if !pow.Verify(req.Key, req.Pow, 10) {
		log.Printf("POW verification failed")
		// POW does not check out.
		return &pb.KeyRegistrationResponse{
			Result: pb.ResultCode_INVALID_REQUEST}, nil
	}

	publicKeyBase58 := base58.Encode(req.Key)
	log.Printf("registering public key %s", publicKeyBase58)
	dbKey := "pkey_" + publicKeyBase58

	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbKey))
		if item != nil {
			return ErrKeyExists
		}

		ts := util.NowMs()
		err = txn.Set([]byte(dbKey), Int64ToBytes(ts))
		return err
	})
	if err != nil {
		if err == ErrKeyExists {
			log.Printf("this key is already registered")
			return &pb.KeyRegistrationResponse{
				Result: pb.ResultCode_KEY_EXISTS}, nil
		}
		log.Printf("error writing public key: %s", err)
		return &pb.KeyRegistrationResponse{
			Result: pb.ResultCode_INTERNAL_ERROR}, nil
	}
	log.Printf("key %s registered successfully", publicKeyBase58)
	return &pb.KeyRegistrationResponse{
		Result: pb.ResultCode_OK}, nil
}

func Int64ToBytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}
