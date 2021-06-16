package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"time"

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
)

type CmdArgs struct {
	DataDir            string
	Port               int
	LookupServerURL    string
	MaxMessageAgeHours int
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", defaultPort, "port to listen to")
	flag.StringVar(&args.DataDir, "data-dir", "", "base directory")
	flag.StringVar(&args.LookupServerURL, "lookup-server-url", defaultIdentityServerURL, "URL of the lookup server")
	flag.IntVar(&args.MaxMessageAgeHours, "max-message-age-hours", defaultMaxMessageAgeHours, "max message age, in hours")
	flag.Parse()

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

	lookupService, conn, err := connectToLookupService(args.LookupServerURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to lookup server")
	}
	defer conn.Close()

	dumpServer, err := server.NewDumpServer(args.DataDir, lookupService, args.MaxMessageAgeHours)
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
