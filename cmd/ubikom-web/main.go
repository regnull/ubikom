package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/event"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var notificationMessage = `To: %s@ubikom.cc
From: Ubikom Web <%s@ubikom.cc>
Subject: New registration
Date: %s
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 7bit
Content-Language: en-US

Rejoice! For a new user just registered via the Ubikom Web!

The newly registered name is %s.

That is all. Have a nice day!
`

const (
	defaultPowStrength      = 23
	minNameLength           = 3
	minPasswordLength       = 6
	defaultRateLimitPerHour = 200
	defaultKeyOwner         = "0x24fa5B1d7FBe98A9316101E311F0c409791EaA76"
)

type CmdArgs struct {
	Port                             int
	LookupServiceURL                 string
	IdentityServiceURL               string
	ProxyManagementServiceURL        string
	Timeout                          time.Duration
	CertFile                         string
	KeyFile                          string
	UbikomKeyFile                    string
	UbikomName                       string
	NotificationName                 string
	PowStrength                      int
	RateLimitPerHour                 int
	BlockchainNodeURL                string
	UseLegacyLookupService           bool
	KeyRegistryContractAddress       string
	NameRegistryContractAddress      string
	ConnectorRegistryContractAddress string
	WelcomeMessageDir                string
}

type Server struct {
	lookupClient          pb.LookupServiceClient
	identityClient        pb.IdentityServiceClient
	proxyManagementClient pb.ProxyServiceClient
	keys                  map[string][]byte
	privateKey            *easyecc.PrivateKey
	name                  string
	notificationName      string
	powStrength           int
	rateLimiter           *rate.Limiter
	eventSender           *event.Sender
	blockchain            *bc.Blockchain
	welcomeMessageDir     string
}

func NewServer(lookupClient pb.LookupServiceClient, identityClient pb.IdentityServiceClient,
	proxyManagementClient pb.ProxyServiceClient,
	privateKey *easyecc.PrivateKey, name string, notificationName string, powStrength int,
	rateLimitPerHour int, blockhain *bc.Blockchain, welcomeMessageDir string) *Server {
	return &Server{
		lookupClient:          lookupClient,
		identityClient:        identityClient,
		proxyManagementClient: proxyManagementClient,
		keys:                  make(map[string][]byte),
		privateKey:            privateKey,
		name:                  name,
		notificationName:      notificationName,
		powStrength:           powStrength,
		rateLimiter:           rate.NewLimiter(rate.Every(time.Hour), rateLimitPerHour),
		eventSender:           event.NewSender("ubikom-event-processor", "ubikom-web", "web", privateKey, lookupClient),
		blockchain:            blockhain,
		welcomeMessageDir:     welcomeMessageDir,
	}
}

func (s *Server) HandleNameLookup(w http.ResponseWriter, r *http.Request) {
	err := s.eventSender.WebPageServed(r.Context(), "name_lookup", "", "", r.UserAgent())
	if err != nil {
		log.Error().Err(err).Msg("failed to send event")
	}
	name := r.URL.Query().Get("name")
	_, err = s.lookupClient.LookupName(r.Context(), &pb.LookupNameRequest{
		Name: name,
	})

	w.Header().Add("Access-Control-Allow-Origin", "*")

	if !util.ValidateName(name) {
		log.Warn().Str("name", name).Msg("invalid name")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err != nil && util.StatusCodeFromError(err) != codes.NotFound {
		log.Error().Err(err).Msg("name lookup request failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	found := err == nil

	w.Header().Add("Content-Type", "application/json")

	// If we got here, the name record was found.
	fmt.Fprintf(w, `{
	"name": "%s", 
	"available": %v
}`, name, !found)
}

type EasySetupRequest struct {
	Name            string `json:"name"`
	Password        string `json:"password"`
	EmailKeyOnly    bool   `json:"email_key_only"`
	CaptchaId       string `json:"captcha_id"`
	CaptchaSolution string `json:"captcha_solution"`
	Language        string `json:"language"`
}

func (s *Server) HandleEasySetup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		// This is a "pre-flight" request, see https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.rateLimiter.Allow() {
		log.Warn().Msg("rate limit triggered")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	name := ""
	password := ""
	captchaId := ""
	captchaSolution := ""
	language := ""
	useMainKey := true
	log.Debug().Str("method", r.Method).Msg("got method")
	if r.Method == "POST" {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Warn().Err(err).Msg("failed to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var req EasySetupRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse request json")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		name = req.Name
		password = req.Password
		useMainKey = !req.EmailKeyOnly
		captchaId = req.CaptchaId
		captchaSolution = req.CaptchaSolution
		if req.Language != "" {
			language = req.Language
		} else {
			language = "en"
		}
	} else {
		name = r.URL.Query().Get("name")
		password = r.URL.Query().Get("password")
		if r.URL.Query().Get("email_key_only") != "" {
			useMainKey = false
		}
	}

	captchaOk := captcha.VerifyString(captchaId, captchaSolution)
	log.Debug().Str("captcha_id", captchaId).
		Str("captcha_solution", captchaSolution).Bool("ok", captchaOk).
		Msg("captcha data")
	if !captchaOk {
		log.Warn().Msg("captcha verification failed")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if len(name) < minNameLength {
		log.Warn().Str("name", name).Msg("name is too short")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !util.ValidateName(name) {
		log.Warn().Str("name", name).Msg("invalid name")
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

	var mainKey *easyecc.PrivateKey
	if useMainKey {
		mainKey, err = easyecc.NewRandomPrivateKey()
		if err != nil {
			log.Error().Err(err).Msg("failed to generate private key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Register the main key.

		err = protoutil.RegisterKey(r.Context(), s.identityClient, mainKey, s.powStrength)
		if err != nil {
			log.Error().Err(err).Msg("failed to register the main key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Info().Msg("main key is registered")
	}

	// Create the email key.

	salt := util.Hash256([]byte(strings.ToLower(name)))
	emailKey := easyecc.NewPrivateKeyFromPassword([]byte(password), salt)

	// Register the email key.

	err = protoutil.RegisterKey(r.Context(), s.identityClient, emailKey, s.powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register the email key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("email key is registered")

	if useMainKey {
		// Register email key as a child of main key.

		err = protoutil.RegisterChildKey(r.Context(), s.identityClient, mainKey, emailKey.PublicKey(), s.powStrength)
		if err != nil {
			log.Error().Err(err).Msg("failed to register email key as a child of main key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Info().Msg("key relationship is updated")
	} else {
		mainKey = emailKey
	}

	// Register name.

	err = protoutil.RegisterName(r.Context(), s.identityClient, mainKey, emailKey.PublicKey(), name, s.powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register name")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Str("name", name).Msg("name is registered")

	// Register address.

	err = protoutil.RegisterAddress(r.Context(), s.identityClient, mainKey, emailKey.PublicKey(), name, globals.PublicDumpServiceURL, s.powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register address")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("address is registered")

	//go func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	s.blockchain.MaybeRegisterUser(ctx, name, name, password)
	//}()

	var mnemonicQuoted []string
	var keyID string
	if useMainKey {
		keyID = uuid.New().String()
		s.keys[keyID] = mainKey.Secret().Bytes()

		mnemonic, err := mainKey.Mnemonic()
		if err != nil {
			log.Error().Err(err).Msg("failed to get key mnemonic")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		mnemonicList := strings.Split(mnemonic, " ")
		mnemonicQuoted = make([]string, len(mnemonicList))
		for i := range mnemonicList {
			mnemonicQuoted[i] = "\"" + mnemonicList[i] + "\""
		}
	}

	if s.privateKey != nil && s.name != "" && s.notificationName != "" {
		body := fmt.Sprintf(notificationMessage, s.notificationName, s.name,
			time.Now().Format("02 Jan 06 15:04:05 -0700"), name)
		body, err = mail.AddReceivedHeader(body, []string{"by Ubikom client"})
		if err != nil {
			log.Error().Err(err).Msg("error adding received header")
			w.WriteHeader(http.StatusInternalServerError)
		}
		err = protoutil.SendEmail(r.Context(), s.privateKey, []byte(body), s.name, s.notificationName, s.lookupClient)
		if err != nil {
			log.Error().Err(err).Str("from", s.name).Str("to", s.notificationName).Msg("error sending notification message")
		}
	}

	if s.privateKey != nil && s.name != "" {
		welcomeMessage, err := getWelcomeMessage(s.welcomeMessageDir, language)
		if err != nil {
			welcomeMessage, err = getWelcomeMessage(s.welcomeMessageDir, "en")
			if err == nil {
				log.Info().Str("language", "en").Msg("found welcome message")
			}
		} else {
			log.Info().Str("language", language).Msg("found welcome message")
		}

		if err == nil {
			// Succeeded in getting welcome message template.
			body := fmt.Sprintf(welcomeMessage, name, s.name,
				time.Now().Format("02 Jan 06 15:04:05 -0700"))
			body, err = mail.AddReceivedHeader(body, []string{"by Ubikom client"})
			if err != nil {
				log.Error().Err(err).Msg("error adding received header")
				w.WriteHeader(http.StatusInternalServerError)
			}
			err = protoutil.SendEmail(r.Context(), s.privateKey, []byte(body), s.name, name, s.lookupClient)
			if err != nil {
				log.Error().Err(err).Str("from", s.name).Str("to", s.notificationName).Msg("error sending welcome message")
			}
		} else {
			log.Error().Err(err).Msg("failed to get welcome message")
		}
	} else {
		log.Warn().Msg("cannot send welcome message")
	}

	w.Header().Add("Content-Type", "application/json")
	if useMainKey {
		fmt.Fprintf(w, `{
		"name": "%s",
		"email": "%s@ubikom.cc",
		"user_name": "%s", 
		"server_url": "alpha.ubikom.cc",
		"incoming_server_url": "imap.ubikom.cc",
		"outgoing_server_url": "smtp.ubikom.cc",
		"incoming_server_port": "993",
		"outgoing_server_port": "495",
		"key_mnemonic": [%s],
		"key_id": "%s",
		"password": "%s"
}`, name, name, name, strings.Join(mnemonicQuoted, ", "), keyID, password)
	} else {
		fmt.Fprintf(w, `{
			"name": "%s",
			"email": "%s@ubikom.cc",
			"user_name": "%s", 
			"server_url": "alpha.ubikom.cc",
			"incoming_server_url": "imap.ubikom.cc",
			"outgoing_server_url": "smtp.ubikom.cc",
			"incoming_server_port": "993",
			"outgoing_server_port": "495",
			"password": "%s"
	}`, name, name, name, password)
	}

	err = s.eventSender.WebPageServed(r.Context(), "easy_setup", name,
		emailKey.PublicKey().Address(), r.UserAgent())
	if err != nil {
		log.Error().Err(err).Msg("failed to send event")
	}
}

type ChangePasswordRequest struct {
	Name        string `json:"name"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

