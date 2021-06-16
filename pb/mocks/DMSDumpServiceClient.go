// Code generated by mockery 2.7.5. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	pb "github.com/regnull/ubikom/pb"
)

// DMSDumpServiceClient is an autogenerated mock type for the DMSDumpServiceClient type
type DMSDumpServiceClient struct {
	mock.Mock
}

// Receive provides a mock function with given fields: ctx, in, opts
func (_m *DMSDumpServiceClient) Receive(ctx context.Context, in *pb.ReceiveRequest, opts ...grpc.CallOption) (*pb.ReceiveResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *pb.ReceiveResponse
	if rf, ok := ret.Get(0).(func(context.Context, *pb.ReceiveRequest, ...grpc.CallOption) *pb.ReceiveResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pb.ReceiveResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *pb.ReceiveRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Send provides a mock function with given fields: ctx, in, opts
func (_m *DMSDumpServiceClient) Send(ctx context.Context, in *pb.SendRequest, opts ...grpc.CallOption) (*pb.SendResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *pb.SendResponse
	if rf, ok := ret.Get(0).(func(context.Context, *pb.SendRequest, ...grpc.CallOption) *pb.SendResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pb.SendResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *pb.SendRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}