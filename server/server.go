package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dgraph-io/badger/v3"
	"github.com/golang/protobuf/proto"

	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pow"
	"teralyt.com/ubikom/util"
)

var ErrKeyExists = fmt.Errorf("key already exists")
var ErrNotAuthorized = fmt.Errorf("not authorized")
var ErrNotFound = fmt.Errorf("not found")

type Server struct {
	pb.UnimplementedIdentityServiceServer
	pb.UnimplementedLookupServiceServer

	db *badger.DB
}

func NewServer(db *badger.DB) *Server {
	return &Server{db: db}
}

func (s *Server) RegisterKey(ctx context.Context, req *pb.SignedWithPow) (*pb.Result, error) {
	if bytes.Compare(req.GetContent(), req.GetKey()) != 0 {
		return &pb.Result{
			Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	if !verifyPowAndSignature(req) {
		return &pb.Result{
			Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	publicKeyBase58 := base58.Encode(req.GetContent())
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
			return &pb.Result{
				Result: pb.ResultCode_RC_RECORD_EXISTS}, nil
		}
		log.Printf("error writing public key: %s", err)
		return &pb.Result{
			Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}
	log.Printf("key %s registered successfully", publicKeyBase58)
	return &pb.Result{
		Result: pb.ResultCode_RC_OK}, nil
}

func (s *Server) RegisterName(ctx context.Context, req *pb.SignedWithPow) (*pb.Result, error) {
	if !verifyPowAndSignature(req) {
		return &pb.Result{
			Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	nameRegistrationReq := &pb.NameRegistrationRequest{}
	err := proto.Unmarshal(req.GetContent(), nameRegistrationReq)
	if err != nil {
		return &pb.Result{
			Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	if !validateName(nameRegistrationReq.GetName()) {
		return &pb.Result{
			Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	dbKey := "name_" + nameRegistrationReq.GetName()
	err = s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbKey))
		if item != nil {
			return ErrKeyExists
		}

		err = txn.Set([]byte(dbKey), req.GetKey())
		return err
	})
	if err != nil {
		if err == ErrKeyExists {
			log.Printf("this key is already registered")
			return &pb.Result{
				Result: pb.ResultCode_RC_RECORD_EXISTS}, nil
		}
		log.Printf("error writing public key: %s", err)
		return &pb.Result{
			Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}
	log.Printf("name %s registered successfully", nameRegistrationReq.GetName())
	return &pb.Result{
		Result: pb.ResultCode_RC_OK}, nil
}

func (s *Server) RegisterAddress(ctx context.Context, req *pb.SignedWithPow) (*pb.Result, error) {
	if !verifyPowAndSignature(req) {
		return &pb.Result{
			Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	addressRegistrationReq := &pb.AddressRegistrationRequest{}
	err := proto.Unmarshal(req.GetContent(), addressRegistrationReq)
	if err != nil {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}
	err = s.db.Update(func(txn *badger.Txn) error {
		// Make sure the requested name is registered to this key.
		nameKey := "name_" + addressRegistrationReq.GetName()
		item, err := txn.Get([]byte(nameKey))
		if err != nil {
			return fmt.Errorf("error getting name, %w", err)
		}
		if item == nil {
			return fmt.Errorf("name %s is not registered", addressRegistrationReq.GetName())
		}

		// This is the key associated with the name. It must match the key that signed
		// this request.
		var registeredKey []byte
		err = item.Value(func(val []byte) error {
			registeredKey = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return fmt.Errorf("error getting name, %w", err)
		}

		if bytes.Compare(registeredKey, req.GetKey()) != 0 {
			return ErrNotAuthorized
		}

		addressKey := "address_" + addressRegistrationReq.GetName() + "_" + addressRegistrationReq.GetProtocol().String()
		err = txn.Set([]byte(addressKey), []byte(addressRegistrationReq.GetAddress()))
		return err
	})
	if err != nil {
		return &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}

	return &pb.Result{Result: pb.ResultCode_RC_OK}, nil
}

func (s *Server) LookupName(ctx context.Context, req *pb.LookupNameRequest) (*pb.LookupNameResponse, error) {
	var key []byte
	err := s.db.View(func(txn *badger.Txn) error {
		nameKey := "name_" + req.GetName()
		item, err := txn.Get([]byte(nameKey))
		if err != nil {
			return fmt.Errorf("error getting name, %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		err = item.Value(func(val []byte) error {
			key = append([]byte{}, val...)
			return nil
		})
		return err
	})
	if err == ErrNotFound {
		return &pb.LookupNameResponse{Result: pb.ResultCode_RC_RECORD_NOT_FOUND}, nil
	}
	if err != nil {
		return &pb.LookupNameResponse{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}

	return &pb.LookupNameResponse{
		Result: pb.ResultCode_RC_OK,
		Key:    key,
	}, nil
}

func (s *Server) LookupAddress(ctx context.Context, req *pb.LookupAddressRequest) (*pb.LookupAddressResponse, error) {
	var addressBytes []byte
	err := s.db.View(func(txn *badger.Txn) error {
		addressKey := "address_" + req.GetName() + "_" + req.GetProtocol().String()
		item, err := txn.Get([]byte(addressKey))
		if err != nil {
			return fmt.Errorf("error getting name, %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		err = item.Value(func(val []byte) error {
			addressBytes = append([]byte{}, val...)
			return nil
		})
		return err
	})
	if err == ErrNotFound {
		return &pb.LookupAddressResponse{Result: pb.ResultCode_RC_RECORD_NOT_FOUND}, nil
	}
	if err != nil {
		return &pb.LookupAddressResponse{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}
	return &pb.LookupAddressResponse{
		Result:  pb.ResultCode_RC_OK,
		Address: string(addressBytes),
	}, nil
}

func Int64ToBytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func verifyPow(req *pb.SignedWithPow) bool {
	if len(req.Pow) > 16 {
		// POW is too long.
		log.Printf("invalid POW")
		return false
	}

	if !pow.Verify(req.GetContent(), req.Pow, 10) {
		// POW does not check out.
		log.Printf("POW verification failed")
		return false
	}

	return true
}

func verifySignature(req *pb.SignedWithPow) bool {
	key, err := ecc.NewPublicFromSerializedCompressed(req.Key)
	if err != nil {
		log.Printf("invalid serialized compressed key")
		return false
	}

	sig := &ecc.Signature{
		R: new(big.Int).SetBytes(req.Signature.R),
		S: new(big.Int).SetBytes(req.Signature.S)}

	if !sig.Verify(key, util.Hash256(req.Content)) {
		log.Printf("signature verification failed")
		return false
	}
	return true
}

func verifyPowAndSignature(req *pb.SignedWithPow) bool {
	if !verifyPow(req) {
		return false
	}

	if !verifySignature(req) {
		return false
	}

	return true
}

func validateName(name string) bool {
	if len(name) < 5 || len(name) > 64 {
		return false
	}

	// TODO: More checks here.

	return true
}
