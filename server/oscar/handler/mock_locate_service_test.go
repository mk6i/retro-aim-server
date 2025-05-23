// Code generated by mockery v2.53.3. DO NOT EDIT.

package handler

import (
	context "context"

	state "github.com/mk6i/retro-aim-server/state"
	mock "github.com/stretchr/testify/mock"

	wire "github.com/mk6i/retro-aim-server/wire"
)

// mockLocateService is an autogenerated mock type for the LocateService type
type mockLocateService struct {
	mock.Mock
}

type mockLocateService_Expecter struct {
	mock *mock.Mock
}

func (_m *mockLocateService) EXPECT() *mockLocateService_Expecter {
	return &mockLocateService_Expecter{mock: &_m.Mock}
}

// DirInfo provides a mock function with given fields: ctx, frame, body
func (_m *mockLocateService) DirInfo(ctx context.Context, frame wire.SNACFrame, body wire.SNAC_0x02_0x0B_LocateGetDirInfo) (wire.SNACMessage, error) {
	ret := _m.Called(ctx, frame, body)

	if len(ret) == 0 {
		panic("no return value specified for DirInfo")
	}

	var r0 wire.SNACMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, wire.SNACFrame, wire.SNAC_0x02_0x0B_LocateGetDirInfo) (wire.SNACMessage, error)); ok {
		return rf(ctx, frame, body)
	}
	if rf, ok := ret.Get(0).(func(context.Context, wire.SNACFrame, wire.SNAC_0x02_0x0B_LocateGetDirInfo) wire.SNACMessage); ok {
		r0 = rf(ctx, frame, body)
	} else {
		r0 = ret.Get(0).(wire.SNACMessage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, wire.SNACFrame, wire.SNAC_0x02_0x0B_LocateGetDirInfo) error); ok {
		r1 = rf(ctx, frame, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockLocateService_DirInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DirInfo'
type mockLocateService_DirInfo_Call struct {
	*mock.Call
}

// DirInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - frame wire.SNACFrame
//   - body wire.SNAC_0x02_0x0B_LocateGetDirInfo
func (_e *mockLocateService_Expecter) DirInfo(ctx interface{}, frame interface{}, body interface{}) *mockLocateService_DirInfo_Call {
	return &mockLocateService_DirInfo_Call{Call: _e.mock.On("DirInfo", ctx, frame, body)}
}

func (_c *mockLocateService_DirInfo_Call) Run(run func(ctx context.Context, frame wire.SNACFrame, body wire.SNAC_0x02_0x0B_LocateGetDirInfo)) *mockLocateService_DirInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(wire.SNACFrame), args[2].(wire.SNAC_0x02_0x0B_LocateGetDirInfo))
	})
	return _c
}

func (_c *mockLocateService_DirInfo_Call) Return(_a0 wire.SNACMessage, _a1 error) *mockLocateService_DirInfo_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockLocateService_DirInfo_Call) RunAndReturn(run func(context.Context, wire.SNACFrame, wire.SNAC_0x02_0x0B_LocateGetDirInfo) (wire.SNACMessage, error)) *mockLocateService_DirInfo_Call {
	_c.Call.Return(run)
	return _c
}

// RightsQuery provides a mock function with given fields: ctx, inFrame
func (_m *mockLocateService) RightsQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
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

// mockLocateService_RightsQuery_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RightsQuery'
type mockLocateService_RightsQuery_Call struct {
	*mock.Call
}

// RightsQuery is a helper method to define mock.On call
//   - ctx context.Context
//   - inFrame wire.SNACFrame
func (_e *mockLocateService_Expecter) RightsQuery(ctx interface{}, inFrame interface{}) *mockLocateService_RightsQuery_Call {
	return &mockLocateService_RightsQuery_Call{Call: _e.mock.On("RightsQuery", ctx, inFrame)}
}

func (_c *mockLocateService_RightsQuery_Call) Run(run func(ctx context.Context, inFrame wire.SNACFrame)) *mockLocateService_RightsQuery_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(wire.SNACFrame))
	})
	return _c
}

