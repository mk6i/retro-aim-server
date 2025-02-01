// Code generated by mockery v2.52.1. DO NOT EDIT.

package toc

import (
	context "context"

	state "github.com/mk6i/retro-aim-server/state"
	mock "github.com/stretchr/testify/mock"

	wire "github.com/mk6i/retro-aim-server/wire"
)

// mockBuddyService is an autogenerated mock type for the BuddyService type
type mockBuddyService struct {
	mock.Mock
}

type mockBuddyService_Expecter struct {
	mock *mock.Mock
}

func (_m *mockBuddyService) EXPECT() *mockBuddyService_Expecter {
	return &mockBuddyService_Expecter{mock: &_m.Mock}
}

// AddBuddies provides a mock function with given fields: ctx, sess, inBody
func (_m *mockBuddyService) AddBuddies(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x04_BuddyAddBuddies) error {
	ret := _m.Called(ctx, sess, inBody)

	if len(ret) == 0 {
		panic("no return value specified for AddBuddies")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNAC_0x03_0x04_BuddyAddBuddies) error); ok {
		r0 = rf(ctx, sess, inBody)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockBuddyService_AddBuddies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddBuddies'
type mockBuddyService_AddBuddies_Call struct {
	*mock.Call
}

// AddBuddies is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - inBody wire.SNAC_0x03_0x04_BuddyAddBuddies
func (_e *mockBuddyService_Expecter) AddBuddies(ctx interface{}, sess interface{}, inBody interface{}) *mockBuddyService_AddBuddies_Call {
	return &mockBuddyService_AddBuddies_Call{Call: _e.mock.On("AddBuddies", ctx, sess, inBody)}
}

func (_c *mockBuddyService_AddBuddies_Call) Run(run func(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x04_BuddyAddBuddies)) *mockBuddyService_AddBuddies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNAC_0x03_0x04_BuddyAddBuddies))
	})
	return _c
}

func (_c *mockBuddyService_AddBuddies_Call) Return(_a0 error) *mockBuddyService_AddBuddies_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockBuddyService_AddBuddies_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNAC_0x03_0x04_BuddyAddBuddies) error) *mockBuddyService_AddBuddies_Call {
	_c.Call.Return(run)
	return _c
}

// BroadcastBuddyDeparted provides a mock function with given fields: ctx, sess
func (_m *mockBuddyService) BroadcastBuddyDeparted(ctx context.Context, sess *state.Session) error {
	ret := _m.Called(ctx, sess)

	if len(ret) == 0 {
		panic("no return value specified for BroadcastBuddyDeparted")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session) error); ok {
		r0 = rf(ctx, sess)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockBuddyService_BroadcastBuddyDeparted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BroadcastBuddyDeparted'
type mockBuddyService_BroadcastBuddyDeparted_Call struct {
	*mock.Call
}

// BroadcastBuddyDeparted is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
func (_e *mockBuddyService_Expecter) BroadcastBuddyDeparted(ctx interface{}, sess interface{}) *mockBuddyService_BroadcastBuddyDeparted_Call {
	return &mockBuddyService_BroadcastBuddyDeparted_Call{Call: _e.mock.On("BroadcastBuddyDeparted", ctx, sess)}
}

func (_c *mockBuddyService_BroadcastBuddyDeparted_Call) Run(run func(ctx context.Context, sess *state.Session)) *mockBuddyService_BroadcastBuddyDeparted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session))
	})
	return _c
}

func (_c *mockBuddyService_BroadcastBuddyDeparted_Call) Return(_a0 error) *mockBuddyService_BroadcastBuddyDeparted_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockBuddyService_BroadcastBuddyDeparted_Call) RunAndReturn(run func(context.Context, *state.Session) error) *mockBuddyService_BroadcastBuddyDeparted_Call {
	_c.Call.Return(run)
	return _c
}

// DelBuddies provides a mock function with given fields: _a0, sess, inBody
func (_m *mockBuddyService) DelBuddies(_a0 context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x05_BuddyDelBuddies) error {
	ret := _m.Called(_a0, sess, inBody)

	if len(ret) == 0 {
		panic("no return value specified for DelBuddies")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNAC_0x03_0x05_BuddyDelBuddies) error); ok {
		r0 = rf(_a0, sess, inBody)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockBuddyService_DelBuddies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DelBuddies'
type mockBuddyService_DelBuddies_Call struct {
	*mock.Call
}

// DelBuddies is a helper method to define mock.On call
//   - _a0 context.Context
//   - sess *state.Session
//   - inBody wire.SNAC_0x03_0x05_BuddyDelBuddies
func (_e *mockBuddyService_Expecter) DelBuddies(_a0 interface{}, sess interface{}, inBody interface{}) *mockBuddyService_DelBuddies_Call {
	return &mockBuddyService_DelBuddies_Call{Call: _e.mock.On("DelBuddies", _a0, sess, inBody)}
}

func (_c *mockBuddyService_DelBuddies_Call) Run(run func(_a0 context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x05_BuddyDelBuddies)) *mockBuddyService_DelBuddies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNAC_0x03_0x05_BuddyDelBuddies))
	})
	return _c
}

func (_c *mockBuddyService_DelBuddies_Call) Return(_a0 error) *mockBuddyService_DelBuddies_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockBuddyService_DelBuddies_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNAC_0x03_0x05_BuddyDelBuddies) error) *mockBuddyService_DelBuddies_Call {
	_c.Call.Return(run)
	return _c
}

// RightsQuery provides a mock function with given fields: ctx, inFrame
func (_m *mockBuddyService) RightsQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	ret := _m.Called(ctx, inFrame)

	if len(ret) == 0 {
		panic("no return value specified for RightsQuery")
	}

	var r0 wire.SNACMessage
	if rf, ok := ret.Get(0).(func(context.Context, wire.SNACFrame) wire.SNACMessage); ok {
		r0 = rf(ctx, inFrame)
	} else {
		r0 = ret.Get(0).(wire.SNACMessage)
	}

	return r0
}

// mockBuddyService_RightsQuery_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RightsQuery'
type mockBuddyService_RightsQuery_Call struct {
	*mock.Call
}

// RightsQuery is a helper method to define mock.On call
//   - ctx context.Context
//   - inFrame wire.SNACFrame
func (_e *mockBuddyService_Expecter) RightsQuery(ctx interface{}, inFrame interface{}) *mockBuddyService_RightsQuery_Call {
	return &mockBuddyService_RightsQuery_Call{Call: _e.mock.On("RightsQuery", ctx, inFrame)}
}

func (_c *mockBuddyService_RightsQuery_Call) Run(run func(ctx context.Context, inFrame wire.SNACFrame)) *mockBuddyService_RightsQuery_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(wire.SNACFrame))
	})
	return _c
}

func (_c *mockBuddyService_RightsQuery_Call) Return(_a0 wire.SNACMessage) *mockBuddyService_RightsQuery_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockBuddyService_RightsQuery_Call) RunAndReturn(run func(context.Context, wire.SNACFrame) wire.SNACMessage) *mockBuddyService_RightsQuery_Call {
	_c.Call.Return(run)
	return _c
}

// newMockBuddyService creates a new instance of mockBuddyService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockBuddyService(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockBuddyService {
	mock := &mockBuddyService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
