// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	filesystem "github.com/hostfactor/api/go/blueprint/filesystem"
	mock "github.com/stretchr/testify/mock"

	provideractions "github.com/hostfactor/diazo/pkg/provideractions"

	reaction "github.com/hostfactor/api/go/blueprint/reaction"
)

// FileReactionConditionBuilder is an autogenerated mock type for the FileReactionConditionBuilder type
type FileReactionConditionBuilder struct {
	mock.Mock
}

// Build provides a mock function with given fields:
func (_m *FileReactionConditionBuilder) Build() *reaction.FileReactionCondition {
	ret := _m.Called()

	var r0 *reaction.FileReactionCondition
	if rf, ok := ret.Get(0).(func() *reaction.FileReactionCondition); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*reaction.FileReactionCondition)
		}
	}

	return r0
}

// Directories provides a mock function with given fields: d
func (_m *FileReactionConditionBuilder) Directories(d ...string) provideractions.FileReactionConditionBuilder {
	_va := make([]interface{}, len(d))
	for _i := range d {
		_va[_i] = d[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 provideractions.FileReactionConditionBuilder
	if rf, ok := ret.Get(0).(func(...string) provideractions.FileReactionConditionBuilder); ok {
		r0 = rf(d...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(provideractions.FileReactionConditionBuilder)
		}
	}

	return r0
}

// DoesntMatch provides a mock function with given fields: fm
func (_m *FileReactionConditionBuilder) DoesntMatch(fm *filesystem.FileMatcher) provideractions.FileReactionConditionBuilder {
	ret := _m.Called(fm)

	var r0 provideractions.FileReactionConditionBuilder
	if rf, ok := ret.Get(0).(func(*filesystem.FileMatcher) provideractions.FileReactionConditionBuilder); ok {
		r0 = rf(fm)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(provideractions.FileReactionConditionBuilder)
		}
	}

	return r0
}

// Matches provides a mock function with given fields: fm
func (_m *FileReactionConditionBuilder) Matches(fm *filesystem.FileMatcher) provideractions.FileReactionConditionBuilder {
	ret := _m.Called(fm)

	var r0 provideractions.FileReactionConditionBuilder
	if rf, ok := ret.Get(0).(func(*filesystem.FileMatcher) provideractions.FileReactionConditionBuilder); ok {
		r0 = rf(fm)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(provideractions.FileReactionConditionBuilder)
		}
	}

	return r0
}

// Op provides a mock function with given fields: op
func (_m *FileReactionConditionBuilder) Op(op ...reaction.FileReactionCondition_FileOp) provideractions.FileReactionConditionBuilder {
	_va := make([]interface{}, len(op))
	for _i := range op {
		_va[_i] = op[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 provideractions.FileReactionConditionBuilder
	if rf, ok := ret.Get(0).(func(...reaction.FileReactionCondition_FileOp) provideractions.FileReactionConditionBuilder); ok {
		r0 = rf(op...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(provideractions.FileReactionConditionBuilder)
		}
	}

	return r0
}

type mockConstructorTestingTNewFileReactionConditionBuilder interface {
	mock.TestingT
	Cleanup(func())
}

// NewFileReactionConditionBuilder creates a new instance of FileReactionConditionBuilder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFileReactionConditionBuilder(t mockConstructorTestingTNewFileReactionConditionBuilder) *FileReactionConditionBuilder {
	mock := &FileReactionConditionBuilder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
