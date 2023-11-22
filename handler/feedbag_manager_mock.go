// Code generated by mockery v2.34.2. DO NOT EDIT.

package handler

import (
	oscar "github.com/mkaminski/goaim/oscar"
	mock "github.com/stretchr/testify/mock"

	state "github.com/mkaminski/goaim/state"

	time "time"
)

// mockFeedbagManager is an autogenerated mock type for the FeedbagManager type
type mockFeedbagManager struct {
	mock.Mock
}

type mockFeedbagManager_Expecter struct {
	mock *mock.Mock
}

func (_m *mockFeedbagManager) EXPECT() *mockFeedbagManager_Expecter {
	return &mockFeedbagManager_Expecter{mock: &_m.Mock}
}

// Blocked provides a mock function with given fields: sn1, sn2
func (_m *mockFeedbagManager) Blocked(sn1 string, sn2 string) (state.BlockedState, error) {
	ret := _m.Called(sn1, sn2)

	var r0 state.BlockedState
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (state.BlockedState, error)); ok {
		return rf(sn1, sn2)
	}
	if rf, ok := ret.Get(0).(func(string, string) state.BlockedState); ok {
		r0 = rf(sn1, sn2)
	} else {
		r0 = ret.Get(0).(state.BlockedState)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(sn1, sn2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockFeedbagManager_Blocked_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Blocked'
type mockFeedbagManager_Blocked_Call struct {
	*mock.Call
}

// Blocked is a helper method to define mock.On call
//   - sn1 string
//   - sn2 string
func (_e *mockFeedbagManager_Expecter) Blocked(sn1 interface{}, sn2 interface{}) *mockFeedbagManager_Blocked_Call {
	return &mockFeedbagManager_Blocked_Call{Call: _e.mock.On("Blocked", sn1, sn2)}
}

func (_c *mockFeedbagManager_Blocked_Call) Run(run func(sn1 string, sn2 string)) *mockFeedbagManager_Blocked_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *mockFeedbagManager_Blocked_Call) Return(_a0 state.BlockedState, _a1 error) *mockFeedbagManager_Blocked_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockFeedbagManager_Blocked_Call) RunAndReturn(run func(string, string) (state.BlockedState, error)) *mockFeedbagManager_Blocked_Call {
	_c.Call.Return(run)
	return _c
}

// Buddies provides a mock function with given fields: screenName
func (_m *mockFeedbagManager) Buddies(screenName string) ([]string, error) {
	ret := _m.Called(screenName)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(screenName)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(screenName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(screenName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockFeedbagManager_Buddies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Buddies'
type mockFeedbagManager_Buddies_Call struct {
	*mock.Call
}

// Buddies is a helper method to define mock.On call
//   - screenName string
func (_e *mockFeedbagManager_Expecter) Buddies(screenName interface{}) *mockFeedbagManager_Buddies_Call {
	return &mockFeedbagManager_Buddies_Call{Call: _e.mock.On("Buddies", screenName)}
}

func (_c *mockFeedbagManager_Buddies_Call) Run(run func(screenName string)) *mockFeedbagManager_Buddies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockFeedbagManager_Buddies_Call) Return(_a0 []string, _a1 error) *mockFeedbagManager_Buddies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockFeedbagManager_Buddies_Call) RunAndReturn(run func(string) ([]string, error)) *mockFeedbagManager_Buddies_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: screenName, items
func (_m *mockFeedbagManager) Delete(screenName string, items []oscar.FeedbagItem) error {
	ret := _m.Called(screenName, items)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []oscar.FeedbagItem) error); ok {
		r0 = rf(screenName, items)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockFeedbagManager_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type mockFeedbagManager_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - screenName string
//   - items []oscar.FeedbagItem
func (_e *mockFeedbagManager_Expecter) Delete(screenName interface{}, items interface{}) *mockFeedbagManager_Delete_Call {
	return &mockFeedbagManager_Delete_Call{Call: _e.mock.On("Delete", screenName, items)}
}

func (_c *mockFeedbagManager_Delete_Call) Run(run func(screenName string, items []oscar.FeedbagItem)) *mockFeedbagManager_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]oscar.FeedbagItem))
	})
	return _c
}

