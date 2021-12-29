package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/ubchain/keyregistry"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"google.golang.org/grpc"
)

func main() {
	var (
		nodeURL         string
		contractAddress string
		userName        string
		password        string
		LookupURL       string
	)
	flag.StringVar(&nodeURL, "node-url", "http://127.0.0.1:7545", "URL of the node to connect to")
	flag.StringVar(&contractAddress, "contract-address", "", "contract address")
	flag.StringVar(&userName, "user-name", "", "user name")
	flag.StringVar(&password, "password", "", "account password")
	flag.StringVar(&LookupURL, "lookup-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.Parse()

	if userName == "" {
		log.Fatal("--user-namemust be specified")
	}

	if contractAddress == "" {
		log.Fatal("--contact-address must be specified")
	}

	if password == "" {
		log.Fatal("--password must be specified")
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Millisecond * time.Duration(5*time.Second)),
	}
	lookupConn, err := grpc.Dial(LookupURL, opts...)
	if err != nil {
		log.Fatal("failed to connect to the lookup server")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	ctx := context.Background()
	privateKey, err := util.GetKeyFromNamePassword(ctx, userName, password, lookupClient)

	// Connect to the node.
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		log.Fatal(err)
	}

	address := crypto.PubkeyToAddress(*privateKey.PublicKey().ToECDSA())
	fmt.Printf("address: %s\n", address.Hex())

	// Get nonce.
	nonce, err := client.PendingNonceAt(ctx, address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("got nonce: %d\n", nonce)

	// Recommended gas limit.
	gasLimit := uint64(300000)

	// Get gas price.
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("gas price: %d\n", gasPrice)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("chain ID: %d\n", chainID)

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey.ToECDSA(), chainID)
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	instance, err := keyregistry.NewKeyregistry(address, client)
	if err != nil {
		log.Fatal(err)
	}

	key := [32]byte{}
	value := [32]byte{}
	copy(key[:], []byte("hello"))
	copy(value[:], []byte("world"))

	tx, err := instance.Register(auth, privateKey.PublicKey().X().Bytes())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s\n", tx.Hash().Hex())
}
