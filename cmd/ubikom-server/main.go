package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/server"
)

const (
	defaultPort       = 8825
	defaultHomeSubDir = ".ubikom"
	defaultDBSubDir   = "db"
)

type CmdArgs struct {
	BaseDir string
	Port    int
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", defaultPort, "port to listen to")
	flag.StringVar(&args.BaseDir, "base-dir", "", "base directory")
	flag.Parse()

	dbDir, err := getDBDir(args.BaseDir)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get database directory")
	}
	db, err := badger.Open(badger.DefaultOptions(dbDir))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize the database")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", args.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	srv := server.NewServer(db)
	pb.RegisterIdentityServiceServer(grpcServer, srv)
	pb.RegisterLookupServiceServer(grpcServer, srv)
	log.Info().Int("port", args.Port).Msg("server is up and running")
	grpcServer.Serve(lis)
}

func getDBDir(baseDir string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error retrieving home directory: %w", err)
	}
	if baseDir == "" {
		dir := path.Join(homeDir, defaultHomeSubDir)
		_ = os.Mkdir(dir, 0700)
		dir = path.Join(dir, defaultDBSubDir)
		_ = os.Mkdir(dir, 0700)
		return dir, nil
	}
	dbDir := path.Join(baseDir, defaultDBSubDir)
	_ = os.Mkdir(dbDir, 0700)
	return dbDir, nil
}
