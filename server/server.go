package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"teralyt.com/ubikom/pb"
)

type Server struct {
	pb.UnimplementedIdentityServiceServer
}

func (s *Server) RegisterKey(context.Context, *pb.KeyRegistrationRequest) (*pb.KeyRegistrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterKey not implemented")
}
