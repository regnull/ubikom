package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pow"
	"teralyt.com/ubikom/util"
)

const (
	leadingZeros = 10
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

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
	req, err := CreateSignedWithPOW(privateKey, compressedKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create request")
	}

	res, err := client.RegisterKey(context.TODO(), req)
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
	content, err := proto.Marshal(nameRegistrationReq)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal proto")
	}

	req, err = CreateSignedWithPOW(privateKey, content)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create request")
	}

	res, err = client.RegisterName(context.TODO(), req)
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
	lookupRes, err := lookupClient.LookupName(context.TODO(), &pb.LookupNameRequest{
		Name: name})

	if err != nil {
		log.Fatal().Err(err).Msg("name lookup call failed")
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Fatal().Str("result", res.GetResult().String()).Msg("name lookup call failed")
	}

	if bytes.Compare(compressedKey, lookupRes.Key) == 0 {
		log.Info().Msg("keys match!")
	} else {
		log.Fatal().Msg("keys do not match")
	}
}

func CreateSignedWithPOW(privateKey *ecc.PrivateKey, content []byte) (*pb.SignedWithPow, error) {
	compressedKey := privateKey.PublicKey().SerializeCompressed()

	log.Info().Msg("generating POW...")
	reqPow := pow.Compute(content, leadingZeros)
	log.Info().Hex("pow", reqPow).Msg("POW found")

	hash := util.Hash256(content)
	sig, err := privateKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request, %w", err)
	}

	req := &pb.SignedWithPow{
		Content: content,
		Pow:     reqPow,
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		Key: compressedKey,
	}
	return req, nil
}
