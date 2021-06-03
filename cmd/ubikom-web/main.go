package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	powStrength       = 23
	dumpAddress       = "alpha.ubikom.cc:8826"
	minNameLength     = 3
	minPasswordLength = 6
)

type CmdArgs struct {
	Port               int
	LookupServiceURL   string
	IdentityServiceURL string
	Timeout            time.Duration
	CertFile           string
	KeyFile            string
}

type Server struct {
	lookupClient   pb.LookupServiceClient
	identityClient pb.IdentityServiceClient
	keys           map[string][]byte
}

func NewServer(lookupClient pb.LookupServiceClient, identityClient pb.IdentityServiceClient) *Server {
	return &Server{
		lookupClient:   lookupClient,
		identityClient: identityClient,
		keys:           make(map[string][]byte),
	}
}

func (s *Server) HandleNameLookup(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	_, err := s.lookupClient.LookupName(r.Context(), &pb.LookupNameRequest{
		Name: name,
	})
	w.Header().Add("Content-Type", "application/json")
	if err != nil && util.StatusCodeFromError(err) != codes.NotFound {
		log.Error().Err(err).Msg("name lookup request failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	found := err == nil

	// If we got here, the name record was found.
	fmt.Fprintf(w, `{
	"name": "%s", 
	"available": %v
}`, name, !found)
}

func (s *Server) HandleEasySetup(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	password := r.URL.Query().Get("password")

	if len(name) < minNameLength {
		log.Warn().Str("name", name).Msg("name is too short")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(password) < minPasswordLength {
		log.Warn().Str("name", name).Msg("password is too short")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := s.lookupClient.LookupName(r.Context(), &pb.LookupNameRequest{Name: name})
	if err == nil {
		// This name is taken.
		log.Warn().Str("name", name).Msg("name is not available")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err != nil && util.StatusCodeFromError(err) != codes.NotFound {
		log.Error().Err(err).Msg("failed to check name availability")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mainKey, err := easyecc.NewRandomPrivateKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate private key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Register the main key.

	err = protoutil.RegisterKey(r.Context(), s.identityClient, mainKey, powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register the main key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("main key is registered")

	// Create the email key.

	var saltArr [8]byte
	_, err = rand.Read(saltArr[:])
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate salt")
	}
	salt := saltArr[:]
	userName := base58.Encode(salt[:])
	emailKey := easyecc.NewPrivateKeyFromPassword([]byte(password), salt)

	// Register the email key.

	err = protoutil.RegisterKey(r.Context(), s.identityClient, emailKey, powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register the email key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("email key is registered")

	// Register email key as a child of main key.

	err = protoutil.RegisterChildKey(r.Context(), s.identityClient, mainKey, emailKey.PublicKey(), powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register email key as a child of main key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("key relationship is updated")

	// Register name.

	err = protoutil.RegisterName(r.Context(), s.identityClient, mainKey, emailKey.PublicKey(), name, powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register name")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("name is registered")

	// Register address.

	err = protoutil.RegisterAddress(r.Context(), s.identityClient, mainKey, emailKey.PublicKey(), name, dumpAddress, powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register address")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("address is registered")

	keyID := uuid.New().String()
	s.keys[keyID] = mainKey.Secret().Bytes()

	mnemonic, err := mainKey.Mnemonic()
	if err != nil {
		log.Error().Err(err).Msg("failed to get key mnemonic")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	mnemonicList := strings.Split(mnemonic, " ")
	mnemonicQuoted := make([]string, len(mnemonicList))
	for i := range mnemonicList {
		mnemonicQuoted[i] = "\"" + mnemonicList[i] + "\""
	}

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"name": "%s",
		"user_name": "%s", 
		"server_url": "alpha.ubikom.cc",
		"key_mnemonic": [%s],
		"key_id": "%s",
		"password": "%s"
}`, name, userName, strings.Join(mnemonicQuoted, ", "), keyID, password)
}

func (s *Server) HandleGetKey(w http.ResponseWriter, r *http.Request) {
	keyID := r.URL.Query().Get("key_id")
	key, ok := s.keys[keyID]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", "attachment; filename=\"ubikom.private_key\"")
	w.Write(key)
	// TODO: Remove this after the testing is done.
	// delete(s.keys, keyID)
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", 8088, "HTTP port")
	flag.StringVar(&args.LookupServiceURL, "lookup-service-url", globals.PublicLookupServiceURL, "lookup service url")
	flag.StringVar(&args.IdentityServiceURL, "identity-service-url", globals.PublicIdentityServiceURL, "identity service url")
	flag.DurationVar(&args.Timeout, "timeout", 5*time.Second, "timeout when connecting to the lookup service")
	flag.StringVar(&args.CertFile, "cert-file", "", "certificate file")
	flag.StringVar(&args.KeyFile, "key-file", "", "key file")
	flag.Parse()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(args.Timeout),
	}

	log.Info().Str("url", args.LookupServiceURL).Msg("connecting to lookup service")
	lookupConn, err := grpc.Dial(args.LookupServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup service")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	log.Info().Str("url", args.IdentityServiceURL).Msg("connecting to identity service")
	identityConn, err := grpc.Dial(args.IdentityServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the identity service")
	}
	defer lookupConn.Close()

	identityClient := pb.NewIdentityServiceClient(identityConn)

	server := NewServer(lookupClient, identityClient)

	http.HandleFunc("/lookupName", server.HandleNameLookup)
	http.HandleFunc("/easySetup", server.HandleEasySetup)
	http.HandleFunc("/getKey", server.HandleGetKey)
	log.Info().Int("port", args.Port).Msg("listening...")

	if args.CertFile != "" && args.KeyFile != "" {
		log.Fatal().Err(http.ListenAndServeTLS(fmt.Sprintf(":%d", args.Port), args.CertFile, args.KeyFile, nil))
	} else {
		log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", args.Port), nil))
	}
}
