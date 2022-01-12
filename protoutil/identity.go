package protoutil

import (
	"context"

	"github.com/regnull/easyecc"

	"github.com/regnull/ubikom/pb"
	"google.golang.org/protobuf/proto"
)

func RegisterKey(ctx context.Context, client pb.IdentityServiceClient, key *easyecc.PrivateKey, powStrength int) error {
	registerKeyReq := &pb.KeyRegistrationRequest{
		Key: key.PublicKey().SerializeCompressed()}
	reqBytes, err := proto.Marshal(registerKeyReq)
	if err != nil {
		return err
	}

	req, err := CreateSignedWithPOW(key, reqBytes, powStrength)
	if err != nil {
		return err
	}

	_, err = client.RegisterKey(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func DisableKey(ctx context.Context, client pb.IdentityServiceClient, key *easyecc.PrivateKey, powStrength int) error {
	registerKeyReq := &pb.KeyDisableRequest{
		Key: key.PublicKey().SerializeCompressed()}
	reqBytes, err := proto.Marshal(registerKeyReq)
	if err != nil {
		return err
	}

	req, err := CreateSignedWithPOW(key, reqBytes, powStrength)
	if err != nil {
		return err
	}

	_, err = client.DisableKey(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func RegisterChildKey(ctx context.Context, client pb.IdentityServiceClient, key *easyecc.PrivateKey,
	childKey *easyecc.PublicKey, powStrength int) error {
	registerKeyReq := &pb.KeyRelationshipRegistrationRequest{
		TargetKey:    childKey.SerializeCompressed(),
		Relationship: pb.KeyRelationship_KR_PARENT}
	reqBytes, err := proto.Marshal(registerKeyReq)
	if err != nil {
		return err
	}

	req, err := CreateSignedWithPOW(key, reqBytes, powStrength)
	if err != nil {
		return err
	}

	_, err = client.RegisterKeyRelationship(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func RegisterName(ctx context.Context, client pb.IdentityServiceClient, key *easyecc.PrivateKey,
	targetKey *easyecc.PublicKey, name string, powStrength int) error {
	registerKeyReq := &pb.NameRegistrationRequest{
		Key:  targetKey.SerializeCompressed(),
		Name: name}
	reqBytes, err := proto.Marshal(registerKeyReq)
	if err != nil {
		return err
	}

	req, err := CreateSignedWithPOW(key, reqBytes, powStrength)
	if err != nil {
		return err
	}

	_, err = client.RegisterName(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func RegisterAddress(ctx context.Context, client pb.IdentityServiceClient, key *easyecc.PrivateKey,
	targetKey *easyecc.PublicKey, name string, address string, powStrength int) error {
	registerAddressReq := &pb.AddressRegistrationRequest{
		Key:      targetKey.SerializeCompressed(),
		Name:     name,
		Protocol: pb.Protocol(pb.Protocol_PL_DMS),
		Address:  address}

	reqBytes, err := proto.Marshal(registerAddressReq)
	if err != nil {
		return err
	}

	req, err := CreateSignedWithPOW(key, reqBytes, powStrength)
	if err != nil {
		return err
	}

	_, err = client.RegisterAddress(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
