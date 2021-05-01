package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	defaultPort              = 8826
	defaultIdentityServerURL = "localhost:8825"
)

type CmdArgs struct {
	DataDir           string
	Port              int
	IdentityServerURL string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", defaultPort, "port to listen to")
	flag.StringVar(&args.DataDir, "data-dir", "", "base directory")
	flag.StringVar(&args.IdentityServerURL, "identity-server-url", defaultIdentityServerURL, "URL of the identity server")
	flag.Parse()

	if args.DataDir == "" {
		log.Fatal().Msg("data directory must be specified")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", args.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	// pb.RegisterIdentityServiceServer(grpcServer, srv)
	// pb.RegisterLookupServiceServer(grpcServer, srv)
	log.Info().Int("port", args.Port).Msg("server is up and running")
	grpcServer.Serve(lis)
}
