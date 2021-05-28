package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/regnull/easyecc"

	"github.com/btcsuite/btcutil/base58"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	powStrength = 23
	dumpAddress = "alpha.ubikom.cc:8826"
)

type CmdArgs struct {
	IdentityServiceURL string
	LookupServiceURL   string
	MainKeyLoc         string
	EmailKeyLoc        string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.IdentityServiceURL, "identity-url", globals.PublicIdentityServiceURL, "identity service url")
	flag.StringVar(&args.LookupServiceURL, "lookup-url", globals.PublicLookupServiceURL, "lookup service url")
	flag.StringVar(&args.MainKeyLoc, "main-key-location", "", "main key location")
	flag.StringVar(&args.EmailKeyLoc, "email-key-location", "", "email key location")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter the name you would like to use: ")
	name, err := reader.ReadString('\n')
	name = strings.TrimSuffix(name, "\n")
	name = strings.TrimSuffix(name, "\r")

	if err != nil {
		log.Fatal().Err(err).Msg("error reading name")
	}
	if len(name) < 3 {
		log.Fatal().Msg("the name must be at least 3 characters long")
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}

	// Connect to lookup service.

	lookupConn, err := grpc.Dial(args.LookupServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the server")
	}
	defer lookupConn.Close()
	lookupClient := pb.NewLookupServiceClient(lookupConn)

	// Connect to identity service.

	identityConn, err := grpc.Dial(args.IdentityServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the server")
	}
	defer identityConn.Close()
	identityClient := pb.NewIdentityServiceClient(identityConn)

	ctx := context.Background()

	available, err := protoutil.CheckNameAvailability(ctx, lookupClient, name)
	if err != nil {
		log.Fatal().Err(err).Msg("name lookup failed")
	}
	if !available {
		log.Fatal().Str("name", name).Msg("this name is taken")
	}

	fmt.Print("Enter new password: ")
	password, err := reader.ReadString('\n')
	password = strings.TrimSuffix(password, "\n")
	password = strings.TrimSuffix(password, "\r")
	if len(password) < 8 {
		log.Fatal().Msg("password must be at least 8 characters long")
	}

	// Create the main key.

	keyLoc := args.MainKeyLoc

	if keyLoc == "" {
		keyLoc, err = util.GetDefaultKeyLocation()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key location")
		}
		keyDir := path.Dir(keyLoc)
		err = os.MkdirAll(keyDir, 0700)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create key directory")
		}
	}

	mainKey, err := easyecc.NewRandomPrivateKey()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate private key")
	}
	err = mainKey.Save(keyLoc, "")
	if err != nil {
		log.Fatal().Err(err).Str("location", keyLoc).Msg("failed to save private key")
	}

	// Register the main key.

	err = protoutil.RegisterKey(ctx, identityClient, mainKey, powStrength)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register the main key")
	}
	log.Info().Msg("main key is registered")

	// Create the email key.

	emailKeyLoc := args.EmailKeyLoc

	if emailKeyLoc == "" {
		keyLoc, err = util.GetDefaultKeyLocation()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get key location")
		}
		keyDir := path.Dir(keyLoc)
		emailKeyLoc = path.Join(keyDir, "email.key")
	}
	var saltArr [8]byte
	_, err = rand.Read(saltArr[:])
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate salt")
	}
	salt := saltArr[:]
	userName := base58.Encode(salt[:])
	emailKey := easyecc.NewPrivateKeyFromPassword([]byte(password), salt)
	err = emailKey.Save(emailKeyLoc, "")
	if err != nil {
		log.Fatal().Err(err).Str("location", emailKeyLoc).Msg("failed to save email key")
	}

	// Register the email key.

	err = protoutil.RegisterKey(ctx, identityClient, emailKey, powStrength)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register the email key")
	}
	log.Info().Msg("email key is registered")

	// Register email key as a child of main key.

	err = protoutil.RegisterChildKey(ctx, identityClient, mainKey, emailKey.PublicKey(), powStrength)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register email key as a child of main key")
	}
	log.Info().Msg("key relationship is updated")

	// Register name.

	err = protoutil.RegisterName(ctx, identityClient, mainKey, emailKey.PublicKey(), name, powStrength)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register name")
	}
	log.Info().Msg("name is registered")

	// Register address.

	err = protoutil.RegisterAddress(ctx, identityClient, mainKey, emailKey.PublicKey(), name, dumpAddress, powStrength)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register address")
	}
	log.Info().Msg("address is registered")

	fmt.Printf("\nUse the following information in your email client:\n")
	fmt.Printf("User name: %s\n", userName)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("POP and SMTP server address: %s\n", "alpha.ubikom.cc")
}
