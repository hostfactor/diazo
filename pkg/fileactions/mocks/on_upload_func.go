// Code generated by mockery 2.9.0. DO NOT EDIT.

package mocks

import (
	fileactions "github.com/hostfactor/diazo/pkg/fileactions"
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
