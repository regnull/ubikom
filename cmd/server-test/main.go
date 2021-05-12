package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	leadingZeros = 23
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial("localhost:8825", opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the server")
	}
	defer conn.Close()
	client := pb.NewIdentityServiceClient(conn)

	privateKey, err := ecc.NewRandomPrivateKey()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate private key")
	}

	// Register public key.

	log.Info().Msg("registering private key...")

	compressedKey := privateKey.PublicKey().SerializeCompressed()

	keyRegistrationReq := &pb.KeyRegistrationRequest{
		Key: compressedKey}

	content, err := proto.Marshal(keyRegistrationReq)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to serialize key registration request")
	}

	req, err := protoutil.CreateSignedWithPOW(privateKey, content, leadingZeros)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create request")
	}

	res, err := client.RegisterKey(ctx, req)
	if err != nil {
		log.Fatal().Err(err).Msg("key registration call failed")
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Fatal().Str("result", res.GetResult().String()).Msg("key registraion call failed")
	}

	log.Info().Msg("public key registered")

	// Register name.

	log.Info().Msg("registering name...")

	name := fmt.Sprintf("test_name_%d", util.NowMs())
	nameRegistrationReq := &pb.NameRegistrationRequest{
		Name: name}
	content, err = proto.Marshal(nameRegistrationReq)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal proto")
	}

	req, err = protoutil.CreateSignedWithPOW(privateKey, content, leadingZeros)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create request")
	}

	res, err = client.RegisterName(ctx, req)
	if err != nil {
		log.Fatal().Err(err).Msg("name registration call failed")
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Fatal().Str("result", res.GetResult().String()).Msg("name registraion call failed")
	}

	log.Info().Str("name", name).Msg("name is registered successfully")

	// Lookup name.

	log.Info().Msg("checking name...")

	lookupClient := pb.NewLookupServiceClient(conn)
	lookupRes, err := lookupClient.LookupName(ctx, &pb.LookupNameRequest{
		Name: name})

	if err != nil {
		log.Fatal().Err(err).Msg("name lookup call failed")
	}
	if lookupRes.GetResult().GetResult() != pb.ResultCode_RC_OK {
		log.Fatal().Str("result", res.GetResult().String()).Msg("name lookup call failed")
	}

	if bytes.Compare(compressedKey, lookupRes.Key) == 0 {
		log.Info().Msg("keys match!")
	} else {
		log.Fatal().Msg("keys do not match")
	}

	// Register address.

	log.Info().Msg("registering address...")

	address := "111.222.333.444"
	addressRegistrationReq := &pb.AddressRegistrationRequest{
		Name:     name,
		Protocol: pb.Protocol_PL_DMS,
		Address:  address}

	content, err = proto.Marshal(addressRegistrationReq)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal proto")
	}

	req, err = protoutil.CreateSignedWithPOW(privateKey, content, leadingZeros)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create request")
	}

	res, err = client.RegisterAddress(ctx, req)
	if err != nil {
		log.Fatal().Err(err).Msg("address registration call failed")
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Fatal().Str("result", res.GetResult().String()).Msg("address registraion call failed")
	}

	log.Info().Str("name", name).Str("address", address).Msg("address is registered successfully")

	// Lookup address.

	log.Info().Msg("looking up address...")
	addressLookupRes, err := lookupClient.LookupAddress(ctx, &pb.LookupAddressRequest{
		Name:     name,
		Protocol: pb.Protocol_PL_DMS})
	if err != nil {
		log.Fatal().Err(err).Msg("address lookup call failed")
	}
	if addressLookupRes.Result != pb.ResultCode_RC_OK {
		log.Fatal().Str("result", res.GetResult().String()).Msg("name lookup call failed")
	}

	if addressLookupRes.Address != address {
		log.Fatal().Str("expected", address).Str("actual", addressLookupRes.Address).Msg("addresses do not match")
	}

	log.Info().Msg("addresses match!")
}
