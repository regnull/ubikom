package main

import (
	"fmt"
	"net"
	"os"

	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	defaultPort               = 8826
	defaultDataDir            = "$HOME/.ubikom/dump"
	defaultMaxMessageAgeHours = 14 * 24
	defaultNetwork            = "main"
	defaultLogLevel           = "info"
	defaultLogNoColor         = false

	// Configuration options names.
	configPort               = "port"
	configDataDir            = "data-dir"
	configMaxMessageAgeHours = "max-message-age-hours"
	configNetwork            = "network"
	configLogLevel           = "log-level"
	configLogNoColor         = "log-no-color"
	configInfuraProjectId    = "infura-project-id"
	configContractAddress    = "contract-address"
)

type CmdArgs struct {
	DataDir            string
	Port               int
	MaxMessageAgeHours int
	Network            string
	InfuraProjectId    string
	ContractAddress    string
	LogLevel           string
	LogNoColor         bool
	ConfigFile         string
}

func SetIfNotDefault(name string, v any, def any) {
	if v != def {
		viper.Set(name, v)
	}
}

func InitConfigOrDie() {
	// Init log with temporary values.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Set defaults.
	viper.SetDefault(configPort, defaultPort)
	viper.SetDefault(configDataDir, defaultDataDir)
	viper.SetDefault(configMaxMessageAgeHours, defaultMaxMessageAgeHours)
	viper.SetDefault(configNetwork, defaultNetwork)
	viper.SetDefault(configLogLevel, defaultLogLevel)
	viper.SetDefault(configLogNoColor, defaultLogNoColor)
	viper.SetDefault(configContractAddress, globals.MainnetNameRegistryAddress)

	// Command line flags.
	var args CmdArgs
	flag.IntVar(&args.Port, configPort, 0, "port to listen to")
	flag.StringVar(&args.DataDir, configDataDir, "", "base directory")
	flag.IntVar(&args.MaxMessageAgeHours, configMaxMessageAgeHours, 0, "max message age, in hours")
	flag.StringVar(&args.Network, configNetwork, "", "ethereum network to use")
	flag.StringVar(&args.InfuraProjectId, configInfuraProjectId, "", "infura project id")
	flag.StringVar(&args.ContractAddress, configContractAddress, "", "name registry contract address")
	flag.StringVar(&args.LogLevel, configLogLevel, "", "log level")
	flag.BoolVar(&args.LogNoColor, configLogNoColor, false, "disable colors for logging")
	flag.StringVar(&args.ConfigFile, "config", "", "config file location")
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	// Environment variables overrides.
	viper.BindEnv(configNetwork, "UBK_NETWORK")
	viper.BindEnv(configInfuraProjectId, "UBK_INFURA_PROJECT_ID")
	viper.BindEnv(configContractAddress, "UBK_CONTRACT_ADDRESS")
	viper.BindEnv(configLogLevel, "UBK_LOG_LEVEL")
	viper.BindEnv(configLogNoColor, "UBK_LOG_NO_COLOR")

	// Config file overrides.
	if args.ConfigFile != "" {
		viper.SetConfigFile(args.ConfigFile)
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Str("path", args.ConfigFile).Msg("failed to read config file")
		}
	}

	// Validate config.
	if viper.GetString(configInfuraProjectId) == "" {
		log.Fatal().Msg("infura project id must be specified")
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "01/02 15:04:05", NoColor: viper.GetBool(configLogNoColor)})

	InitConfigOrDie()

	logLevel, err := zerolog.ParseLevel(viper.GetString(configLogLevel))
	if err != nil {
		log.Fatal().Str("level", viper.GetString(configLogLevel)).Msg("invalid log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	dataDir := os.ExpandEnv(viper.GetString(configDataDir))
	log.Info().Str("data-dir", dataDir).Msg("got data directory")

	lookupClient, err := getLookupService()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize lookup client")
	}

	dumpServer, err := server.NewDumpServer(dataDir, lookupClient, viper.GetInt(configMaxMessageAgeHours))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create data store")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt(configPort)))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterDMSDumpServiceServer(grpcServer, dumpServer)
	log.Info().Int("port", viper.GetInt(configPort)).Msg("server is up and running")
	grpcServer.Serve(lis)
}

func getLookupService() (pb.LookupServiceClient, error) {
	nodeURL, err := bc.GetNodeURL(viper.GetString(configNetwork), viper.GetString(configInfuraProjectId))
	if err != nil {
		return nil, fmt.Errorf("failed to get network URL: %w", err)
	}
	log.Debug().Str("node-url", nodeURL).Msg("using blockchain node")

	contractAddress, err := bc.GetContractAddress(viper.GetString(configNetwork), viper.GetString(configContractAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to get contract address: %w", err)
	}
	log.Debug().Str("contract-address", contractAddress).Msg("using contract")

	lookup, err := bc.NewBlockchain(nodeURL, contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain node: %w", err)
	}

	return lookup, nil
}
