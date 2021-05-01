package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"teralyt.com/ubikom/pb"
)

type DumpServer struct {
	pb.UnimplementedDMSDumpServiceServer
}

func (s *DumpServer) Send(ctx context.Context, req *pb.DMSMessage) (*pb.Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (s *DumpServer) Receive(ctx context.Context, req *pb.Signed) (*pb.ResultWithContent, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Receive not implemented")
}
