package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/imap"
	"github.com/regnull/ubikom/imap/db"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/pop"
	"github.com/regnull/ubikom/smtp"
	"github.com/regnull/ubikom/store"
	"github.com/regnull/ubikom/util"
)

type Args struct {
	DumpURL               string `yaml:"dump-url"`
	LookupURL             string `yaml:"lookup-url"`
	GetKeyFromUser        bool   `yaml:"get-key-from-user"`
	KeyLocation           string `yaml:"key-location"`
	PopUser               string `yaml:"pop-user"`
	PopPassword           string `yaml:"pop-password"`
	PopDomain             string `yaml:"pop-domain"`
	PopPort               int    `yaml:"pop-port"`
	ImapStorePath         string `yaml:"imap-store-path"`
	ImapDomain            string `yaml:"imap-domain"`
	ImapPort              int    `yaml:"imap-port"`
	ImapUser              string `yaml:"imap-user"`
	ImapPassword          string `yaml:"imap-password"`
	SmtpDomain            string `yaml:"smtp-domain"`
	SmtpPort              int    `yaml:"smtp-port"`
	SmtpUser              string `yaml:"smtp-user"`
	SmtpPassword          string `yaml:"smtp-password"`
	ConnectionTimeoutMsec int    `yaml:"connection-timeout-msec"`
	LogLevel              string `yaml:"log-level"`
	TLSCertFile           string `yaml:"tls-cert-file"`
	TLSKeyFile            string `yaml:"tls-key-file"`
	LocalStorePath        string `yaml:"local-store-path"`
	MaxMessageAgeHours    int    `yaml:"max-message-age-hours"`
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	configFile := util.GetConfigFromArgs(os.Args)

	configArgs := &Args{}
	err := util.FindAndParseConfigFile(configFile, &configArgs)
	if err != nil {
		log.Warn().Err(err).Msg("not using config file")
	}

	// Parse the command line arguments, use config file as defaults.
	var args Args
	var ignoreConfig string
	flag.StringVar(&ignoreConfig, "config", "", "location of the config file")
	flag.StringVar(&args.DumpURL, "dump-url", configArgs.DumpURL, "dump service URL")
	flag.StringVar(&args.LookupURL, "lookup-url", configArgs.LookupURL, "lookup service URL")
	flag.BoolVar(&args.GetKeyFromUser, "get-key-from-user", configArgs.GetKeyFromUser, "get key from POP/SMTP user")
	flag.StringVar(&args.KeyLocation, "key", configArgs.KeyLocation, "private key location")
	flag.StringVar(&args.PopUser, "pop-user", configArgs.PopUser, "name to be used by POP server")
	flag.StringVar(&args.PopPassword, "pop-password", configArgs.PopPassword, "password to be used by POP server")
	flag.StringVar(&args.PopDomain, "pop-domain", configArgs.PopDomain, "domain to be used by POP server")
	flag.IntVar(&args.PopPort, "pop-port", configArgs.PopPort, "port to be used by POP server")
	flag.StringVar(&args.ImapDomain, "imap-domain", configArgs.ImapDomain, "domain to be used by IMAP server")
	flag.IntVar(&args.ImapPort, "imap-port", configArgs.ImapPort, "port to be used by IMAP server")
	flag.StringVar(&args.ImapStorePath, "imap-store-path", configArgs.ImapStorePath, "IMAP store path")
	flag.StringVar(&args.ImapUser, "imap-user", configArgs.ImapUser, "IMAP user")
	flag.StringVar(&args.ImapPassword, "imap-password", configArgs.ImapPassword, "IMAP password")
	flag.StringVar(&args.SmtpDomain, "smtp-domain", configArgs.SmtpDomain, "domain for SMTP server")
	flag.IntVar(&args.SmtpPort, "smtp-port", configArgs.SmtpPort, "port used by SMTP server")
	flag.StringVar(&args.SmtpUser, "smtp-user", configArgs.SmtpUser, "user to be used by SMTP server")
	flag.StringVar(&args.SmtpPassword, "smtp-password", configArgs.SmtpPassword, "password to be used by SMTP server")
	flag.IntVar(&args.ConnectionTimeoutMsec, "connection-timeout-msec", configArgs.ConnectionTimeoutMsec, "connection timeout, milliseconds")
	flag.StringVar(&args.LogLevel, "log-level", configArgs.LogLevel, "log level")
	flag.StringVar(&args.TLSCertFile, "tls-cert-file", configArgs.TLSCertFile, "TLS certificate file")
	flag.StringVar(&args.TLSKeyFile, "tls-key-file", configArgs.TLSKeyFile, "TLS key file")
	flag.StringVar(&args.LocalStorePath, "local-store-path", configArgs.LocalStorePath, "path for the local messages store")
	flag.IntVar(&args.MaxMessageAgeHours, "max-message-age-hours", configArgs.MaxMessageAgeHours, "max message age, in hours")
	flag.Parse()

	err = verifyArgs(&args)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid arguments")
	}

	// Set the log level.
	logLevel, err := zerolog.ParseLevel(args.LogLevel)
	if err != nil {
		log.Fatal().Str("level", args.LogLevel).Msg("invalid log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	// Connect to the dump and lookup servers.

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Millisecond * time.Duration(args.ConnectionTimeoutMsec)),
	}

	log.Debug().Str("url", args.DumpURL).Msg("connecting to dump service")
	dumpConn, err := grpc.Dial(args.DumpURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the dump server")
	}
	defer dumpConn.Close()

	dumpClient := pb.NewDMSDumpServiceClient(dumpConn)

	log.Debug().Str("url", args.LookupURL).Msg("connecting to lookup service")
	lookupConn, err := grpc.Dial(args.LookupURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	var key *easyecc.PrivateKey
	if !args.GetKeyFromUser {
		key, err = easyecc.NewPrivateKeyFromFile(args.KeyLocation, "")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load private key")
		}
	}

	if args.TLSCertFile != "" && args.TLSKeyFile != "" {
		log.Info().Str("cert-file", args.TLSCertFile).Str("key-file", args.TLSKeyFile).Msg("using TLS")
	}

	var localStore store.Store
	if args.LocalStorePath != "" {
		localStore, err = store.NewBadger(args.LocalStorePath, time.Duration(args.MaxMessageAgeHours)*time.Hour)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create local store")
		}
	}

	popOpts := &pop.ServerOptions{
		Ctx:          context.Background(),
		Domain:       args.PopDomain,
		Port:         args.PopPort,
		User:         args.PopUser,
		Password:     args.PopPassword,
		DumpClient:   dumpClient,
		LookupClient: lookupClient,
		Key:          key,
		CertFile:     args.TLSCertFile,
		KeyFile:      args.TLSKeyFile,
		LocalStore:   localStore,
	}

	var wg sync.WaitGroup
	wg.Add(3)

	popServer := pop.NewServer(popOpts)
	go func() {
		err := popServer.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("POP server failed to initialize")
		}
		wg.Done()
	}()

	smtpOpts := &smtp.ServerOptions{
		Domain:       args.SmtpDomain,
		Port:         args.SmtpPort,
		User:         args.SmtpUser,
		Password:     args.SmtpPassword,
		LookupClient: lookupClient,
		DumpClient:   dumpClient,
		PrivateKey:   key,
		CertFile:     args.TLSCertFile,
		KeyFile:      args.TLSKeyFile,
	}
	smtpServer, err := smtp.NewServer(smtpOpts)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize SMTP server")
	}
	go func() {
		err := smtpServer.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("SMTP server failed to initialize")
		}
		wg.Done()
	}()

	imapBadger, err := db.NewBadger(args.ImapStorePath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize IMAP Badger DB")
	}
	imapOpts := &imap.ServerOptions{
		Domain:         args.ImapDomain,
		Port:           args.ImapPort,
		User:           args.ImapUser,
		Password:       args.ImapPassword,
		PrivateKey:     key,
		CertFile:       args.TLSCertFile,
		KeyFile:        args.TLSKeyFile,
		LookupClient:   lookupClient,
		DumpClient:     dumpClient,
		Badger:         imapBadger,
		PrintDebugInfo: true,
	}
	imapServer := imap.NewServer(imapOpts)
	go func() {
		err := imapServer.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("IMAP server failed to initialize")
		}
		wg.Done()
	}()

	wg.Wait()
}