func (_c *mockLocateService_RightsQuery_Call) Return(_a0 wire.SNACMessage) *mockLocateService_RightsQuery_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockLocateService_RightsQuery_Call) RunAndReturn(run func(context.Context, wire.SNACFrame) wire.SNACMessage) *mockLocateService_RightsQuery_Call {
	_c.Call.Return(run)
	return _c
}

// SetDirInfo provides a mock function with given fields: ctx, sess, inFrame, inBody
func (_m *mockLocateService) SetDirInfo(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x09_LocateSetDirInfo) (wire.SNACMessage, error) {
	ret := _m.Called(ctx, sess, inFrame, inBody)

	if len(ret) == 0 {
		panic("no return value specified for SetDirInfo")
	}

	var r0 wire.SNACMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x09_LocateSetDirInfo) (wire.SNACMessage, error)); ok {
		return rf(ctx, sess, inFrame, inBody)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x09_LocateSetDirInfo) wire.SNACMessage); ok {
		r0 = rf(ctx, sess, inFrame, inBody)
	} else {
		r0 = ret.Get(0).(wire.SNACMessage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x09_LocateSetDirInfo) error); ok {
		r1 = rf(ctx, sess, inFrame, inBody)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockLocateService_SetDirInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetDirInfo'
type mockLocateService_SetDirInfo_Call struct {
	*mock.Call
}

// SetDirInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - inFrame wire.SNACFrame
//   - inBody wire.SNAC_0x02_0x09_LocateSetDirInfo
func (_e *mockLocateService_Expecter) SetDirInfo(ctx interface{}, sess interface{}, inFrame interface{}, inBody interface{}) *mockLocateService_SetDirInfo_Call {
	return &mockLocateService_SetDirInfo_Call{Call: _e.mock.On("SetDirInfo", ctx, sess, inFrame, inBody)}
}

func (_c *mockLocateService_SetDirInfo_Call) Run(run func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x09_LocateSetDirInfo)) *mockLocateService_SetDirInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNACFrame), args[3].(wire.SNAC_0x02_0x09_LocateSetDirInfo))
	})
	return _c
}

func (_c *mockLocateService_SetDirInfo_Call) Return(_a0 wire.SNACMessage, _a1 error) *mockLocateService_SetDirInfo_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockLocateService_SetDirInfo_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x09_LocateSetDirInfo) (wire.SNACMessage, error)) *mockLocateService_SetDirInfo_Call {
	_c.Call.Return(run)
	return _c
}

// SetInfo provides a mock function with given fields: ctx, sess, inBody
func (_m *mockLocateService) SetInfo(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x02_0x04_LocateSetInfo) error {
	ret := _m.Called(ctx, sess, inBody)

	if len(ret) == 0 {
		panic("no return value specified for SetInfo")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNAC_0x02_0x04_LocateSetInfo) error); ok {
		r0 = rf(ctx, sess, inBody)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockLocateService_SetInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetInfo'
type mockLocateService_SetInfo_Call struct {
	*mock.Call
}

// SetInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - inBody wire.SNAC_0x02_0x04_LocateSetInfo
func (_e *mockLocateService_Expecter) SetInfo(ctx interface{}, sess interface{}, inBody interface{}) *mockLocateService_SetInfo_Call {
	return &mockLocateService_SetInfo_Call{Call: _e.mock.On("SetInfo", ctx, sess, inBody)}
}

func (_c *mockLocateService_SetInfo_Call) Run(run func(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x02_0x04_LocateSetInfo)) *mockLocateService_SetInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNAC_0x02_0x04_LocateSetInfo))
	})
	return _c
}

func (_c *mockLocateService_SetInfo_Call) Return(_a0 error) *mockLocateService_SetInfo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockLocateService_SetInfo_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNAC_0x02_0x04_LocateSetInfo) error) *mockLocateService_SetInfo_Call {
	_c.Call.Return(run)
	return _c
}

