// Code generated by mockery v1.0.0. DO NOT EDIT.

package storage

import context "context"
import mock "github.com/stretchr/testify/mock"

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// Delete provides a mock function with given fields: ctx, servicePath, key
func (_m *MockStorage) Delete(ctx context.Context, servicePath string, key string) error {
	ret := _m.Called(ctx, servicePath, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, servicePath, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteKeys provides a mock function with given fields: ctx, servicePath, keys
func (_m *MockStorage) DeleteKeys(ctx context.Context, servicePath string, keys []string) error {
	ret := _m.Called(ctx, servicePath, keys)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) error); ok {
		r0 = rf(ctx, servicePath, keys)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MetadataKeys provides a mock function with given fields:
func (_m *MockStorage) MetadataKeys() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Read provides a mock function with given fields: ctx, servicePath, key
func (_m *MockStorage) Read(ctx context.Context, servicePath string, key string) (string, error) {
	ret := _m.Called(ctx, servicePath, key)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, servicePath, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, servicePath, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadAll provides a mock function with given fields: ctx, servicePath
func (_m *MockStorage) ReadAll(ctx context.Context, servicePath string) (map[string]string, error) {
	ret := _m.Called(ctx, servicePath)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(context.Context, string) map[string]string); ok {
		r0 = rf(ctx, servicePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, servicePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadAllMetadata provides a mock function with given fields: ctx, servicePath
func (_m *MockStorage) ReadAllMetadata(ctx context.Context, servicePath string) ([]KeyMetadata, error) {
	ret := _m.Called(ctx, servicePath)

	var r0 []KeyMetadata
	if rf, ok := ret.Get(0).(func(context.Context, string) []KeyMetadata); ok {
		r0 = rf(ctx, servicePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]KeyMetadata)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, servicePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadKeys provides a mock function with given fields: ctx, servicePath, keys
func (_m *MockStorage) ReadKeys(ctx context.Context, servicePath string, keys []string) (map[string]string, error) {
	ret := _m.Called(ctx, servicePath, keys)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) map[string]string); ok {
		r0 = rf(ctx, servicePath, keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, servicePath, keys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// String provides a mock function with given fields:
func (_m *MockStorage) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Write provides a mock function with given fields: ctx, servicePath, key, value
func (_m *MockStorage) Write(ctx context.Context, servicePath string, key string, value string) error {
	ret := _m.Called(ctx, servicePath, key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, servicePath, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WriteKeys provides a mock function with given fields: ctx, servicePath, config
func (_m *MockStorage) WriteKeys(ctx context.Context, servicePath string, config map[string]string) error {
	ret := _m.Called(ctx, servicePath, config)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, map[string]string) error); ok {
		r0 = rf(ctx, servicePath, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
