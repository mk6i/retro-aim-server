// Code generated by mockery v2.52.1. DO NOT EDIT.

package foodgroup

import (
	state "github.com/mk6i/retro-aim-server/state"
	mock "github.com/stretchr/testify/mock"

	time "time"

	wire "github.com/mk6i/retro-aim-server/wire"
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

// Feedbag provides a mock function with given fields: screenName
func (_m *mockFeedbagManager) Feedbag(screenName state.IdentScreenName) ([]wire.FeedbagItem, error) {
	ret := _m.Called(screenName)

	if len(ret) == 0 {
		panic("no return value specified for Feedbag")
	}

	var r0 []wire.FeedbagItem
	var r1 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) ([]wire.FeedbagItem, error)); ok {
		return rf(screenName)
	}
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) []wire.FeedbagItem); ok {
		r0 = rf(screenName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]wire.FeedbagItem)
		}
	}

	if rf, ok := ret.Get(1).(func(state.IdentScreenName) error); ok {
		r1 = rf(screenName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockFeedbagManager_Feedbag_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Feedbag'
type mockFeedbagManager_Feedbag_Call struct {
	*mock.Call
}

// Feedbag is a helper method to define mock.On call
//   - screenName state.IdentScreenName
func (_e *mockFeedbagManager_Expecter) Feedbag(screenName interface{}) *mockFeedbagManager_Feedbag_Call {
	return &mockFeedbagManager_Feedbag_Call{Call: _e.mock.On("Feedbag", screenName)}
}

func (_c *mockFeedbagManager_Feedbag_Call) Run(run func(screenName state.IdentScreenName)) *mockFeedbagManager_Feedbag_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName))
	})
	return _c
}

func (_c *mockFeedbagManager_Feedbag_Call) Return(_a0 []wire.FeedbagItem, _a1 error) *mockFeedbagManager_Feedbag_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockFeedbagManager_Feedbag_Call) RunAndReturn(run func(state.IdentScreenName) ([]wire.FeedbagItem, error)) *mockFeedbagManager_Feedbag_Call {
	_c.Call.Return(run)
	return _c
}

// FeedbagDelete provides a mock function with given fields: screenName, items
func (_m *mockFeedbagManager) FeedbagDelete(screenName state.IdentScreenName, items []wire.FeedbagItem) error {
	ret := _m.Called(screenName, items)

	if len(ret) == 0 {
		panic("no return value specified for FeedbagDelete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName, []wire.FeedbagItem) error); ok {
		r0 = rf(screenName, items)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockFeedbagManager_FeedbagDelete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FeedbagDelete'
type mockFeedbagManager_FeedbagDelete_Call struct {
	*mock.Call
}

// FeedbagDelete is a helper method to define mock.On call
//   - screenName state.IdentScreenName
//   - items []wire.FeedbagItem
func (_e *mockFeedbagManager_Expecter) FeedbagDelete(screenName interface{}, items interface{}) *mockFeedbagManager_FeedbagDelete_Call {
	return &mockFeedbagManager_FeedbagDelete_Call{Call: _e.mock.On("FeedbagDelete", screenName, items)}
}

func (_c *mockFeedbagManager_FeedbagDelete_Call) Run(run func(screenName state.IdentScreenName, items []wire.FeedbagItem)) *mockFeedbagManager_FeedbagDelete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName), args[1].([]wire.FeedbagItem))
	})
	return _c
}

func (_c *mockFeedbagManager_FeedbagDelete_Call) Return(_a0 error) *mockFeedbagManager_FeedbagDelete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFeedbagManager_FeedbagDelete_Call) RunAndReturn(run func(state.IdentScreenName, []wire.FeedbagItem) error) *mockFeedbagManager_FeedbagDelete_Call {
	_c.Call.Return(run)
	return _c
}

