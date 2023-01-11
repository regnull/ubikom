package lookup

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

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

func Get(network string, projectId string, contractAddress string,
	legacyLookupServerUrl string, legacyNodeUrl string,
	useLegacyLookupService bool) (pb.LookupServiceClient, func(), error) {
	// We always use the new blockchain lookup service as the first priority.
	// If arguments for the legacy lookup service are specified, we will use
	// them as fallback.

	nodeURL, err := bc.GetNodeURL(network, projectId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get network URL: %w", err)
	}
	log.Debug().Str("node-url", nodeURL).Msg("using blockchain node")

	cntrAddress, err := bc.GetContractAddress(network, contractAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get contract address: %w", err)
	}
	log.Debug().Str("contract-address", cntrAddress).Msg("using contract")

	// This is our main blockchain-based lookup service. Eventually, it will be the only one.
	// For now, we will fallback on the existing old-style blockchain lookup service, or
	// standalone lookup service. Those will go away.
	blockchainV2Lookup, err := bc.NewBlockchainV2(nodeURL, cntrAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to blockchain node: %w", err)
	}

	if legacyLookupServerUrl == "" || legacyNodeUrl == "" {
		return blockchainV2Lookup, nil, nil
	}

	// Standalone lookup service - to be deprecated.
	log.Warn().Str("url", legacyLookupServerUrl).Msg("using legacy lookup service")
	lookupService, conn, err := connectToLookupService(legacyLookupServerUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to lookup server")
	}

	// Old-style blockchain lookup service - to be deprecated.
	log.Warn().Str("url", legacyNodeUrl).Msg("using legacy blockchain")
	client, err := ethclient.Dial(legacyNodeUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to blockchain node")
	}
	blockchain := bc.NewBlockchain(client, globals.KeyRegistryContractAddress,
		globals.NameRegistryContractAddress, globals.ConnectorRegistryContractAddress, nil)

	var combinedLookupClient pb.LookupServiceClient
	if useLegacyLookupService {
		log.Info().Msg("using legacy lookup service")
		combinedLookupClient = lookupService
	} else {
		combinedLookupClient = bc.NewLookupServiceClient(blockchain, lookupService, false)
	}

	// For now, we use old lookup service as a fallback.
	combinedLookupClient = bc.NewLookupServiceV2(blockchainV2Lookup, combinedLookupClient)
	return combinedLookupClient, func() { conn.Close() }, nil
}
