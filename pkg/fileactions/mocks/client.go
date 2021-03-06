// Code generated by mockery 2.9.0. DO NOT EDIT.

package mocks

import (
	actions "github.com/hostfactor/api/go/blueprint/actions"
	fileactions "github.com/hostfactor/diazo/pkg/fileactions"

	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Download provides a mock function with given fields: root, dl, opts
func (_m *Client) Download(root string, dl *actions.DownloadFile, opts fileactions.DownloadOpts) error {
	ret := _m.Called(root, dl, opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *actions.DownloadFile, fileactions.DownloadOpts) error); ok {
		r0 = rf(root, dl, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Extract provides a mock function with given fields: file
func (_m *Client) Extract(file *actions.ExtractFiles) error {
	ret := _m.Called(file)

	var r0 error
	if rf, ok := ret.Get(0).(func(*actions.ExtractFiles) error); ok {
		r0 = rf(file)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MoveFile provides a mock function with given fields: a
func (_m *Client) MoveFile(a *actions.MoveFile) error {
	ret := _m.Called(a)

	var r0 error
	if rf, ok := ret.Get(0).(func(*actions.MoveFile) error); ok {
		r0 = rf(a)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rename provides a mock function with given fields: r
func (_m *Client) Rename(r *actions.RenameFiles) error {
	ret := _m.Called(r)

	var r0 error
	if rf, ok := ret.Get(0).(func(*actions.RenameFiles) error); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unzip provides a mock function with given fields: file
func (_m *Client) Unzip(file *actions.UnzipFile) error {
	ret := _m.Called(file)

	var r0 error
	if rf, ok := ret.Get(0).(func(*actions.UnzipFile) error); ok {
		r0 = rf(file)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upload provides a mock function with given fields: root, u, opts
func (_m *Client) Upload(root string, u *actions.UploadFile, opts fileactions.UploadOpts) error {
	ret := _m.Called(root, u, opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *actions.UploadFile, fileactions.UploadOpts) error); ok {
		r0 = rf(root, u, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Zip provides a mock function with given fields: z
func (_m *Client) Zip(z *actions.ZipFile) error {
	ret := _m.Called(z)

	var r0 error
	if rf, ok := ret.Get(0).(func(*actions.ZipFile) error); ok {
		r0 = rf(z)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
