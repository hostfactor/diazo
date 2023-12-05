// Code generated by mockery v2.36.0. DO NOT EDIT.

package userfilesmocks

import (
	userfiles "github.com/hostfactor/diazo/pkg/userfiles"
	mock "github.com/stretchr/testify/mock"
)

// UrlSigner is an autogenerated mock type for the UrlSigner type
type UrlSigner struct {
	mock.Mock
}

type UrlSigner_Expecter struct {
	mock *mock.Mock
}

func (_m *UrlSigner) EXPECT() *UrlSigner_Expecter {
	return &UrlSigner_Expecter{mock: &_m.Mock}
}

// SignedUrl provides a mock function with given fields: fileDesc, httpMethod, folder
func (_m *UrlSigner) SignedUrl(fileDesc userfiles.FileDesc, httpMethod string, folder string) (string, error) {
	ret := _m.Called(fileDesc, httpMethod, folder)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(userfiles.FileDesc, string, string) (string, error)); ok {
		return rf(fileDesc, httpMethod, folder)
	}
	if rf, ok := ret.Get(0).(func(userfiles.FileDesc, string, string) string); ok {
		r0 = rf(fileDesc, httpMethod, folder)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(userfiles.FileDesc, string, string) error); ok {
		r1 = rf(fileDesc, httpMethod, folder)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UrlSigner_SignedUrl_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SignedUrl'
type UrlSigner_SignedUrl_Call struct {
	*mock.Call
}

// SignedUrl is a helper method to define mock.On call
//   - fileDesc userfiles.FileDesc
//   - httpMethod string
//   - folder string
func (_e *UrlSigner_Expecter) SignedUrl(fileDesc interface{}, httpMethod interface{}, folder interface{}) *UrlSigner_SignedUrl_Call {
	return &UrlSigner_SignedUrl_Call{Call: _e.mock.On("SignedUrl", fileDesc, httpMethod, folder)}
}

func (_c *UrlSigner_SignedUrl_Call) Run(run func(fileDesc userfiles.FileDesc, httpMethod string, folder string)) *UrlSigner_SignedUrl_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(userfiles.FileDesc), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *UrlSigner_SignedUrl_Call) Return(_a0 string, _a1 error) *UrlSigner_SignedUrl_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *UrlSigner_SignedUrl_Call) RunAndReturn(run func(userfiles.FileDesc, string, string) (string, error)) *UrlSigner_SignedUrl_Call {
	_c.Call.Return(run)
	return _c
}

// NewUrlSigner creates a new instance of UrlSigner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUrlSigner(t interface {
	mock.TestingT
	Cleanup(func())
}) *UrlSigner {
	mock := &UrlSigner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
