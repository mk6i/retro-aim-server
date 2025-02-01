// Code generated by mockery v2.52.1. DO NOT EDIT.

package toc

import (
	context "context"

	state "github.com/mk6i/retro-aim-server/state"
	mock "github.com/stretchr/testify/mock"

	wire "github.com/mk6i/retro-aim-server/wire"
)

// mockPermitDenyService is an autogenerated mock type for the PermitDenyService type
type mockPermitDenyService struct {
	mock.Mock
}

type mockPermitDenyService_Expecter struct {
	mock *mock.Mock
}

func (_m *mockPermitDenyService) EXPECT() *mockPermitDenyService_Expecter {
	return &mockPermitDenyService_Expecter{mock: &_m.Mock}
}

// AddDenyListEntries provides a mock function with given fields: ctx, sess, body
func (_m *mockPermitDenyService) AddDenyListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries) error {
	ret := _m.Called(ctx, sess, body)

	if len(ret) == 0 {
		panic("no return value specified for AddDenyListEntries")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries) error); ok {
		r0 = rf(ctx, sess, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockPermitDenyService_AddDenyListEntries_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddDenyListEntries'
type mockPermitDenyService_AddDenyListEntries_Call struct {
	*mock.Call
}

// AddDenyListEntries is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - body wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries
func (_e *mockPermitDenyService_Expecter) AddDenyListEntries(ctx interface{}, sess interface{}, body interface{}) *mockPermitDenyService_AddDenyListEntries_Call {
	return &mockPermitDenyService_AddDenyListEntries_Call{Call: _e.mock.On("AddDenyListEntries", ctx, sess, body)}
}

func (_c *mockPermitDenyService_AddDenyListEntries_Call) Run(run func(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries)) *mockPermitDenyService_AddDenyListEntries_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries))
	})
	return _c
}

func (_c *mockPermitDenyService_AddDenyListEntries_Call) Return(_a0 error) *mockPermitDenyService_AddDenyListEntries_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockPermitDenyService_AddDenyListEntries_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries) error) *mockPermitDenyService_AddDenyListEntries_Call {
	_c.Call.Return(run)
	return _c
}

// AddPermListEntries provides a mock function with given fields: ctx, sess, body
func (_m *mockPermitDenyService) AddPermListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries) error {
	ret := _m.Called(ctx, sess, body)

	if len(ret) == 0 {
		panic("no return value specified for AddPermListEntries")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries) error); ok {
		r0 = rf(ctx, sess, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockPermitDenyService_AddPermListEntries_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddPermListEntries'
type mockPermitDenyService_AddPermListEntries_Call struct {
	*mock.Call
}

// AddPermListEntries is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - body wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries
func (_e *mockPermitDenyService_Expecter) AddPermListEntries(ctx interface{}, sess interface{}, body interface{}) *mockPermitDenyService_AddPermListEntries_Call {
	return &mockPermitDenyService_AddPermListEntries_Call{Call: _e.mock.On("AddPermListEntries", ctx, sess, body)}
}

func (_c *mockPermitDenyService_AddPermListEntries_Call) Run(run func(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries)) *mockPermitDenyService_AddPermListEntries_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries))
	})
	return _c
}

func (_c *mockPermitDenyService_AddPermListEntries_Call) Return(_a0 error) *mockPermitDenyService_AddPermListEntries_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockPermitDenyService_AddPermListEntries_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries) error) *mockPermitDenyService_AddPermListEntries_Call {
	_c.Call.Return(run)
	return _c
}

// DelDenyListEntries provides a mock function with given fields: ctx, sess, body
func (_m *mockPermitDenyService) DelDenyListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries) error {
	ret := _m.Called(ctx, sess, body)

	if len(ret) == 0 {
		panic("no return value specified for DelDenyListEntries")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries) error); ok {
		r0 = rf(ctx, sess, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockPermitDenyService_DelDenyListEntries_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DelDenyListEntries'
type mockPermitDenyService_DelDenyListEntries_Call struct {
	*mock.Call
}

// DelDenyListEntries is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - body wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries
func (_e *mockPermitDenyService_Expecter) DelDenyListEntries(ctx interface{}, sess interface{}, body interface{}) *mockPermitDenyService_DelDenyListEntries_Call {
	return &mockPermitDenyService_DelDenyListEntries_Call{Call: _e.mock.On("DelDenyListEntries", ctx, sess, body)}
}

func (_c *mockPermitDenyService_DelDenyListEntries_Call) Run(run func(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries)) *mockPermitDenyService_DelDenyListEntries_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries))
	})
	return _c
}

