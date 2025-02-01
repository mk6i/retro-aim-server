// Code generated by mockery v2.52.1. DO NOT EDIT.

package foodgroup

import (
	state "github.com/mk6i/retro-aim-server/state"
	mock "github.com/stretchr/testify/mock"
)

// mockOfflineMessageManager is an autogenerated mock type for the OfflineMessageManager type
type mockOfflineMessageManager struct {
	mock.Mock
}

type mockOfflineMessageManager_Expecter struct {
	mock *mock.Mock
}

func (_m *mockOfflineMessageManager) EXPECT() *mockOfflineMessageManager_Expecter {
	return &mockOfflineMessageManager_Expecter{mock: &_m.Mock}
}

// DeleteMessages provides a mock function with given fields: recip
func (_m *mockOfflineMessageManager) DeleteMessages(recip state.IdentScreenName) error {
	ret := _m.Called(recip)

	if len(ret) == 0 {
		panic("no return value specified for DeleteMessages")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) error); ok {
		r0 = rf(recip)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockOfflineMessageManager_DeleteMessages_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteMessages'
type mockOfflineMessageManager_DeleteMessages_Call struct {
	*mock.Call
}

// DeleteMessages is a helper method to define mock.On call
//   - recip state.IdentScreenName
func (_e *mockOfflineMessageManager_Expecter) DeleteMessages(recip interface{}) *mockOfflineMessageManager_DeleteMessages_Call {
	return &mockOfflineMessageManager_DeleteMessages_Call{Call: _e.mock.On("DeleteMessages", recip)}
}

func (_c *mockOfflineMessageManager_DeleteMessages_Call) Run(run func(recip state.IdentScreenName)) *mockOfflineMessageManager_DeleteMessages_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName))
	})
	return _c
}

func (_c *mockOfflineMessageManager_DeleteMessages_Call) Return(_a0 error) *mockOfflineMessageManager_DeleteMessages_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockOfflineMessageManager_DeleteMessages_Call) RunAndReturn(run func(state.IdentScreenName) error) *mockOfflineMessageManager_DeleteMessages_Call {
	_c.Call.Return(run)
	return _c
}

// RetrieveMessages provides a mock function with given fields: recip
func (_m *mockOfflineMessageManager) RetrieveMessages(recip state.IdentScreenName) ([]state.OfflineMessage, error) {
	ret := _m.Called(recip)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveMessages")
	}

	var r0 []state.OfflineMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) ([]state.OfflineMessage, error)); ok {
		return rf(recip)
	}
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) []state.OfflineMessage); ok {
		r0 = rf(recip)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]state.OfflineMessage)
		}
	}

	if rf, ok := ret.Get(1).(func(state.IdentScreenName) error); ok {
		r1 = rf(recip)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockOfflineMessageManager_RetrieveMessages_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RetrieveMessages'
type mockOfflineMessageManager_RetrieveMessages_Call struct {
	*mock.Call
}

// RetrieveMessages is a helper method to define mock.On call
//   - recip state.IdentScreenName
func (_e *mockOfflineMessageManager_Expecter) RetrieveMessages(recip interface{}) *mockOfflineMessageManager_RetrieveMessages_Call {
	return &mockOfflineMessageManager_RetrieveMessages_Call{Call: _e.mock.On("RetrieveMessages", recip)}
}

func (_c *mockOfflineMessageManager_RetrieveMessages_Call) Run(run func(recip state.IdentScreenName)) *mockOfflineMessageManager_RetrieveMessages_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName))
	})
	return _c
}

func (_c *mockOfflineMessageManager_RetrieveMessages_Call) Return(_a0 []state.OfflineMessage, _a1 error) *mockOfflineMessageManager_RetrieveMessages_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockOfflineMessageManager_RetrieveMessages_Call) RunAndReturn(run func(state.IdentScreenName) ([]state.OfflineMessage, error)) *mockOfflineMessageManager_RetrieveMessages_Call {
	_c.Call.Return(run)
	return _c
}

// SaveMessage provides a mock function with given fields: offlineMessage
func (_m *mockOfflineMessageManager) SaveMessage(offlineMessage state.OfflineMessage) error {
	ret := _m.Called(offlineMessage)

	if len(ret) == 0 {
		panic("no return value specified for SaveMessage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(state.OfflineMessage) error); ok {
		r0 = rf(offlineMessage)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockOfflineMessageManager_SaveMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SaveMessage'
type mockOfflineMessageManager_SaveMessage_Call struct {
	*mock.Call
}

// SaveMessage is a helper method to define mock.On call
//   - offlineMessage state.OfflineMessage
func (_e *mockOfflineMessageManager_Expecter) SaveMessage(offlineMessage interface{}) *mockOfflineMessageManager_SaveMessage_Call {
	return &mockOfflineMessageManager_SaveMessage_Call{Call: _e.mock.On("SaveMessage", offlineMessage)}
}

func (_c *mockOfflineMessageManager_SaveMessage_Call) Run(run func(offlineMessage state.OfflineMessage)) *mockOfflineMessageManager_SaveMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.OfflineMessage))
	})
	return _c
}

func (_c *mockOfflineMessageManager_SaveMessage_Call) Return(_a0 error) *mockOfflineMessageManager_SaveMessage_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockOfflineMessageManager_SaveMessage_Call) RunAndReturn(run func(state.OfflineMessage) error) *mockOfflineMessageManager_SaveMessage_Call {
	_c.Call.Return(run)
	return _c
}

// newMockOfflineMessageManager creates a new instance of mockOfflineMessageManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockOfflineMessageManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockOfflineMessageManager {
	mock := &mockOfflineMessageManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
