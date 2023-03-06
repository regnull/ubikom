package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	defaultPort               = 8826
	defaultIdentityServerURL  = "localhost:8825"
	defaultHomeSubDir         = ".ubikom"
	defaultDataSubDir         = "dump"
	defaultMaxMessageAgeHours = 14 * 24
	defaultNetwork            = "main"
	defaultLogLevel           = "info"
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
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", defaultPort, "port to listen to")
	flag.StringVar(&args.DataDir, "data-dir", "", "base directory")
	flag.IntVar(&args.MaxMessageAgeHours, "max-message-age-hours", defaultMaxMessageAgeHours, "max message age, in hours")
	flag.StringVar(&args.Network, "network", defaultNetwork, "ethereum network to use")
	flag.StringVar(&args.InfuraProjectId, "infura-project-id", "", "infura project id")
	flag.StringVar(&args.ContractAddress, "contract-address", "", "name registry contract address")
	flag.StringVar(&args.LogLevel, "log-level", defaultLogLevel, "log level")
	flag.BoolVar(&args.LogNoColor, "log-no-color", false, "disable colors for logging")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "01/02 15:04:05", NoColor: args.LogNoColor})

	// Set the log level.
	logLevel, err := zerolog.ParseLevel(args.LogLevel)
	if err != nil {
		log.Fatal().Str("level", args.LogLevel).Msg("invalid log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	if args.DataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get home directory")
		}
		dir := path.Join(homeDir, defaultHomeSubDir)
		_ = os.Mkdir(dir, 0770)
		dataDir := path.Join(dir, defaultDataSubDir)
		_ = os.Mkdir(dataDir, 0770)
		args.DataDir = dataDir
	}
	log.Info().Str("data-dir", args.DataDir).Msg("got data directory")

	lookupClient, err := getLookupService(&args)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize lookup client")
	}

	dumpServer, err := server.NewDumpServer(args.DataDir, lookupClient, args.MaxMessageAgeHours)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create data store")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", args.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterDMSDumpServiceServer(grpcServer, dumpServer)
	log.Info().Int("port", args.Port).Msg("server is up and running")
	grpcServer.Serve(lis)
}

func getLookupService(args *CmdArgs) (pb.LookupServiceClient, error) {
	// We always use the new blockchain lookup service as the first priority.
	// If arguments for the legacy lookup service are specified, we will use
	// them as fallback.

	nodeURL, err := bc.GetNodeURL(args.Network, args.InfuraProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get network URL: %w", err)
	}
	log.Debug().Str("node-url", nodeURL).Msg("using blockchain node")

	contractAddress, err := bc.GetContractAddress(args.Network, args.ContractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract address: %w", err)
	}
	log.Debug().Str("contract-address", contractAddress).Msg("using contract")

	// This is our main blockchain-based lookup service. Eventually, it will be the only one.
	// For now, we will fallback on the existing old-style blockchain lookup service, or
	// standalone lookup service. Those will go away.
	blockchainV2Lookup, err := bc.NewBlockchainV2(nodeURL, contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain node: %w", err)
	}

	return blockchainV2Lookup, nil
}