func (_c *mockFeedbagManager_Delete_Call) Return(_a0 error) *mockFeedbagManager_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFeedbagManager_Delete_Call) RunAndReturn(run func(string, []oscar.FeedbagItem) error) *mockFeedbagManager_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// InterestedUsers provides a mock function with given fields: screenName
func (_m *mockFeedbagManager) InterestedUsers(screenName string) ([]string, error) {
	ret := _m.Called(screenName)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(screenName)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(screenName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(screenName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockFeedbagManager_InterestedUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InterestedUsers'
type mockFeedbagManager_InterestedUsers_Call struct {
	*mock.Call
}

// InterestedUsers is a helper method to define mock.On call
//   - screenName string
func (_e *mockFeedbagManager_Expecter) InterestedUsers(screenName interface{}) *mockFeedbagManager_InterestedUsers_Call {
	return &mockFeedbagManager_InterestedUsers_Call{Call: _e.mock.On("InterestedUsers", screenName)}
}

func (_c *mockFeedbagManager_InterestedUsers_Call) Run(run func(screenName string)) *mockFeedbagManager_InterestedUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockFeedbagManager_InterestedUsers_Call) Return(_a0 []string, _a1 error) *mockFeedbagManager_InterestedUsers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockFeedbagManager_InterestedUsers_Call) RunAndReturn(run func(string) ([]string, error)) *mockFeedbagManager_InterestedUsers_Call {
	_c.Call.Return(run)
	return _c
}

// LastModified provides a mock function with given fields: screenName
func (_m *mockFeedbagManager) LastModified(screenName string) (time.Time, error) {
	ret := _m.Called(screenName)

	var r0 time.Time
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (time.Time, error)); ok {
		return rf(screenName)
	}
	if rf, ok := ret.Get(0).(func(string) time.Time); ok {
		r0 = rf(screenName)
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(screenName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockFeedbagManager_LastModified_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LastModified'
type mockFeedbagManager_LastModified_Call struct {
	*mock.Call
}

// LastModified is a helper method to define mock.On call
//   - screenName string
func (_e *mockFeedbagManager_Expecter) LastModified(screenName interface{}) *mockFeedbagManager_LastModified_Call {
	return &mockFeedbagManager_LastModified_Call{Call: _e.mock.On("LastModified", screenName)}
}

func (_c *mockFeedbagManager_LastModified_Call) Run(run func(screenName string)) *mockFeedbagManager_LastModified_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockFeedbagManager_LastModified_Call) Return(_a0 time.Time, _a1 error) *mockFeedbagManager_LastModified_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockFeedbagManager_LastModified_Call) RunAndReturn(run func(string) (time.Time, error)) *mockFeedbagManager_LastModified_Call {
	_c.Call.Return(run)
	return _c
}

// Retrieve provides a mock function with given fields: screenName
func (_m *mockFeedbagManager) Retrieve(screenName string) ([]oscar.FeedbagItem, error) {
	ret := _m.Called(screenName)

	var r0 []oscar.FeedbagItem
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]oscar.FeedbagItem, error)); ok {
		return rf(screenName)
	}
	if rf, ok := ret.Get(0).(func(string) []oscar.FeedbagItem); ok {
		r0 = rf(screenName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]oscar.FeedbagItem)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(screenName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockFeedbagManager_Retrieve_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Retrieve'
type mockFeedbagManager_Retrieve_Call struct {
	*mock.Call
}

// Retrieve is a helper method to define mock.On call
//   - screenName string
func (_e *mockFeedbagManager_Expecter) Retrieve(screenName interface{}) *mockFeedbagManager_Retrieve_Call {
	return &mockFeedbagManager_Retrieve_Call{Call: _e.mock.On("Retrieve", screenName)}
}

func (_c *mockFeedbagManager_Retrieve_Call) Run(run func(screenName string)) *mockFeedbagManager_Retrieve_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockFeedbagManager_Retrieve_Call) Return(_a0 []oscar.FeedbagItem, _a1 error) *mockFeedbagManager_Retrieve_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockFeedbagManager_Retrieve_Call) RunAndReturn(run func(string) ([]oscar.FeedbagItem, error)) *mockFeedbagManager_Retrieve_Call {
	_c.Call.Return(run)
	return _c
}

// Upsert provides a mock function with given fields: screenName, items
func (_m *mockFeedbagManager) Upsert(screenName string, items []oscar.FeedbagItem) error {
	ret := _m.Called(screenName, items)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []oscar.FeedbagItem) error); ok {
		r0 = rf(screenName, items)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockFeedbagManager_Upsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Upsert'
type mockFeedbagManager_Upsert_Call struct {
	*mock.Call
}

// Upsert is a helper method to define mock.On call
//   - screenName string
//   - items []oscar.FeedbagItem
func (_e *mockFeedbagManager_Expecter) Upsert(screenName interface{}, items interface{}) *mockFeedbagManager_Upsert_Call {
	return &mockFeedbagManager_Upsert_Call{Call: _e.mock.On("Upsert", screenName, items)}
}

func (_c *mockFeedbagManager_Upsert_Call) Run(run func(screenName string, items []oscar.FeedbagItem)) *mockFeedbagManager_Upsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]oscar.FeedbagItem))
	})
	return _c
}

func (_c *mockFeedbagManager_Upsert_Call) Return(_a0 error) *mockFeedbagManager_Upsert_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFeedbagManager_Upsert_Call) RunAndReturn(run func(string, []oscar.FeedbagItem) error) *mockFeedbagManager_Upsert_Call {
	_c.Call.Return(run)
	return _c
}

// newMockFeedbagManager creates a new instance of mockFeedbagManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockFeedbagManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockFeedbagManager {
	mock := &mockFeedbagManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}