// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	fsnotify "github.com/fsnotify/fsnotify"
	mock "github.com/stretchr/testify/mock"
)

// WatchFileFunc is an autogenerated mock type for the WatchFileFunc type
type WatchFileFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields: event
func (_m *WatchFileFunc) Execute(event fsnotify.Event) {
	_m.Called(event)
}

type mockConstructorTestingTNewWatchFileFunc interface {
	mock.TestingT
	Cleanup(func())
}

// NewWatchFileFunc creates a new instance of WatchFileFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewWatchFileFunc(t mockConstructorTestingTNewWatchFileFunc) *WatchFileFunc {
	mock := &WatchFileFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}