// SetKeywordInfo provides a mock function with given fields: ctx, sess, inFrame, body
func (_m *mockLocateService) SetKeywordInfo(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, body wire.SNAC_0x02_0x0F_LocateSetKeywordInfo) (wire.SNACMessage, error) {
	ret := _m.Called(ctx, sess, inFrame, body)

	if len(ret) == 0 {
		panic("no return value specified for SetKeywordInfo")
	}

	var r0 wire.SNACMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x0F_LocateSetKeywordInfo) (wire.SNACMessage, error)); ok {
		return rf(ctx, sess, inFrame, body)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x0F_LocateSetKeywordInfo) wire.SNACMessage); ok {
		r0 = rf(ctx, sess, inFrame, body)
	} else {
		r0 = ret.Get(0).(wire.SNACMessage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x0F_LocateSetKeywordInfo) error); ok {
		r1 = rf(ctx, sess, inFrame, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockLocateService_SetKeywordInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetKeywordInfo'
type mockLocateService_SetKeywordInfo_Call struct {
	*mock.Call
}

// SetKeywordInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - inFrame wire.SNACFrame
//   - body wire.SNAC_0x02_0x0F_LocateSetKeywordInfo
func (_e *mockLocateService_Expecter) SetKeywordInfo(ctx interface{}, sess interface{}, inFrame interface{}, body interface{}) *mockLocateService_SetKeywordInfo_Call {
	return &mockLocateService_SetKeywordInfo_Call{Call: _e.mock.On("SetKeywordInfo", ctx, sess, inFrame, body)}
}

func (_c *mockLocateService_SetKeywordInfo_Call) Run(run func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, body wire.SNAC_0x02_0x0F_LocateSetKeywordInfo)) *mockLocateService_SetKeywordInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNACFrame), args[3].(wire.SNAC_0x02_0x0F_LocateSetKeywordInfo))
	})
	return _c
}

func (_c *mockLocateService_SetKeywordInfo_Call) Return(_a0 wire.SNACMessage, _a1 error) *mockLocateService_SetKeywordInfo_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockLocateService_SetKeywordInfo_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x0F_LocateSetKeywordInfo) (wire.SNACMessage, error)) *mockLocateService_SetKeywordInfo_Call {
	_c.Call.Return(run)
	return _c
}

// UserInfoQuery provides a mock function with given fields: ctx, sess, inFrame, inBody
func (_m *mockLocateService) UserInfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x05_LocateUserInfoQuery) (wire.SNACMessage, error) {
	ret := _m.Called(ctx, sess, inFrame, inBody)

	if len(ret) == 0 {
		panic("no return value specified for UserInfoQuery")
	}

	var r0 wire.SNACMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x05_LocateUserInfoQuery) (wire.SNACMessage, error)); ok {
		return rf(ctx, sess, inFrame, inBody)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x05_LocateUserInfoQuery) wire.SNACMessage); ok {
		r0 = rf(ctx, sess, inFrame, inBody)
	} else {
		r0 = ret.Get(0).(wire.SNACMessage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x05_LocateUserInfoQuery) error); ok {
		r1 = rf(ctx, sess, inFrame, inBody)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockLocateService_UserInfoQuery_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UserInfoQuery'
type mockLocateService_UserInfoQuery_Call struct {
	*mock.Call
}

// UserInfoQuery is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - inFrame wire.SNACFrame
//   - inBody wire.SNAC_0x02_0x05_LocateUserInfoQuery
func (_e *mockLocateService_Expecter) UserInfoQuery(ctx interface{}, sess interface{}, inFrame interface{}, inBody interface{}) *mockLocateService_UserInfoQuery_Call {
	return &mockLocateService_UserInfoQuery_Call{Call: _e.mock.On("UserInfoQuery", ctx, sess, inFrame, inBody)}
}

func (_c *mockLocateService_UserInfoQuery_Call) Run(run func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x05_LocateUserInfoQuery)) *mockLocateService_UserInfoQuery_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.SNACFrame), args[3].(wire.SNAC_0x02_0x05_LocateUserInfoQuery))
	})
	return _c
}

func (_c *mockLocateService_UserInfoQuery_Call) Return(_a0 wire.SNACMessage, _a1 error) *mockLocateService_UserInfoQuery_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockLocateService_UserInfoQuery_Call) RunAndReturn(run func(context.Context, *state.Session, wire.SNACFrame, wire.SNAC_0x02_0x05_LocateUserInfoQuery) (wire.SNACMessage, error)) *mockLocateService_UserInfoQuery_Call {
	_c.Call.Return(run)
	return _c
}

// newMockLocateService creates a new instance of mockLocateService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockLocateService(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockLocateService {
	mock := &mockLocateService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
