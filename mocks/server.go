// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Server is an autogenerated mock type for the Server type
type Server struct {
	mock.Mock
}

// ListenAndServe provides a mock function with given fields:
func (_m *Server) ListenAndServe() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Shutdown provides a mock function with given fields: ctx
func (_m *Server) Shutdown(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewServer interface {
	mock.TestingT
	Cleanup(func())
}

// NewServer creates a new instance of Server. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewServer(t mockConstructorTestingTNewServer) *Server {
	mock := &Server{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
