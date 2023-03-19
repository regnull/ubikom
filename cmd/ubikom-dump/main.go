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
	defaultHomeSubDir         = ".ubikom"
	defaultDataSubDir         = "dump"
	defaultDataDir            = "$HOME/.ubikom/dump"
	defaultMaxMessageAgeHours = 14 * 24
	defaultNetwork            = "main"
	defaultLogLevel           = "info"
	defaultLogNoColor         = false
	envPrefix                 = "UBK_"
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

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	viper.SetDefault("port", defaultPort)
	viper.SetDefault("data-dir", "$HOME/.ubikom/dump")
	viper.SetDefault("max-message-age-hours", defaultMaxMessageAgeHours)
	viper.SetDefault("network", defaultNetwork)
	viper.SetDefault("log-level", defaultLogLevel)
	viper.SetDefault("log-no-color", defaultLogNoColor)
	viper.SetDefault("contract-address", globals.MainnetNameRegistryAddress)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", 0, "port to listen to")
	flag.StringVar(&args.DataDir, "data-dir", "", "base directory")
	flag.IntVar(&args.MaxMessageAgeHours, "max-message-age-hours", 0, "max message age, in hours")
	flag.StringVar(&args.Network, "network", "", "ethereum network to use")
	flag.StringVar(&args.InfuraProjectId, "infura-project-id", "", "infura project id")
	flag.StringVar(&args.ContractAddress, "contract-address", "", "name registry contract address")
	flag.StringVar(&args.LogLevel, "log-level", "", "log level")
	flag.BoolVar(&args.LogNoColor, "log-no-color", false, "disable colors for logging")
	flag.StringVar(&args.ConfigFile, "config", "", "config file location")
	flag.Parse()

	viper.BindEnv("network", "UBK_NETWORK")
	viper.BindEnv("infura-project-id", "UBK_INFURA_PROJECT_ID")
	viper.BindEnv("contract-address", "UBK_CONTRACT_ADDRESS")
	viper.BindEnv("log-level", "UBK_LOG_LEVEL")
	viper.BindEnv("log-no-color", "UBK_LOG_NO_COLOR")

	if args.ConfigFile != "" {
		viper.SetConfigFile(args.ConfigFile)
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Str("path", args.ConfigFile).Msg("failed to read config file")
		}
	}

	viper.BindPFlags(flag.CommandLine)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "01/02 15:04:05", NoColor: viper.GetBool("log-no-color")})

	logLevel, err := zerolog.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.Fatal().Str("level", viper.GetString("log-level")).Msg("invalid log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	dataDir := os.ExpandEnv(viper.GetString("data-dir"))
	log.Info().Str("data-dir", dataDir).Msg("got data directory")

	lookupClient, err := getLookupService()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize lookup client")
	}

	dumpServer, err := server.NewDumpServer(dataDir, lookupClient, viper.GetInt("max-message-age-hours"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create data store")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt("port")))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterDMSDumpServiceServer(grpcServer, dumpServer)
	log.Info().Int("port", viper.GetInt("port")).Msg("server is up and running")
	grpcServer.Serve(lis)
}

func getLookupService() (pb.LookupServiceClient, error) {
	nodeURL, err := bc.GetNodeURL(viper.GetString("network"), viper.GetString("infura-project-id"))
	if err != nil {
		return nil, fmt.Errorf("failed to get network URL: %w", err)
	}
	log.Debug().Str("node-url", nodeURL).Msg("using blockchain node")

	contractAddress, err := bc.GetContractAddress(viper.GetString("network"), viper.GetString("contract-address"))
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
