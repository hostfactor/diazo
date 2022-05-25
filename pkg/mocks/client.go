// Code generated by mockery 2.9.0. DO NOT EDIT.

package mocks

import (
	fs "io/fs"

	providerconfig "github.com/hostfactor/diazo/pkg/providerconfig"
	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Load provides a mock function with given fields: f, providerFilename, settingsFilename
func (_m *Client) Load(f fs.FS, providerFilename string, settingsFilename string) (*providerconfig.LoadedProviderConfig, error) {
	ret := _m.Called(f, providerFilename, settingsFilename)

	var r0 *providerconfig.LoadedProviderConfig
	if rf, ok := ret.Get(0).(func(fs.FS, string, string) *providerconfig.LoadedProviderConfig); ok {
		r0 = rf(f, providerFilename, settingsFilename)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*providerconfig.LoadedProviderConfig)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(fs.FS, string, string) error); ok {
		r1 = rf(f, providerFilename, settingsFilename)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadAll provides a mock function with given fields: f
func (_m *Client) LoadAll(f fs.FS) ([]*providerconfig.LoadedProviderConfig, error) {
	ret := _m.Called(f)

	var r0 []*providerconfig.LoadedProviderConfig
	if rf, ok := ret.Get(0).(func(fs.FS) []*providerconfig.LoadedProviderConfig); ok {
		r0 = rf(f)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*providerconfig.LoadedProviderConfig)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(fs.FS) error); ok {
		r1 = rf(f)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
