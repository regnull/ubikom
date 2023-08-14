package main

import (
	"fmt"
	"net"
	"os"

	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/cfg"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func main() {
	err := cfg.InitConfig([]cfg.ConfigEntry{
		cfg.NewIntConfig("port", 8826, "port to listen to", ""),
		cfg.NewStringConfig("data-dir", "$HOME/.ubikom/dump", "data directory", ""),
		cfg.NewIntConfig("max-message-age-hours", 24*14, "max message age in hours", ""),
		cfg.NewStringConfig("network", "main", "ethereum network to use", "UBK_NETWORK"),
		cfg.NewStringConfig("infura-project-id", "", "infura project id", "INFURA_PROJECT_ID"),
		cfg.NewStringConfig("contract-address", "", "contract address", "UBK_CONTRACT_ADDRESS"),
		cfg.NewStringConfig("log-level", "info", "log level", "UBK_LOG_LEVEL"),
		cfg.NewBoolConfig("log-no-color", false, "disable colors for logging", "UBK_LOG_NO_COLOR"),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	if viper.GetString("infura-project-id") == "" {
		log.Fatal().Msg("infura project id must be specified")
	}

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

func getLookupService() (*bc.Blockchain, error) {
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
