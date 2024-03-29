// Code generated by mockery v2.36.0. DO NOT EDIT.

package actionsmocks

import (
	actions "github.com/hostfactor/diazo/pkg/actions"
	mock "github.com/stretchr/testify/mock"
)

// OnDownloadFunc is an autogenerated mock type for the OnDownloadFunc type
type OnDownloadFunc struct {
	mock.Mock
}

type OnDownloadFunc_Expecter struct {
	mock *mock.Mock
}

func (_m *OnDownloadFunc) EXPECT() *OnDownloadFunc_Expecter {
	return &OnDownloadFunc_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: params
func (_m *OnDownloadFunc) Execute(params actions.OnDownloadFuncParams) {
	_m.Called(params)
}

// OnDownloadFunc_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type OnDownloadFunc_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - params actions.OnDownloadFuncParams
func (_e *OnDownloadFunc_Expecter) Execute(params interface{}) *OnDownloadFunc_Execute_Call {
	return &OnDownloadFunc_Execute_Call{Call: _e.mock.On("Execute", params)}
}

func (_c *OnDownloadFunc_Execute_Call) Run(run func(params actions.OnDownloadFuncParams)) *OnDownloadFunc_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(actions.OnDownloadFuncParams))
	})
	return _c
}

func (_c *OnDownloadFunc_Execute_Call) Return() *OnDownloadFunc_Execute_Call {
	_c.Call.Return()
	return _c
}

func (_c *OnDownloadFunc_Execute_Call) RunAndReturn(run func(actions.OnDownloadFuncParams)) *OnDownloadFunc_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewOnDownloadFunc creates a new instance of OnDownloadFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOnDownloadFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *OnDownloadFunc {
	mock := &OnDownloadFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
