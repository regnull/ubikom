package main

import (
	"flag"
	"log"
	"math/big"
	"os"

	"github.com/regnull/easyecc/v2"
	"github.com/regnull/ubikom/util"
)

func main() {
	var input string
	var output string
	flag.StringVar(&input, "in", "", "input")
	flag.StringVar(&output, "out", "", "output")
	flag.Parse()

	if input == "" {
		log.Fatal("--in is required")
	}
	if output == "" {
		log.Fatal("--out is required")
	}

	b, err := os.ReadFile(input)
	if err != nil {
		log.Fatal("failed to load private key: %v", err)
	}
	secret := new(big.Int)
	secret.SetBytes(b)
	privateKey := easyecc.NewPrivateKeyFromSecret(easyecc.SECP256K1, secret)
	passphrase, err := util.EnterPassphrase()
	if err != nil {
		log.Fatal("failed to get passphase")
	}
	err = privateKey.Save(output, passphrase)
	if err != nil {
		log.Fatal(err)
	}
}
