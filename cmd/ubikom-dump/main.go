package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/globals"
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
	DataDir                string
	Port                   int
	LookupServerURL        string
	MaxMessageAgeHours     int
	BlockchainNodeURL      string
	UseLegacyLookupService bool
	Network                string
	InfuraProjectId        string
	ContractAddress        string
	LogLevel               string
	LogNoColor             bool
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", defaultPort, "port to listen to")
	flag.StringVar(&args.DataDir, "data-dir", "", "base directory")
	flag.StringVar(&args.LookupServerURL, "lookup-server-url", defaultIdentityServerURL, "DEPRECATED: URL of the lookup server")
	flag.IntVar(&args.MaxMessageAgeHours, "max-message-age-hours", defaultMaxMessageAgeHours, "max message age, in hours")
	flag.StringVar(&args.BlockchainNodeURL, "blockchain-node-url", globals.BlockchainNodeURL, "DEPRECATES: blockchain node url (use network flag instead)")
	flag.BoolVar(&args.UseLegacyLookupService, "use-legacy-lookup-service", false, "DEPRECATED: use legacy lookup service")
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

	lookupClient, cleanup, err := getLookupService(&args)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize lookup client")
	}
	defer cleanup()

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

func connectToLookupService(url string) (pb.LookupServiceClient, *grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}

	conn, err := grpc.Dial(url, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to lookup service: %w", err)
	}

	return pb.NewLookupServiceClient(conn), conn, nil
}

func getLookupService(args *CmdArgs) (pb.LookupServiceClient, func(), error) {
	// We always use the new blockchain lookup service as the first priority.
	// If arguments for the legacy lookup service are specified, we will use
	// them as fallback.

	nodeURL, err := bc.GetNodeURL(args.Network, args.InfuraProjectId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get network URL: %w", err)
	}
	log.Debug().Str("node-url", nodeURL).Msg("using blockchain node")

	contractAddress, err := bc.GetContractAddress(args.Network, args.ContractAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get contract address: %w", err)
	}
	log.Debug().Str("contract-address", contractAddress).Msg("using contract")

	// This is our main blockchain-based lookup service. Eventually, it will be the only one.
	// For now, we will fallback on the existing old-style blockchain lookup service, or
	// standalone lookup service. Those will go away.
	blockchainV2Lookup, err := bc.NewBlockchainV2(nodeURL, contractAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to blockchain node: %w", err)
	}

	if args.LookupServerURL == "" || args.BlockchainNodeURL == "" {
		return blockchainV2Lookup, nil, nil
	}

	// Standalone lookup service - to be deprecated.
	log.Warn().Str("url", args.LookupServerURL).Msg("using legacy lookup service")
	lookupService, conn, err := connectToLookupService(args.LookupServerURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to lookup server")
	}

	// Old-style blockchain lookup service - to be deprecated.
	log.Warn().Str("url", args.BlockchainNodeURL).Msg("using legacy blockchain")
	client, err := ethclient.Dial(args.BlockchainNodeURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to blockchain node")
	}
	blockchain := bc.NewBlockchain(client, globals.KeyRegistryContractAddress,
		globals.NameRegistryContractAddress, globals.ConnectorRegistryContractAddress, nil)

	var combinedLookupClient pb.LookupServiceClient
	if args.UseLegacyLookupService {
		log.Info().Msg("using legacy lookup service")
		combinedLookupClient = lookupService
	} else {
		combinedLookupClient = bc.NewLookupServiceClient(blockchain, lookupService, false)
	}

	// For now, we use old lookup service as a fallback.
	combinedLookupClient = bc.NewLookupServiceV2(blockchainV2Lookup, combinedLookupClient)
	return combinedLookupClient, func() { conn.Close() }, nil
}