func verifyArgs(args *Args) error {
	if args.ConnectionTimeoutMsec == 0 {
		args.ConnectionTimeoutMsec = 5000
	}

	if args.LogLevel == "" {
		args.LogLevel = "warn"
	}

	if args.DumpURL == "" {
		return fmt.Errorf("dump url must be specified")
	}

	if args.LookupURL == "" {
		return fmt.Errorf("lookup url must be specified")
	}

	if args.MaxMessageAgeHours == 0 {
		args.MaxMessageAgeHours = 7 * 24
	}

	if !args.GetKeyFromUser {
		if args.PopUser == "" {
			return fmt.Errorf("pop user must be specified")
		}

		if args.PopPassword == "" {
			return fmt.Errorf("pop password must be specified")
		}

		if args.SmtpUser == "" {
			return fmt.Errorf("smtp user must be specified")
		}

		if args.SmtpPassword == "" {
			return fmt.Errorf("smtp password must be specified")
		}

		// Expand home directory even if $HOME is not defined (which is the case on Windows).
		homeDir, err := os.UserHomeDir()
		if err == nil {
			args.KeyLocation = strings.Replace(args.KeyLocation, "${HOME}", homeDir, -1)
		}

		args.KeyLocation = os.ExpandEnv(args.KeyLocation)
	}

	args.LocalStorePath = os.ExpandEnv(args.LocalStorePath)
	args.ImapStorePath = os.ExpandEnv(args.ImapStorePath)

	return nil
}
