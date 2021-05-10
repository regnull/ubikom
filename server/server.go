package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dgraph-io/badger/v3"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"

	"github.com/regnull/ubikom/ecc"
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

	db *badger.DB
}

func NewServer(db *badger.DB) *Server {
	return &Server{db: db}
}

func (s *Server) RegisterKey(ctx context.Context, req *pb.SignedWithPow) (*pb.Result, error) {
	if !verifyPowAndSignature(req) {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	keyRegistrationReq := &pb.KeyRegistrationRequest{}
	err := proto.Unmarshal(req.GetContent(), keyRegistrationReq)
	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal key registration request")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	if bytes.Compare(keyRegistrationReq.GetKey(), req.GetKey()) != 0 {
		log.Warn().Hex("content", req.GetContent()).Hex("key", req.GetKey()).Msg("key and content do not match (key registration request)")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	key := keyRegistrationReq.GetKey()
	publicKey, err := ecc.NewPublicFromSerializedCompressed(key)
	if err != nil {
		log.Warn().Err(err).Msg("failed to create public key from serialized compressed")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	publicKeyBase58 := base58.Encode(key)
	log.Info().Str("key", publicKeyBase58).Msg("registering public key")
	dbKey := "pkey_" + publicKeyBase58

	err = s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbKey))
		if item != nil {
			return ErrKeyExists
		}

		// Register the key.

		rec := &pb.KeyRecord{
			RegistrationTimestamp: util.NowMs()}
		recBytes, err := proto.Marshal(rec)
		if err != nil {
			return fmt.Errorf("error marshaling key record: %w", err)
		}

		err = txn.Set([]byte(dbKey), recBytes)
		if err != nil {
			return err
		}

		// Register key address as name.

		nameKey := "name_" + publicKey.Address()
		err = txn.Set([]byte(nameKey), req.GetKey())
		if err != nil {
			return err
		}

		log.Info().Str("name", publicKey.Address()).Msg("name registered")
		return nil
	})
	if err != nil {
		if err == ErrKeyExists {
			log.Warn().Str("key", publicKeyBase58).Msg("this key is already registered")
			return &pb.Result{Result: pb.ResultCode_RC_RECORD_EXISTS}, nil
		}
		log.Error().Err(err).Str("key", publicKeyBase58).Msg("error writing public key")
		return &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}
	log.Info().Str("key", publicKeyBase58).Msg("key is registered successfully")
	return &pb.Result{Result: pb.ResultCode_RC_OK}, nil
}

