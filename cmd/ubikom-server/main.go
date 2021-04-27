package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
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
	var args CmdArgs
	flag.IntVar(&args.Port, "port", defaultPort, "port to listen to")
	flag.StringVar(&args.BaseDir, "base-dir", "", "base directory")
	flag.Parse()

	dbDir, err := getDBDir(args.BaseDir)
	if err != nil {
		log.Fatal(err)
	}
	db, err := badger.Open(badger.DefaultOptions(dbDir))
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", args.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterIdentityServiceServer(grpcServer, server.NewServer(db))
	log.Printf("listening on port %d...", args.Port)
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
