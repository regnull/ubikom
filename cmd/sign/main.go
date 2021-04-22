package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec"
)

func main() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	msg := "hello, world"
	hash := sha256.Sum256([]byte(msg))

	sig, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("signature: %x\n", sig)

	valid := ecdsa.VerifyASN1(&privateKey.PublicKey, hash[:], sig)
	fmt.Println("signature verified:", valid)
}

func foo() {
	curve := elliptic.P256()
	pk, err := btcec.NewPrivateKey(curve)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("private key %s\n", pk.D.Text(16))

	plainText := "Hello, world!"
	hash := sha256.Sum256([]byte(plainText))
	sig, err := pk.Sign(hash[:])
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("signature: %x\n", sig.Serialize())
	if sig.Verify(hash[:], pk.PubKey()) {
		fmt.Printf("verified!\n")
	}
}
