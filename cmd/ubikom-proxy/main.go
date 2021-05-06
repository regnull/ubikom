package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	"github.com/regnull/ubikom/ecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/pop"
	"github.com/regnull/ubikom/smtp"
	"github.com/regnull/ubikom/util"
)

type Args struct {
	DumpURL               string
	LookupURL             string
	KeyLocation           string
	PopUser               string
	PopPassword           string
	ConnectionTimeoutMsec int
	LogLevel              string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	// Parse the config file, if it exists.
	var configFile string
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.StringVar(&configFile, "config", "", "location of the config file")
	err := flagSet.Parse(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse command line arguments")
	}

	configArgs := &Args{}
	configFile, err = util.GetConfigFileLocation(configFile)
	if err == nil {
		fmt.Printf("using config file: %s\n", configFile)
		config, err := ioutil.ReadFile(configFile)
		if err == nil {
			err := yaml.Unmarshal(config, &configArgs)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to read config file")
			}
		} else {
			log.Warn().Err(err).Msg("config file not found")
		}
	}

	// Parse the command line arguments, use config file as defaults.
	var args Args
	flag.StringVar(&args.DumpURL, "dump-url", configArgs.DumpURL, "dump service URL")
	flag.StringVar(&args.LookupURL, "lookup-url", configArgs.LookupURL, "lookup service URL")
	flag.StringVar(&args.KeyLocation, "key", configArgs.KeyLocation, "private key location")
	flag.StringVar(&args.PopUser, "user", configArgs.PopUser, "name to be used for POP & SMTP")
	flag.StringVar(&args.PopPassword, "password", configArgs.PopPassword, "password to be used by POP client")
	flag.IntVar(&args.ConnectionTimeoutMsec, "connection-timeout-msec", configArgs.ConnectionTimeoutMsec, "connection timeout, milliseconds")
	flag.StringVar(&args.LogLevel, "log-level", configArgs.LogLevel, "log level")
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

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Millisecond * time.Duration(args.ConnectionTimeoutMsec)),
	}

	dumpConn, err := grpc.Dial(args.DumpURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the dump server")
	}
	defer dumpConn.Close()

	dumpClient := pb.NewDMSDumpServiceClient(dumpConn)

	lookupConn, err := grpc.Dial(args.LookupURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	key, err := ecc.LoadPrivateKey(args.KeyLocation)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load private key")
	}

	popOpts := &pop.ServerOptions{
		Ctx:          context.Background(),
		Domain:       "localhost",
		Port:         1100,
		User:         args.PopUser,
		Password:     args.PopPassword,
		DumpClient:   dumpClient,
		LookupClient: lookupClient,
		Key:          key,
		PollInterval: time.Minute,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	popServer := pop.NewServer(popOpts)
	go func() {
		err := popServer.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("POP server failed to initialize")
		}
		wg.Done()
	}()

	smtpOpts := &smtp.ServerOptions{
		Domain: "localhost",
		Port:   1025,
		User:   args.PopUser,
		// TODO: Re-enable password, maybe.
		// Password:     "pumpkin123",
		LookupClient: lookupClient,
		DumpClient:   dumpClient,
		PrivateKey:   key,
	}
	smtpServer := smtp.NewServer(smtpOpts)
	go func() {
		err := smtpServer.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("SMTP server failed to initialize")
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

	if args.PopUser == "" {
		return fmt.Errorf("user must be specified")
	}

	if args.PopPassword == "" {
		return fmt.Errorf("password must be specified")
	}

	args.KeyLocation = os.ExpandEnv(args.KeyLocation)

	return nil
}
