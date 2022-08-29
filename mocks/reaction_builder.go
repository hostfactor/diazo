// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	reaction "github.com/hostfactor/api/go/blueprint/reaction"
)

// reactionBuilder is an autogenerated mock type for the reactionBuilder type
type reactionBuilder struct {
	mock.Mock
}

// Build provides a mock function with given fields:
func (_m *reactionBuilder) Build() *reaction.Reaction {
	ret := _m.Called()

	var r0 *reaction.Reaction
	if rf, ok := ret.Get(0).(func() *reaction.Reaction); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*reaction.Reaction)
		}
	}

	return r0
}

type mockConstructorTestingTnewReactionBuilder interface {
	mock.TestingT
	Cleanup(func())
}

// newReactionBuilder creates a new instance of reactionBuilder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newReactionBuilder(t mockConstructorTestingTnewReactionBuilder) *reactionBuilder {
	mock := &reactionBuilder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
