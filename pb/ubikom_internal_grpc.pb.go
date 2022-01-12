// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ProxyServiceClient is the client API for ProxyService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProxyServiceClient interface {
	CopyMailboxes(ctx context.Context, in *CopyMailboxesRequest, opts ...grpc.CallOption) (*CopyMailboxesResponse, error)
}

type proxyServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProxyServiceClient(cc grpc.ClientConnInterface) ProxyServiceClient {
	return &proxyServiceClient{cc}
}

func (c *proxyServiceClient) CopyMailboxes(ctx context.Context, in *CopyMailboxesRequest, opts ...grpc.CallOption) (*CopyMailboxesResponse, error) {
	out := new(CopyMailboxesResponse)
	err := c.cc.Invoke(ctx, "/Ubikom.ProxyService/CopyMailboxes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProxyServiceServer is the server API for ProxyService service.
// All implementations must embed UnimplementedProxyServiceServer
// for forward compatibility
type ProxyServiceServer interface {
	CopyMailboxes(context.Context, *CopyMailboxesRequest) (*CopyMailboxesResponse, error)
	mustEmbedUnimplementedProxyServiceServer()
}

// UnimplementedProxyServiceServer must be embedded to have forward compatible implementations.
type UnimplementedProxyServiceServer struct {
}

func (*UnimplementedProxyServiceServer) CopyMailboxes(context.Context, *CopyMailboxesRequest) (*CopyMailboxesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CopyMailboxes not implemented")
}
func (*UnimplementedProxyServiceServer) mustEmbedUnimplementedProxyServiceServer() {}

func RegisterProxyServiceServer(s *grpc.Server, srv ProxyServiceServer) {
	s.RegisterService(&_ProxyService_serviceDesc, srv)
}

func _ProxyService_CopyMailboxes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CopyMailboxesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProxyServiceServer).CopyMailboxes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Ubikom.ProxyService/CopyMailboxes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProxyServiceServer).CopyMailboxes(ctx, req.(*CopyMailboxesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ProxyService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Ubikom.ProxyService",
	HandlerType: (*ProxyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CopyMailboxes",
			Handler:    _ProxyService_CopyMailboxes_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ubikom_internal.proto",
}