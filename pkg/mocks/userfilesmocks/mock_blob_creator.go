// Code generated by mockery v2.36.0. DO NOT EDIT.

package userfilesmocks

import (
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// BlobCreator is an autogenerated mock type for the BlobCreator type
type BlobCreator struct {
	mock.Mock
}

type BlobCreator_Expecter struct {
	mock *mock.Mock
}

func (_m *BlobCreator) EXPECT() *BlobCreator_Expecter {
	return &BlobCreator_Expecter{mock: &_m.Mock}
}

// CreateBlob provides a mock function with given fields: fp
func (_m *BlobCreator) CreateBlob(fp string) (io.WriteCloser, error) {
	ret := _m.Called(fp)

	var r0 io.WriteCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (io.WriteCloser, error)); ok {
		return rf(fp)
	}
	if rf, ok := ret.Get(0).(func(string) io.WriteCloser); ok {
		r0 = rf(fp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.WriteCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(fp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlobCreator_CreateBlob_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateBlob'
type BlobCreator_CreateBlob_Call struct {
	*mock.Call
}

// CreateBlob is a helper method to define mock.On call
//   - fp string
func (_e *BlobCreator_Expecter) CreateBlob(fp interface{}) *BlobCreator_CreateBlob_Call {
	return &BlobCreator_CreateBlob_Call{Call: _e.mock.On("CreateBlob", fp)}
}

func (_c *BlobCreator_CreateBlob_Call) Run(run func(fp string)) *BlobCreator_CreateBlob_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *BlobCreator_CreateBlob_Call) Return(_a0 io.WriteCloser, _a1 error) *BlobCreator_CreateBlob_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BlobCreator_CreateBlob_Call) RunAndReturn(run func(string) (io.WriteCloser, error)) *BlobCreator_CreateBlob_Call {
	_c.Call.Return(run)
	return _c
}

// NewBlobCreator creates a new instance of BlobCreator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBlobCreator(t interface {
	mock.TestingT
	Cleanup(func())
}) *BlobCreator {
	mock := &BlobCreator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
