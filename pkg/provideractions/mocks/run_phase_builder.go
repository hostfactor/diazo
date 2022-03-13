// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	blueprint "github.com/hostfactor/api/go/blueprint"
	mock "github.com/stretchr/testify/mock"

	provideractions "github.com/hostfactor/diazo/pkg/provideractions"
)

// RunPhaseBuilder is an autogenerated mock type for the RunPhaseBuilder type
type RunPhaseBuilder struct {
	mock.Mock
}

// Build provides a mock function with given fields:
func (_m *RunPhaseBuilder) Build() *blueprint.RunPhase {
	ret := _m.Called()

	var r0 *blueprint.RunPhase
	if rf, ok := ret.Get(0).(func() *blueprint.RunPhase); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blueprint.RunPhase)
		}
	}

	return r0
}

// Gid provides a mock function with given fields: i
func (_m *RunPhaseBuilder) Gid(i int) provideractions.RunPhaseBuilder {
	ret := _m.Called(i)

	var r0 provideractions.RunPhaseBuilder
	if rf, ok := ret.Get(0).(func(int) provideractions.RunPhaseBuilder); ok {
		r0 = rf(i)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(provideractions.RunPhaseBuilder)
		}
	}

	return r0
}

// Uid provides a mock function with given fields: i
func (_m *RunPhaseBuilder) Uid(i int) provideractions.RunPhaseBuilder {
	ret := _m.Called(i)

	var r0 provideractions.RunPhaseBuilder
	if rf, ok := ret.Get(0).(func(int) provideractions.RunPhaseBuilder); ok {
		r0 = rf(i)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(provideractions.RunPhaseBuilder)
		}
	}

	return r0
}

// When provides a mock function with given fields: when
func (_m *RunPhaseBuilder) When(when ...provideractions.FileTriggerConditionBuilder) provideractions.FileTriggerBuilder {
	_va := make([]interface{}, len(when))
	for _i := range when {
		_va[_i] = when[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 provideractions.FileTriggerBuilder
	if rf, ok := ret.Get(0).(func(...provideractions.FileTriggerConditionBuilder) provideractions.FileTriggerBuilder); ok {
		r0 = rf(when...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(provideractions.FileTriggerBuilder)
		}
	}

	return r0
}