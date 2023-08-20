// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	pb "github.com/regnull/ubikom/pb"
)

// MockDMSDumpServiceClient is an autogenerated mock type for the DMSDumpServiceClient type
type MockDMSDumpServiceClient struct {
	mock.Mock
}

type MockDMSDumpServiceClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDMSDumpServiceClient) EXPECT() *MockDMSDumpServiceClient_Expecter {
	return &MockDMSDumpServiceClient_Expecter{mock: &_m.Mock}
}

// Receive provides a mock function with given fields: ctx, in, opts
func (_m *MockDMSDumpServiceClient) Receive(ctx context.Context, in *pb.ReceiveRequest, opts ...grpc.CallOption) (*pb.ReceiveResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *pb.ReceiveResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *pb.ReceiveRequest, ...grpc.CallOption) (*pb.ReceiveResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *pb.ReceiveRequest, ...grpc.CallOption) *pb.ReceiveResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pb.ReceiveResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *pb.ReceiveRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDMSDumpServiceClient_Receive_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Receive'
type MockDMSDumpServiceClient_Receive_Call struct {
	*mock.Call
}

// Receive is a helper method to define mock.On call
//   - ctx context.Context
//   - in *pb.ReceiveRequest
//   - opts ...grpc.CallOption
func (_e *MockDMSDumpServiceClient_Expecter) Receive(ctx interface{}, in interface{}, opts ...interface{}) *MockDMSDumpServiceClient_Receive_Call {
	return &MockDMSDumpServiceClient_Receive_Call{Call: _e.mock.On("Receive",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *MockDMSDumpServiceClient_Receive_Call) Run(run func(ctx context.Context, in *pb.ReceiveRequest, opts ...grpc.CallOption)) *MockDMSDumpServiceClient_Receive_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*pb.ReceiveRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockDMSDumpServiceClient_Receive_Call) Return(_a0 *pb.ReceiveResponse, _a1 error) *MockDMSDumpServiceClient_Receive_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDMSDumpServiceClient_Receive_Call) RunAndReturn(run func(context.Context, *pb.ReceiveRequest, ...grpc.CallOption) (*pb.ReceiveResponse, error)) *MockDMSDumpServiceClient_Receive_Call {
	_c.Call.Return(run)
	return _c
}

// Send provides a mock function with given fields: ctx, in, opts
func (_m *MockDMSDumpServiceClient) Send(ctx context.Context, in *pb.SendRequest, opts ...grpc.CallOption) (*pb.SendResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *pb.SendResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *pb.SendRequest, ...grpc.CallOption) (*pb.SendResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *pb.SendRequest, ...grpc.CallOption) *pb.SendResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pb.SendResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *pb.SendRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDMSDumpServiceClient_Send_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Send'
type MockDMSDumpServiceClient_Send_Call struct {
	*mock.Call
}

// Send is a helper method to define mock.On call
//   - ctx context.Context
//   - in *pb.SendRequest
//   - opts ...grpc.CallOption
func (_e *MockDMSDumpServiceClient_Expecter) Send(ctx interface{}, in interface{}, opts ...interface{}) *MockDMSDumpServiceClient_Send_Call {
	return &MockDMSDumpServiceClient_Send_Call{Call: _e.mock.On("Send",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *MockDMSDumpServiceClient_Send_Call) Run(run func(ctx context.Context, in *pb.SendRequest, opts ...grpc.CallOption)) *MockDMSDumpServiceClient_Send_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*pb.SendRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockDMSDumpServiceClient_Send_Call) Return(_a0 *pb.SendResponse, _a1 error) *MockDMSDumpServiceClient_Send_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDMSDumpServiceClient_Send_Call) RunAndReturn(run func(context.Context, *pb.SendRequest, ...grpc.CallOption) (*pb.SendResponse, error)) *MockDMSDumpServiceClient_Send_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDMSDumpServiceClient creates a new instance of MockDMSDumpServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDMSDumpServiceClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDMSDumpServiceClient {
	mock := &MockDMSDumpServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
