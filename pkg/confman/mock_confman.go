// Code generated by mockery v1.0.0. DO NOT EDIT.

package confman

import (
	context "context"

	storage "github.com/micvbang/confman-go/pkg/storage"
	mock "github.com/stretchr/testify/mock"
)

// MockConfman is an autogenerated mock type for the Confman type
type MockConfman struct {
	mock.Mock
}

// Copy provides a mock function with given fields: ctx, confman
func (_m *MockConfman) Copy(ctx context.Context, confman Confman) error {
	ret := _m.Called(ctx, confman)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Confman) error); ok {
		r0 = rf(ctx, confman)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Define provides a mock function with given fields: ctx, config
func (_m *MockConfman) Define(ctx context.Context, config map[string]string) error {
	ret := _m.Called(ctx, config)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, map[string]string) error); ok {
		r0 = rf(ctx, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, key
func (_m *MockConfman) Delete(ctx context.Context, key string) error {
	ret := _m.Called(ctx, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAll provides a mock function with given fields: ctx
func (_m *MockConfman) DeleteAll(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteKeys provides a mock function with given fields: ctx, keys
func (_m *MockConfman) DeleteKeys(ctx context.Context, keys []string) error {
	ret := _m.Called(ctx, keys)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) error); ok {
		r0 = rf(ctx, keys)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FormatKeyPath provides a mock function with given fields: key
func (_m *MockConfman) FormatKeyPath(key string) string {
	ret := _m.Called(key)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MetadataKeys provides a mock function with given fields:
func (_m *MockConfman) MetadataKeys() []string {
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

// Move provides a mock function with given fields: ctx, confman
func (_m *MockConfman) Move(ctx context.Context, confman Confman) error {
	ret := _m.Called(ctx, confman)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Confman) error); ok {
		r0 = rf(ctx, confman)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Read provides a mock function with given fields: ctx, key
func (_m *MockConfman) Read(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadAll provides a mock function with given fields: ctx
func (_m *MockConfman) ReadAll(ctx context.Context) (map[string]string, error) {
	ret := _m.Called(ctx)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(context.Context) map[string]string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadAllMetadata provides a mock function with given fields: ctx
func (_m *MockConfman) ReadAllMetadata(ctx context.Context) ([]storage.KeyMetadata, error) {
	ret := _m.Called(ctx)

	var r0 []storage.KeyMetadata
	if rf, ok := ret.Get(0).(func(context.Context) []storage.KeyMetadata); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]storage.KeyMetadata)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadKeys provides a mock function with given fields: ctx, keys
func (_m *MockConfman) ReadKeys(ctx context.Context, keys []string) (map[string]string, error) {
	ret := _m.Called(ctx, keys)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(context.Context, []string) map[string]string); ok {
		r0 = rf(ctx, keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, keys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ServicePath provides a mock function with given fields:
func (_m *MockConfman) ServicePath() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// String provides a mock function with given fields:
func (_m *MockConfman) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Write provides a mock function with given fields: ctx, key, value
func (_m *MockConfman) Write(ctx context.Context, key string, value string) error {
	ret := _m.Called(ctx, key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WriteKeys provides a mock function with given fields: ctx, config
func (_m *MockConfman) WriteKeys(ctx context.Context, config map[string]string) error {
	ret := _m.Called(ctx, config)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, map[string]string) error); ok {
		r0 = rf(ctx, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