// FeedbagLastModified provides a mock function with given fields: screenName
func (_m *mockFeedbagManager) FeedbagLastModified(screenName state.IdentScreenName) (time.Time, error) {
	ret := _m.Called(screenName)

	if len(ret) == 0 {
		panic("no return value specified for FeedbagLastModified")
	}

	var r0 time.Time
	var r1 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) (time.Time, error)); ok {
		return rf(screenName)
	}
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) time.Time); ok {
		r0 = rf(screenName)
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	if rf, ok := ret.Get(1).(func(state.IdentScreenName) error); ok {
		r1 = rf(screenName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockFeedbagManager_FeedbagLastModified_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FeedbagLastModified'
type mockFeedbagManager_FeedbagLastModified_Call struct {
	*mock.Call
}

// FeedbagLastModified is a helper method to define mock.On call
//   - screenName state.IdentScreenName
func (_e *mockFeedbagManager_Expecter) FeedbagLastModified(screenName interface{}) *mockFeedbagManager_FeedbagLastModified_Call {
	return &mockFeedbagManager_FeedbagLastModified_Call{Call: _e.mock.On("FeedbagLastModified", screenName)}
}

func (_c *mockFeedbagManager_FeedbagLastModified_Call) Run(run func(screenName state.IdentScreenName)) *mockFeedbagManager_FeedbagLastModified_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName))
	})
	return _c
}

func (_c *mockFeedbagManager_FeedbagLastModified_Call) Return(_a0 time.Time, _a1 error) *mockFeedbagManager_FeedbagLastModified_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockFeedbagManager_FeedbagLastModified_Call) RunAndReturn(run func(state.IdentScreenName) (time.Time, error)) *mockFeedbagManager_FeedbagLastModified_Call {
	_c.Call.Return(run)
	return _c
}

// FeedbagUpsert provides a mock function with given fields: screenName, items
func (_m *mockFeedbagManager) FeedbagUpsert(screenName state.IdentScreenName, items []wire.FeedbagItem) error {
	ret := _m.Called(screenName, items)

	if len(ret) == 0 {
		panic("no return value specified for FeedbagUpsert")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName, []wire.FeedbagItem) error); ok {
		r0 = rf(screenName, items)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockFeedbagManager_FeedbagUpsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FeedbagUpsert'
type mockFeedbagManager_FeedbagUpsert_Call struct {
	*mock.Call
}

// FeedbagUpsert is a helper method to define mock.On call
//   - screenName state.IdentScreenName
//   - items []wire.FeedbagItem
func (_e *mockFeedbagManager_Expecter) FeedbagUpsert(screenName interface{}, items interface{}) *mockFeedbagManager_FeedbagUpsert_Call {
	return &mockFeedbagManager_FeedbagUpsert_Call{Call: _e.mock.On("FeedbagUpsert", screenName, items)}
}

func (_c *mockFeedbagManager_FeedbagUpsert_Call) Run(run func(screenName state.IdentScreenName, items []wire.FeedbagItem)) *mockFeedbagManager_FeedbagUpsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName), args[1].([]wire.FeedbagItem))
	})
	return _c
}

func (_c *mockFeedbagManager_FeedbagUpsert_Call) Return(_a0 error) *mockFeedbagManager_FeedbagUpsert_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFeedbagManager_FeedbagUpsert_Call) RunAndReturn(run func(state.IdentScreenName, []wire.FeedbagItem) error) *mockFeedbagManager_FeedbagUpsert_Call {
	_c.Call.Return(run)
	return _c
}

// UseFeedbag provides a mock function with given fields: user
func (_m *mockFeedbagManager) UseFeedbag(user state.IdentScreenName) error {
	ret := _m.Called(user)

	if len(ret) == 0 {
		panic("no return value specified for UseFeedbag")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) error); ok {
		r0 = rf(user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockFeedbagManager_UseFeedbag_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UseFeedbag'
type mockFeedbagManager_UseFeedbag_Call struct {
	*mock.Call
}

// UseFeedbag is a helper method to define mock.On call
//   - user state.IdentScreenName
func (_e *mockFeedbagManager_Expecter) UseFeedbag(user interface{}) *mockFeedbagManager_UseFeedbag_Call {
	return &mockFeedbagManager_UseFeedbag_Call{Call: _e.mock.On("UseFeedbag", user)}
}

func (_c *mockFeedbagManager_UseFeedbag_Call) Run(run func(user state.IdentScreenName)) *mockFeedbagManager_UseFeedbag_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName))
	})
	return _c
}

func (_c *mockFeedbagManager_UseFeedbag_Call) Return(_a0 error) *mockFeedbagManager_UseFeedbag_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFeedbagManager_UseFeedbag_Call) RunAndReturn(run func(state.IdentScreenName) error) *mockFeedbagManager_UseFeedbag_Call {
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
