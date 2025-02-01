// Code generated by mockery v2.52.1. DO NOT EDIT.

package http

import (
	mail "net/mail"

	state "github.com/mk6i/retro-aim-server/state"
	mock "github.com/stretchr/testify/mock"
)

// mockAccountRetriever is an autogenerated mock type for the AccountRetriever type
type mockAccountRetriever struct {
	mock.Mock
}

type mockAccountRetriever_Expecter struct {
	mock *mock.Mock
}

func (_m *mockAccountRetriever) EXPECT() *mockAccountRetriever_Expecter {
	return &mockAccountRetriever_Expecter{mock: &_m.Mock}
}

// ConfirmStatusByName provides a mock function with given fields: screnName
func (_m *mockAccountRetriever) ConfirmStatusByName(screnName state.IdentScreenName) (bool, error) {
	ret := _m.Called(screnName)

	if len(ret) == 0 {
		panic("no return value specified for ConfirmStatusByName")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) (bool, error)); ok {
		return rf(screnName)
	}
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) bool); ok {
		r0 = rf(screnName)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(state.IdentScreenName) error); ok {
		r1 = rf(screnName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockAccountRetriever_ConfirmStatusByName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConfirmStatusByName'
type mockAccountRetriever_ConfirmStatusByName_Call struct {
	*mock.Call
}

// ConfirmStatusByName is a helper method to define mock.On call
//   - screnName state.IdentScreenName
func (_e *mockAccountRetriever_Expecter) ConfirmStatusByName(screnName interface{}) *mockAccountRetriever_ConfirmStatusByName_Call {
	return &mockAccountRetriever_ConfirmStatusByName_Call{Call: _e.mock.On("ConfirmStatusByName", screnName)}
}

func (_c *mockAccountRetriever_ConfirmStatusByName_Call) Run(run func(screnName state.IdentScreenName)) *mockAccountRetriever_ConfirmStatusByName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName))
	})
	return _c
}

func (_c *mockAccountRetriever_ConfirmStatusByName_Call) Return(_a0 bool, _a1 error) *mockAccountRetriever_ConfirmStatusByName_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockAccountRetriever_ConfirmStatusByName_Call) RunAndReturn(run func(state.IdentScreenName) (bool, error)) *mockAccountRetriever_ConfirmStatusByName_Call {
	_c.Call.Return(run)
	return _c
}

// EmailAddressByName provides a mock function with given fields: screenName
func (_m *mockAccountRetriever) EmailAddressByName(screenName state.IdentScreenName) (*mail.Address, error) {
	ret := _m.Called(screenName)

	if len(ret) == 0 {
		panic("no return value specified for EmailAddressByName")
	}

	var r0 *mail.Address
	var r1 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) (*mail.Address, error)); ok {
		return rf(screenName)
	}
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) *mail.Address); ok {
		r0 = rf(screenName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mail.Address)
		}
	}

	if rf, ok := ret.Get(1).(func(state.IdentScreenName) error); ok {
		r1 = rf(screenName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockAccountRetriever_EmailAddressByName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EmailAddressByName'
type mockAccountRetriever_EmailAddressByName_Call struct {
	*mock.Call
}

// EmailAddressByName is a helper method to define mock.On call
//   - screenName state.IdentScreenName
func (_e *mockAccountRetriever_Expecter) EmailAddressByName(screenName interface{}) *mockAccountRetriever_EmailAddressByName_Call {
	return &mockAccountRetriever_EmailAddressByName_Call{Call: _e.mock.On("EmailAddressByName", screenName)}
}

func (_c *mockAccountRetriever_EmailAddressByName_Call) Run(run func(screenName state.IdentScreenName)) *mockAccountRetriever_EmailAddressByName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName))
	})
	return _c
}

func (_c *mockAccountRetriever_EmailAddressByName_Call) Return(_a0 *mail.Address, _a1 error) *mockAccountRetriever_EmailAddressByName_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockAccountRetriever_EmailAddressByName_Call) RunAndReturn(run func(state.IdentScreenName) (*mail.Address, error)) *mockAccountRetriever_EmailAddressByName_Call {
	_c.Call.Return(run)
	return _c
}

// RegStatusByName provides a mock function with given fields: screenName
func (_m *mockAccountRetriever) RegStatusByName(screenName state.IdentScreenName) (uint16, error) {
	ret := _m.Called(screenName)

	if len(ret) == 0 {
		panic("no return value specified for RegStatusByName")
	}

	var r0 uint16
	var r1 error
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) (uint16, error)); ok {
		return rf(screenName)
	}
	if rf, ok := ret.Get(0).(func(state.IdentScreenName) uint16); ok {
		r0 = rf(screenName)
	} else {
		r0 = ret.Get(0).(uint16)
	}

	if rf, ok := ret.Get(1).(func(state.IdentScreenName) error); ok {
		r1 = rf(screenName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockAccountRetriever_RegStatusByName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RegStatusByName'
type mockAccountRetriever_RegStatusByName_Call struct {
	*mock.Call
}

// RegStatusByName is a helper method to define mock.On call
//   - screenName state.IdentScreenName
func (_e *mockAccountRetriever_Expecter) RegStatusByName(screenName interface{}) *mockAccountRetriever_RegStatusByName_Call {
	return &mockAccountRetriever_RegStatusByName_Call{Call: _e.mock.On("RegStatusByName", screenName)}
}

func (_c *mockAccountRetriever_RegStatusByName_Call) Run(run func(screenName state.IdentScreenName)) *mockAccountRetriever_RegStatusByName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(state.IdentScreenName))
	})
	return _c
}

func (_c *mockAccountRetriever_RegStatusByName_Call) Return(_a0 uint16, _a1 error) *mockAccountRetriever_RegStatusByName_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockAccountRetriever_RegStatusByName_Call) RunAndReturn(run func(state.IdentScreenName) (uint16, error)) *mockAccountRetriever_RegStatusByName_Call {
	_c.Call.Return(run)
	return _c
}

// newMockAccountRetriever creates a new instance of mockAccountRetriever. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockAccountRetriever(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockAccountRetriever {
	mock := &mockAccountRetriever{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
