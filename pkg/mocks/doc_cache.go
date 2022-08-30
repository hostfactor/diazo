// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	blueprint "github.com/hostfactor/api/go/blueprint"

	mock "github.com/stretchr/testify/mock"
)

// DocCache is an autogenerated mock type for the DocCache type
type DocCache struct {
	mock.Mock
}

// Get provides a mock function with given fields:
func (_m *DocCache) Get() *blueprint.Docs {
	ret := _m.Called()

	var r0 *blueprint.Docs
	if rf, ok := ret.Get(0).(func() *blueprint.Docs); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blueprint.Docs)
		}
	}

	return r0
}

type mockConstructorTestingTNewDocCache interface {
	mock.TestingT
	Cleanup(func())
}

// NewDocCache creates a new instance of DocCache. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDocCache(t mockConstructorTestingTNewDocCache) *DocCache {
	mock := &DocCache{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}