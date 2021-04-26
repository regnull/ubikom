package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/server"
)

const (
	defaultPort = 8825
)

type CmdArgs struct {
	Port int
}

func main() {
	var args CmdArgs
	flag.IntVar(&args.Port, "port", defaultPort, "port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", args.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterIdentityServiceServer(grpcServer, &server.Server{})
	log.Printf("listening on port %d...", args.Port)
	grpcServer.Serve(lis)
}
