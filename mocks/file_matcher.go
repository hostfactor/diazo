// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	filesystem "github.com/hostfactor/api/go/blueprint/filesystem"
	mock "github.com/stretchr/testify/mock"
)

// FileMatcher is an autogenerated mock type for the FileMatcher type
type FileMatcher struct {
	mock.Mock
}

// FileMatcher provides a mock function with given fields:
func (_m *FileMatcher) FileMatcher() *filesystem.FileMatcher {
	ret := _m.Called()

	var r0 *filesystem.FileMatcher
	if rf, ok := ret.Get(0).(func() *filesystem.FileMatcher); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*filesystem.FileMatcher)
		}
	}

	return r0
}