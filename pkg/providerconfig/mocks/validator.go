// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	blueprint "github.com/hostfactor/api/go/blueprint"
	mock "github.com/stretchr/testify/mock"
)

// Validator is an autogenerated mock type for the Validator type
type Validator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: val
func (_m *Validator) Validate(val *blueprint.Value) error {
	ret := _m.Called(val)

	var r0 error
	if rf, ok := ret.Get(0).(func(*blueprint.Value) error); ok {
		r0 = rf(val)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewValidator interface {
	mock.TestingT
	Cleanup(func())
}

// NewValidator creates a new instance of Validator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewValidator(t mockConstructorTestingTNewValidator) *Validator {
	mock := &Validator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}