func (s *Server) RegisterKeyRelationship(ctx context.Context, req *pb.SignedWithPow) (*pb.Result, error) {
	if !verifyPowAndSignature(req) {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	keyRelReq := &pb.KeyRelationshipRegistrationRequest{}
	err := proto.Unmarshal(req.GetContent(), keyRelReq)
	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal key relationship registration request")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	if len(keyRelReq.GetTargetKey()) != 33 {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	if keyRelReq.GetRelationship() != pb.KeyRelationship_KR_PARENT {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	key := req.GetKey()
	publicKeyBase58 := base58.Encode(key)
	log.Info().Str("key", publicKeyBase58).Msg("registering public key")
	dbKey := "pkey_" + publicKeyBase58
	err = s.db.Update(func(txn *badger.Txn) error {

		// Retrieve the key.

		item, err := txn.Get([]byte(dbKey))
		if err != nil {
			return fmt.Errorf("error getting key record: %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		keyRec := &pb.KeyRecord{}
		err = item.Value(func(val []byte) error {
			err := proto.Unmarshal(val, keyRec)
			if err != nil {
				return fmt.Errorf("failed to unmarshal key record: %w", err)
			}
			return nil
		})

		// Add parent to the key record.

		parentKey := keyRelReq.GetTargetKey()

		for _, parent := range keyRec.GetParentKey() {
			if bytes.Compare(parentKey, parent) == 0 {
				// This key is already registered as a parent.
				return ErrKeyExists
			}
		}

		keyRec.ParentKey = append(keyRec.ParentKey, keyRelReq.GetTargetKey())
		if len(keyRec.GetParentKey()) > 16 {
			return fmt.Errorf("too many parent keys")
		}

		// Update key registration.

		recBytes, err := proto.Marshal(keyRec)
		if err != nil {
			return fmt.Errorf("error marshaling key record: %w", err)
		}

		err = txn.Set([]byte(dbKey), recBytes)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to register parent key")
		return &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}
	log.Info().Str("key", publicKeyBase58).Msg("parent key registered successfully")
	return &pb.Result{Result: pb.ResultCode_RC_OK}, nil
}

func (s *Server) DisableKey(ctx context.Context, req *pb.SignedWithPow) (*pb.Result, error) {
	if !verifyPowAndSignature(req) {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	keyDisReq := &pb.KeyDisableRequest{}
	err := proto.Unmarshal(req.GetContent(), keyDisReq)
	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal key disable request")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}
	key := keyDisReq.GetKey()
	publicKeyBase58 := base58.Encode(key)
	log.Info().Str("key", publicKeyBase58).Msg("registering public key")
	dbKey := "pkey_" + publicKeyBase58
	err = s.db.Update(func(txn *badger.Txn) error {

		// Retrieve the key.

		item, err := txn.Get([]byte(dbKey))
		if err != nil {
			return fmt.Errorf("error getting key record: %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		keyRec := &pb.KeyRecord{}
		err = item.Value(func(val []byte) error {
			err := proto.Unmarshal(val, keyRec)
			if err != nil {
				return fmt.Errorf("failed to unmarshal key record: %w", err)
			}
			return nil
		})

		// TODO: Verify that the signing key has rights to disable this key.

		keyRec.Disabled = true
		keyRec.DisabledTimestamp = util.NowMs()
		keyRec.DisabledBy = req.GetKey()

		// Update key registration.

		recBytes, err := proto.Marshal(keyRec)
		if err != nil {
			return fmt.Errorf("error marshaling key record: %w", err)
		}

		err = txn.Set([]byte(dbKey), recBytes)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to disable key")
		return &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}

	log.Info().Str("key", fmt.Sprintf("%x", key)).Msg("key disabled")

	return &pb.Result{Result: pb.ResultCode_RC_OK}, nil
}

func (s *Server) RegisterName(ctx context.Context, req *pb.SignedWithPow) (*pb.Result, error) {
	if !verifyPowAndSignature(req) {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	nameRegistrationReq := &pb.NameRegistrationRequest{}
	err := proto.Unmarshal(req.GetContent(), nameRegistrationReq)
	if err != nil {
		log.Warn().Msg("invalid name registration request")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	if !util.ValidateName(nameRegistrationReq.GetName()) {
		log.Warn().Str("name", nameRegistrationReq.GetName()).Msg("invalid name")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	log.Info().Str("name", nameRegistrationReq.GetName()).Msg("registering name")

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
			log.Warn().Str("name", nameRegistrationReq.GetName()).Msg("this name is already registered")
			return &pb.Result{
				Result: pb.ResultCode_RC_RECORD_EXISTS}, nil
		}
		log.Error().Str("name", nameRegistrationReq.GetName()).Err(err).Msg("error writing name")
		return &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}
	log.Info().Str("name", nameRegistrationReq.GetName()).Msg("name registered successfully")
	return &pb.Result{
		Result: pb.ResultCode_RC_OK}, nil
}

func (s *Server) RegisterAddress(ctx context.Context, req *pb.SignedWithPow) (*pb.Result, error) {
	if !verifyPowAndSignature(req) {
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	addressRegistrationReq := &pb.AddressRegistrationRequest{}
	err := proto.Unmarshal(req.GetContent(), addressRegistrationReq)
	if err != nil {
		log.Warn().Msg("invalid address registration request")
		return &pb.Result{Result: pb.ResultCode_RC_INVALID_REQUEST}, nil
	}

	log.Info().Str("name", addressRegistrationReq.GetName()).
		Str("address", addressRegistrationReq.GetAddress()).Msg("registering address")

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
		log.Error().Str("name", addressRegistrationReq.GetName()).
			Str("address", addressRegistrationReq.GetAddress()).Msg("error registering address")
		return &pb.Result{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}

	log.Info().Str("name", addressRegistrationReq.GetName()).
		Str("address", addressRegistrationReq.GetAddress()).Msg("address registered successfully")
	return &pb.Result{Result: pb.ResultCode_RC_OK}, nil
}

func (s *Server) LookupKey(ctx context.Context, req *pb.LookupKeyRequest) (*pb.LookupKeyResponse, error) {
	key := req.GetKey()
	publicKeyBase58 := base58.Encode(key)
	log.Info().Str("key", publicKeyBase58).Msg("key lookup request")
	dbKey := "pkey_" + publicKeyBase58
	keyRec := &pb.KeyRecord{}
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbKey))
		if err != nil {
			return fmt.Errorf("error getting key record: %w", err)
		}
		if item == nil {
			return ErrNotFound
		}

		err = item.Value(func(val []byte) error {
			err := proto.Unmarshal(val, keyRec)
			if err != nil {
				return fmt.Errorf("failed to unmarshal key record: %w", err)
			}
			return nil
		})
		return nil
	})

	if err == ErrNotFound {
		return &pb.LookupKeyResponse{Result: pb.ResultCode_RC_RECORD_NOT_FOUND}, nil
	}

	res := &pb.LookupKeyResponse{
		Result:                pb.ResultCode_RC_OK,
		RegistrationTimestamp: keyRec.GetRegistrationTimestamp(),
		Disabled:              keyRec.GetDisabled(),
		ParentKey:             keyRec.GetParentKey(),
	}
	if keyRec.GetDisabled() {
		res.DisabledTimestamp = keyRec.GetDisabledTimestamp()
		res.DisabledBy = keyRec.GetDisabledBy()
	}
	return res, nil
}

func (s *Server) LookupName(ctx context.Context, req *pb.LookupNameRequest) (*pb.LookupNameResponse, error) {
	log.Info().Str("name", req.GetName()).Msg("name lookup request")
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
		log.Info().Str("name", req.GetName()).Msg("name not found")
		return &pb.LookupNameResponse{Result: pb.ResultCode_RC_RECORD_NOT_FOUND}, nil
	}
	if err != nil {
		log.Error().Str("name", req.GetName()).Err(err).Msg("error getting name")
		return &pb.LookupNameResponse{Result: pb.ResultCode_RC_INTERNAL_ERROR}, nil
	}

	return &pb.LookupNameResponse{
		Result: pb.ResultCode_RC_OK,
		Key:    key,
	}, nil
}

func (s *Server) LookupAddress(ctx context.Context, req *pb.LookupAddressRequest) (*pb.LookupAddressResponse, error) {
	log.Info().Str("name", req.GetName()).Str("protocol", req.GetProtocol().String()).Msg("address lookup request")
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
		log.Info().Str("name", req.GetName()).Str("protocol", req.GetProtocol().String()).Msg("address not found")
		return &pb.LookupAddressResponse{Result: pb.ResultCode_RC_RECORD_NOT_FOUND}, nil
	}
	if err != nil {
		log.Error().Str("name", req.GetName()).Str("protocol", req.GetProtocol().String()).Msg("error getting address")
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

func verifyPowAndSignature(req *pb.SignedWithPow) bool {
	if !verifyPow(req) {
		return false
	}

	if !protoutil.VerifySignature(req.Signature, req.Key, req.Content) {
		return false
	}

	return true
}
