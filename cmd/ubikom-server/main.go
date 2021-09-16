package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	defaultPort        = 8825
	defaultHomeSubDir  = ".ubikom"
	defaultDBSubDir    = "db"
	defaultPowStrength = 10
)

type HealthChecker struct{}

func (h *HealthChecker) Check(ctx context.Context,
	req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	log.Debug().Msg("health check")
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
	log.Debug().Msg("streaming health check")
	srv.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
	<-srv.Context().Done()
	return nil
}

type CmdArgs struct {
	BaseDir     string
	Port        int
	PowStrength int
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", defaultPort, "port to listen to")
	flag.StringVar(&args.BaseDir, "base-dir", "", "base directory")
	flag.IntVar(&args.PowStrength, "pow-strength", defaultPowStrength, "POW strength required")
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

	// Initialize health checker.
	healthService := &HealthChecker{}
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)

	srv := server.NewServer(db, args.PowStrength)
	pb.RegisterIdentityServiceServer(grpcServer, srv)
	pb.RegisterLookupServiceServer(grpcServer, srv)
	log.Info().Int("port", args.Port).Msg("server is up and running")
	grpcServer.Serve(lis)
}

func getDBDir(baseDir string) (string, error) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error retrieving home directory: %w", err)
		}
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
