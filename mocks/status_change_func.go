// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	actions "github.com/hostfactor/api/go/blueprint/actions"
	mock "github.com/stretchr/testify/mock"
)

// StatusChangeFunc is an autogenerated mock type for the StatusChangeFunc type
type StatusChangeFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields: s
func (_m *StatusChangeFunc) Execute(s actions.SetStatus_Status) {
	_m.Called(s)
}

type mockConstructorTestingTNewStatusChangeFunc interface {
	mock.TestingT
	Cleanup(func())
}

// NewStatusChangeFunc creates a new instance of StatusChangeFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStatusChangeFunc(t mockConstructorTestingTNewStatusChangeFunc) *StatusChangeFunc {
	mock := &StatusChangeFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}