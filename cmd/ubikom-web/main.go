package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	defaultPowStrength      = 23
	minNameLength           = 3
	minPasswordLength       = 6
	defaultRateLimitPerHour = 200
	defaultNetwork          = "main"
)

type CmdArgs struct {
	Port                      int
	ProxyManagementServiceURL string
	Timeout                   time.Duration
	CertFile                  string
	KeyFile                   string
	UbikomKeyFile             string
	UbikomName                string
	NotificationName          string
	PowStrength               int
	RateLimitPerHour          int
	Network                   string
	InfuraProjectId           string
	ContractAddress           string
}

type Server struct {
	proxyManagementClient pb.ProxyServiceClient
	privateKey            *easyecc.PrivateKey
	name                  string
	notificationName      string
	powStrength           int
	rateLimiter           *rate.Limiter
}

func NewServer(proxyManagementClient pb.ProxyServiceClient,
	privateKey *easyecc.PrivateKey, name string, notificationName string, powStrength int,
	rateLimitPerHour int) *Server {
	return &Server{
		proxyManagementClient: proxyManagementClient,
		privateKey:            privateKey,
		name:                  name,
		notificationName:      notificationName,
		powStrength:           powStrength,
		rateLimiter:           rate.NewLimiter(rate.Every(time.Hour), rateLimitPerHour),
	}
}

type ChangePasswordRequest struct {
	Name        string `json:"name"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

func setPreflightHeaders(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.WriteHeader(http.StatusNoContent)

}

func (s *Server) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		setPreflightHeaders(w)
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")

	if r.Method != "POST" {
		log.Warn().Str("method", r.Method).Msg("invalid HTTP method")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Warn().Err(err).Msg("failed to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var req ChangePasswordRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse request json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().Str("user", req.Name).Msg("received change password request")

	key := util.GetPrivateKeyFromNameAndPassword(req.Name, req.Password)
	newKey := util.GetPrivateKeyFromNameAndPassword(req.Name, req.NewPassword)

	log.Info().Str("user", req.Name).Msg("received check mailbox key request")
	if s.proxyManagementClient == nil {
		log.Warn().Msg("not connected to proxy management service, will not copy mailboxes")
		return
	}
	req1 := &pb.CopyMailboxesRequest{
		OldKey: key.Secret().Bytes(),
		NewKey: newKey.Secret().Bytes(),
	}
	_, err = s.proxyManagementClient.CopyMailboxes(r.Context(), req1)
	if err != nil {
		code := status.Code(err)
		if code == codes.PermissionDenied {
			log.Info().Str("user", req.Name).Msg("permission denied")
			w.WriteHeader((http.StatusForbidden))
			return
		}
		if code == codes.NotFound {
			log.Info().Str("user", req.Name).Msg("not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("copy mailboxes request failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("copy mailboxes request succeeded")
}

type CheckMailboxKeyRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (s *Server) HandleCheckMailboxKey(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		setPreflightHeaders(w)
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")

	if r.Method != "POST" {
		log.Warn().Str("method", r.Method).Msg("invalid HTTP method")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Warn().Err(err).Msg("failed to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var req CheckMailboxKeyRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse request json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	privateKey := util.GetPrivateKeyFromNameAndPassword(req.Name, req.Password)

	log.Info().Str("user", req.Name).Msg("received check mailbox key request")
	if s.proxyManagementClient != nil {
		req := &pb.CheckMailboxKeyRequest{
			Name: req.Name,
			Key:  privateKey.Secret().Bytes(),
		}
		_, err := s.proxyManagementClient.CheckMailboxKey(r.Context(), req)
		if err != nil {
			code := status.Code(err)
			if code == codes.PermissionDenied {
				log.Info().Str("user", req.Name).Msg("permission denied")
				w.WriteHeader((http.StatusForbidden))
				return
			}
			if code == codes.NotFound {
				log.Info().Str("user", req.Name).Msg("not found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
			log.Error().Err(err).Msg("check mailbox key failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Info().Msg("maibox key is ok")
	} else {
		log.Warn().Msg("not connected to proxy management service, will not check key")
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "01/02 15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", 8088, "HTTP port")
	flag.DurationVar(&args.Timeout, "timeout", 5*time.Second, "timeout when connecting to the lookup service")
	flag.StringVar(&args.CertFile, "cert-file", "", "certificate file")
	flag.StringVar(&args.KeyFile, "key-file", "", "key file")
	flag.StringVar(&args.UbikomKeyFile, "ubikom-key-file", "", "ubikom key file")
	flag.StringVar(&args.UbikomName, "ubikom-name", "", "ubikom name")
	flag.StringVar(&args.NotificationName, "notification-name", "", "where to send notifications")
	flag.IntVar(&args.PowStrength, "pow-strength", defaultPowStrength, "POW strength")
	flag.IntVar(&args.RateLimitPerHour, "rate-limit-per-hour", defaultRateLimitPerHour, "rate limit per hour for identity creation")
	flag.StringVar(&args.ProxyManagementServiceURL, "proxy-management-service-url", "", "proxy management service url")
	flag.StringVar(&args.Network, "network", defaultNetwork, "ethereum network to use")
	flag.StringVar(&args.InfuraProjectId, "infura-project-id", "", "infura project id")
	flag.StringVar(&args.ContractAddress, "contract-address", "", "name registry contract address")
	flag.Parse()

	ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	var err error

	var proxyManagementClient pb.ProxyServiceClient

	if args.ProxyManagementServiceURL != "" {
		log.Info().Str("url", args.ProxyManagementServiceURL).Msg("connecting to proxy management service")
		timeoutCtx, cancel := context.WithTimeout(ctx, args.Timeout)
		proxyManagementConn, err := grpc.DialContext(timeoutCtx, args.ProxyManagementServiceURL, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the proxy management service")
		}
		defer cancel()

		proxyManagementClient = pb.NewProxyServiceClient(proxyManagementConn)
	} else {
		log.Info().Msg("running without connection to proxy management service")
	}

	var privateKey *easyecc.PrivateKey
	if args.UbikomKeyFile != "" {
		privateKey, err = easyecc.NewPrivateKeyFromFile(args.UbikomKeyFile, "")
		if err != nil {
			log.Fatal().Err(err).Str("location", args.UbikomKeyFile).Msg("cannot load private key")
		}
	}

	server := NewServer(proxyManagementClient, privateKey, args.UbikomName,
		args.NotificationName, args.PowStrength, args.RateLimitPerHour)

	http.HandleFunc("/changePassword", server.HandleChangePassword)
	http.HandleFunc("/check_mailbox_key", server.HandleCheckMailboxKey)
	log.Info().Int("port", args.Port).Msg("listening...")

	if args.CertFile != "" && args.KeyFile != "" {
		err := http.ListenAndServeTLS(fmt.Sprintf(":%d", args.Port), args.CertFile, args.KeyFile, nil)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	} else {
		err := http.ListenAndServe(fmt.Sprintf(":%d", args.Port), nil)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}
