// Code generated by mockery 2.9.0. DO NOT EDIT.

package mocks

import (
	blueprint "github.com/hostfactor/api/go/blueprint"
	mock "github.com/stretchr/testify/mock"

	provideractions "github.com/hostfactor/diazo/pkg/provideractions"
)

// FileTriggerBuilder is an autogenerated mock type for the FileTriggerBuilder type
type FileTriggerBuilder struct {
	mock.Mock
}

// Then provides a mock function with given fields: then
func (_m *FileTriggerBuilder) Then(then ...provideractions.FileTriggerActionBuilder) provideractions.RunPhaseBuilder {
	_va := make([]interface{}, len(then))
	for _i := range then {
		_va[_i] = then[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 provideractions.RunPhaseBuilder
	if rf, ok := ret.Get(0).(func(...provideractions.FileTriggerActionBuilder) provideractions.RunPhaseBuilder); ok {
		r0 = rf(then...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(provideractions.RunPhaseBuilder)
		}
	}

	return r0
}

// build provides a mock function with given fields:
func (_m *FileTriggerBuilder) build() *blueprint.FileTrigger {
	ret := _m.Called()

	var r0 *blueprint.FileTrigger
	if rf, ok := ret.Get(0).(func() *blueprint.FileTrigger); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blueprint.FileTrigger)
		}
	}

	return r0
}