func (s *Server) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		// This is a "pre-flight" request, see https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusNoContent)
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

	key, err := util.GetKeyFromNamePassword(r.Context(), req.Name, req.Password, s.lookupClient)
	if err != nil {
		log.Error().Err(err).Msg("failed to get private key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newKey := util.GenerateCanonicalKeyFromNamePassword(req.Name, req.NewPassword)

	err = protoutil.RegisterKey(r.Context(), s.identityClient, newKey, s.powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register new key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Msg("new key is registered")

	if s.proxyManagementClient != nil {
		req := &pb.CopyMailboxesRequest{
			OldKey: key.Secret().Bytes(),
			NewKey: newKey.Secret().Bytes(),
		}
		_, err := s.proxyManagementClient.CopyMailboxes(r.Context(), req)
		if err != nil {
			log.Error().Err(err).Msg("failed to copy mailboxes")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Info().Msg("mailboxes are copied")
	} else {
		log.Warn().Msg("not connected to proxy management service, will not copy mailboxes")
	}

	err = protoutil.RegisterName(r.Context(), s.identityClient, key, newKey.PublicKey(), req.Name, s.powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to register name")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info().Str("name", req.Name).Msg("name registration changed")

	err = protoutil.DisableKey(r.Context(), s.identityClient, key, s.powStrength)
	if err != nil {
		log.Error().Err(err).Msg("failed to disable the old key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.eventSender.WebPageServed(r.Context(), "change_password",
		req.Name, newKey.PublicKey().Address(), r.UserAgent())
	if err != nil {
		log.Error().Err(err).Msg("failed to send event")
	}
}

func (s *Server) HandleGetKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

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

func (s *Server) HandleNewCaptcha(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		// This is a "pre-flight" request, see https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	id := captcha.New()
	fmt.Fprintf(w, `{
		"id": "%s"
}`, id)
	log.Debug().Str("id", id).Msg("processed new captcha request")

	err := s.eventSender.WebPageServed(r.Context(), "new_captcha", "", "", r.UserAgent())
	if err != nil {
		log.Error().Err(err).Msg("failed to send event")
	}
}

func getCurrentExecDir() (dir string, err error) {
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		fmt.Printf("exec.LookPath(%s), err: %s\n", os.Args[0], err)
		return "", err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	dir = filepath.Dir(absPath)

	return dir, nil
}

func getWelcomeMessage(dir string, lang string) (string, error) {
	filePath := path.Join(dir, fmt.Sprintf("welcome_%s.txt", lang))
	if !path.IsAbs(filePath) {
		execDir, err := getCurrentExecDir()
		if err != nil {
			return "", err
		}
		filePath = path.Join(execDir, filePath)
	}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "01/02 15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", 8088, "HTTP port")
	flag.StringVar(&args.LookupServiceURL, "lookup-service-url", globals.PublicLookupServiceURL, "lookup service url")
	flag.StringVar(&args.IdentityServiceURL, "identity-service-url", globals.PublicIdentityServiceURL, "identity service url")
	flag.DurationVar(&args.Timeout, "timeout", 5*time.Second, "timeout when connecting to the lookup service")
	flag.StringVar(&args.CertFile, "cert-file", "", "certificate file")
	flag.StringVar(&args.KeyFile, "key-file", "", "key file")
	flag.StringVar(&args.UbikomKeyFile, "ubikom-key-file", "", "ubikom key file")
	flag.StringVar(&args.UbikomName, "ubikom-name", "", "ubikom name")
	flag.StringVar(&args.NotificationName, "notification-name", "", "where to send notifications")
	flag.IntVar(&args.PowStrength, "pow-strength", defaultPowStrength, "POW strength")
	flag.IntVar(&args.RateLimitPerHour, "rate-limit-per-hour", defaultRateLimitPerHour, "rate limit per hour for identity creation")
	flag.StringVar(&args.BlockchainNodeURL, "blockchain-node-url", globals.BlockchainNodeURL, "blockchain node URL")
	flag.BoolVar(&args.UseLegacyLookupService, "use-legacy-lookup-service", false, "use legacy lookup service")
	flag.StringVar(&args.KeyRegistryContractAddress, "key-registry-contract-address", globals.KeyRegistryContractAddress, "key registry contract address")
	flag.StringVar(&args.NameRegistryContractAddress, "name-registry-contract-address", globals.NameRegistryContractAddress, "name registry contract address")
	flag.StringVar(&args.ConnectorRegistryContractAddress, "connector-registry-contract-address", globals.ConnectorRegistryContractAddress, "connector registry contract address")
	flag.StringVar(&args.ProxyManagementServiceURL, "proxy-management-service-url", "", "proxy management service url")
	flag.StringVar(&args.WelcomeMessageDir, "welcome-message-dir", "welcome", "directory for welcome message")
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

	var proxyManagementClient pb.ProxyServiceClient

	if args.ProxyManagementServiceURL != "" {
		log.Info().Str("url", args.ProxyManagementServiceURL).Msg("connecting to proxy management service")
		proxyManagementConn, err := grpc.Dial(args.ProxyManagementServiceURL, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to the proxy management service")
		}

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

	var blockchain *bc.Blockchain
	// Connect to the blockchain node.
	log.Info().Str("url", args.BlockchainNodeURL).Msg("connecting to blockchain node")
	bcClient, err := ethclient.Dial(args.BlockchainNodeURL)
	if err == nil {
		log.Debug().Str("node-url", args.BlockchainNodeURL).Msg("connected to blockchain node")
		blockchain = bc.NewBlockchain(bcClient, args.KeyRegistryContractAddress,
			args.NameRegistryContractAddress, args.ConnectorRegistryContractAddress, privateKey)
	}
	if err != nil {
		log.Error().Err(err).Msg("cannot connect to blockchain")
	}

	var combinedLookupClient pb.LookupServiceClient
	if args.UseLegacyLookupService {
		log.Info().Msg("using legacy lookup service")
		combinedLookupClient = lookupClient
	} else {
		combinedLookupClient = bc.NewLookupServiceClient(blockchain, lookupClient, false)
	}

	server := NewServer(combinedLookupClient, identityClient, proxyManagementClient, privateKey, args.UbikomName,
		args.NotificationName, args.PowStrength, args.RateLimitPerHour, blockchain, args.WelcomeMessageDir)

	http.HandleFunc("/lookupName", server.HandleNameLookup)
	http.HandleFunc("/easySetup", server.HandleEasySetup)
	http.HandleFunc("/getKey", server.HandleGetKey)
	http.HandleFunc("/changePassword", server.HandleChangePassword)
	http.HandleFunc("/new_captcha", server.HandleNewCaptcha)
	http.Handle("/captcha/", captcha.Server(captcha.StdWidth, captcha.StdHeight))
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
