// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	fileactions "github.com/hostfactor/diazo/pkg/actions/fileactions"
	mock "github.com/stretchr/testify/mock"
)

// OnUploadFunc is an autogenerated mock type for the OnUploadFunc type
type OnUploadFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields: params
func (_m *OnUploadFunc) Execute(params fileactions.OnUploadFuncParams) {
	_m.Called(params)
}

type mockConstructorTestingTNewOnUploadFunc interface {
	mock.TestingT
	Cleanup(func())
}

// NewOnUploadFunc creates a new instance of OnUploadFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOnUploadFunc(t mockConstructorTestingTNewOnUploadFunc) *OnUploadFunc {
	mock := &OnUploadFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
