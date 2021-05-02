package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"teralyt.com/ubikom/pb"
)

type DumpServer struct {
	pb.UnimplementedDMSDumpServiceServer

	baseDir string
}

func NewDumpServer(baseDir string) *DumpServer {
	return &DumpServer{baseDir: baseDir}
}

func (s *DumpServer) Send(ctx context.Context, req *pb.DMSMessage) (*pb.Result, error) {
	// Verify signature.
	// Verify that sender's address is owned by the right public key.
	// Save message.
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (s *DumpServer) Receive(ctx context.Context, req *pb.Signed) (*pb.ResultWithContent, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Receive not implemented")
}