func (_c *mockPermitDenyService_DelDenyListEntries_Call) Return(_a0 error) *mockPermitDenyService_DelDenyListEntries_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockPermitDenyService_DelDenyListEntries_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries) error) *mockPermitDenyService_DelDenyListEntries_Call {
	_c.Call.Return(run)
	return _c
}

// DelPermListEntries provides a mock function with given fields: ctx, sess, body
func (_m *mockPermitDenyService) DelPermListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries) error {
	ret := _m.Called(ctx, sess, body)

	if len(ret) == 0 {
		panic("no return value specified for DelPermListEntries")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries) error); ok {
		r0 = rf(ctx, sess, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockPermitDenyService_DelPermListEntries_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DelPermListEntries'
type mockPermitDenyService_DelPermListEntries_Call struct {
	*mock.Call
}

// DelPermListEntries is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - body wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries
func (_e *mockPermitDenyService_Expecter) DelPermListEntries(ctx interface{}, sess interface{}, body interface{}) *mockPermitDenyService_DelPermListEntries_Call {
	return &mockPermitDenyService_DelPermListEntries_Call{Call: _e.mock.On("DelPermListEntries", ctx, sess, body)}
}

func (_c *mockPermitDenyService_DelPermListEntries_Call) Run(run func(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries)) *mockPermitDenyService_DelPermListEntries_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries))
	})
	return _c
}

func (_c *mockPermitDenyService_DelPermListEntries_Call) Return(_a0 error) *mockPermitDenyService_DelPermListEntries_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockPermitDenyService_DelPermListEntries_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries) error) *mockPermitDenyService_DelPermListEntries_Call {
	_c.Call.Return(run)
	return _c
}

// RightsQuery provides a mock function with given fields: _a0, frame
func (_m *mockPermitDenyService) RightsQuery(_a0 context.Context, frame wire.SNACFrame) wire.SNACMessage {
	ret := _m.Called(_a0, frame)

	if len(ret) == 0 {
		panic("no return value specified for RightsQuery")
	}

	var r0 wire.SNACMessage
	if rf, ok := ret.Get(0).(func(context.Context, wire.SNACFrame) wire.SNACMessage); ok {
		r0 = rf(_a0, frame)
	} else {
		r0 = ret.Get(0).(wire.SNACMessage)
	}

	return r0
}

// mockPermitDenyService_RightsQuery_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RightsQuery'
type mockPermitDenyService_RightsQuery_Call struct {
	*mock.Call
}

// RightsQuery is a helper method to define mock.On call
//   - _a0 context.Context
//   - frame wire.SNACFrame
func (_e *mockPermitDenyService_Expecter) RightsQuery(_a0 interface{}, frame interface{}) *mockPermitDenyService_RightsQuery_Call {
	return &mockPermitDenyService_RightsQuery_Call{Call: _e.mock.On("RightsQuery", _a0, frame)}
}

func (_c *mockPermitDenyService_RightsQuery_Call) Run(run func(_a0 context.Context, frame wire.SNACFrame)) *mockPermitDenyService_RightsQuery_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(wire.SNACFrame))
	})
	return _c
}

func (_c *mockPermitDenyService_RightsQuery_Call) Return(_a0 wire.SNACMessage) *mockPermitDenyService_RightsQuery_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockPermitDenyService_RightsQuery_Call) RunAndReturn(run func(context.Context, wire.SNACFrame) wire.SNACMessage) *mockPermitDenyService_RightsQuery_Call {
	_c.Call.Return(run)
	return _c
}

// newMockPermitDenyService creates a new instance of mockPermitDenyService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockPermitDenyService(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockPermitDenyService {
	mock := &mockPermitDenyService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
