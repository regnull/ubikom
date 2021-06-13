// Code generated by mockery 2.7.5. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// ExternalSender is an autogenerated mock type for the ExternalSender type
type ExternalSender struct {
	mock.Mock
}

// Send provides a mock function with given fields: from, message
func (_m *ExternalSender) Send(from string, message string) error {
	ret := _m.Called(from, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(from, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
