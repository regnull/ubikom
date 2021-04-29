package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/big"

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
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial("localhost:8825", opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewIdentityServiceClient(conn)

	secret, _ := new(big.Int).SetString("c5e05b56182c7cef2ecf7edd3f27764095c524ba74db470c2b1838a1a7234bde", 16)

	//privateKey, err := ecc.NewRandomPrivateKey()
	privateKey := ecc.NewPrivateKey(secret)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Register public key.

	log.Printf("registering private key...")

	compressedKey := privateKey.PublicKey().SerializeCompressed()
	log.Printf("generating POW...")
	reqPow := pow.Compute(compressedKey, leadingZeros)
	log.Printf("POW found: %x", reqPow)

	hash := util.Hash256(compressedKey)
	sig, err := privateKey.Sign(hash)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("got hash: %x", hash)
	log.Printf("got private key: %x", privateKey.GetKey())
	log.Printf("got signature: %x, %x", sig.R, sig.S)

	req := &pb.SignedWithPow{
		Content: compressedKey,
		Pow:     reqPow,
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		Key: compressedKey,
	}

	res, err := client.RegisterKey(context.TODO(), req)
	if err != nil {
		log.Fatal(err)
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Fatalf("got response code: %d", res.Result)
	}

	log.Printf("public key registered")

	// Register name.

	log.Printf("registering name...")

	name := fmt.Sprintf("test_name_%d", util.NowMs())
	nameRegistrationReq := &pb.NameRegistrationRequest{
		Name: name}
	content, err := proto.Marshal(nameRegistrationReq)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("generating POW...")
	reqPow = pow.Compute(content, leadingZeros)
	log.Printf("POW found: %x", reqPow)

	hash = util.Hash256(content)
	sig, err = privateKey.Sign(hash)
	if err != nil {
		log.Fatal(err)
	}

	req = &pb.SignedWithPow{
		Content: content,
		Pow:     reqPow,
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		Key: compressedKey,
	}

	res, err = client.RegisterName(context.TODO(), req)
	if err != nil {
		log.Fatal(err)
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Fatalf("got response code: %d", res.Result)
	}

	log.Printf("name %s registered", name)

	// Lookup name.

	log.Printf("checking name...")

	lookupClient := pb.NewLookupServiceClient(conn)
	lookupRes, err := lookupClient.LookupName(context.TODO(), &pb.LookupNameRequest{
		Name: name})

	if err != nil {
		log.Fatal(err)
	}
	if res.Result != pb.ResultCode_RC_OK {
		log.Fatalf("got response code: %d", lookupRes.Result)
	}

	if bytes.Compare(compressedKey, lookupRes.Key) == 0 {
		log.Printf("keys match!")
	} else {
		log.Fatalf("keys do not match: expected %x, actual %x", compressedKey, lookupRes.Key)
	}
}
