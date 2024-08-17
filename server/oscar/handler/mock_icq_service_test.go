// Code generated by mockery v2.43.2. DO NOT EDIT.

package handler

import (
	context "context"

	state "github.com/mk6i/retro-aim-server/state"
	mock "github.com/stretchr/testify/mock"

	wire "github.com/mk6i/retro-aim-server/wire"
)

// mockICQService is an autogenerated mock type for the ICQService type
type mockICQService struct {
	mock.Mock
}

type mockICQService_Expecter struct {
	mock *mock.Mock
}

func (_m *mockICQService) EXPECT() *mockICQService_Expecter {
	return &mockICQService_Expecter{mock: &_m.Mock}
}

// DeleteMsgReq provides a mock function with given fields: ctx, sess, seq
func (_m *mockICQService) DeleteMsgReq(ctx context.Context, sess *state.Session, seq uint16) error {
	ret := _m.Called(ctx, sess, seq)

	if len(ret) == 0 {
		panic("no return value specified for DeleteMsgReq")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, uint16) error); ok {
		r0 = rf(ctx, sess, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_DeleteMsgReq_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteMsgReq'
type mockICQService_DeleteMsgReq_Call struct {
	*mock.Call
}

// DeleteMsgReq is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - seq uint16
func (_e *mockICQService_Expecter) DeleteMsgReq(ctx interface{}, sess interface{}, seq interface{}) *mockICQService_DeleteMsgReq_Call {
	return &mockICQService_DeleteMsgReq_Call{Call: _e.mock.On("DeleteMsgReq", ctx, sess, seq)}
}

func (_c *mockICQService_DeleteMsgReq_Call) Run(run func(ctx context.Context, sess *state.Session, seq uint16)) *mockICQService_DeleteMsgReq_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(uint16))
	})
	return _c
}

func (_c *mockICQService_DeleteMsgReq_Call) Return(_a0 error) *mockICQService_DeleteMsgReq_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_DeleteMsgReq_Call) RunAndReturn(run func(context.Context, *state.Session, uint16) error) *mockICQService_DeleteMsgReq_Call {
	_c.Call.Return(run)
	return _c
}

// FindByDetails provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) FindByDetails(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for FindByDetails")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_FindByDetails_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByDetails'
type mockICQService_FindByDetails_Call struct {
	*mock.Call
}

// FindByDetails is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails
//   - seq uint16
func (_e *mockICQService_Expecter) FindByDetails(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_FindByDetails_Call {
	return &mockICQService_FindByDetails_Call{Call: _e.mock.On("FindByDetails", ctx, sess, req, seq)}
}

func (_c *mockICQService_FindByDetails_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails, seq uint16)) *mockICQService_FindByDetails_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_FindByDetails_Call) Return(_a0 error) *mockICQService_FindByDetails_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_FindByDetails_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails, uint16) error) *mockICQService_FindByDetails_Call {
	_c.Call.Return(run)
	return _c
}

// FindByEmail provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) FindByEmail(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for FindByEmail")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_FindByEmail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByEmail'
type mockICQService_FindByEmail_Call struct {
	*mock.Call
}

// FindByEmail is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail
//   - seq uint16
func (_e *mockICQService_Expecter) FindByEmail(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_FindByEmail_Call {
	return &mockICQService_FindByEmail_Call{Call: _e.mock.On("FindByEmail", ctx, sess, req, seq)}
}

func (_c *mockICQService_FindByEmail_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail, seq uint16)) *mockICQService_FindByEmail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_FindByEmail_Call) Return(_a0 error) *mockICQService_FindByEmail_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_FindByEmail_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail, uint16) error) *mockICQService_FindByEmail_Call {
	_c.Call.Return(run)
	return _c
}

// FindByInterests provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) FindByInterests(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for FindByInterests")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_FindByInterests_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByInterests'
type mockICQService_FindByInterests_Call struct {
	*mock.Call
}

// FindByInterests is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages
//   - seq uint16
func (_e *mockICQService_Expecter) FindByInterests(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_FindByInterests_Call {
	return &mockICQService_FindByInterests_Call{Call: _e.mock.On("FindByInterests", ctx, sess, req, seq)}
}

func (_c *mockICQService_FindByInterests_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages, seq uint16)) *mockICQService_FindByInterests_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_FindByInterests_Call) Return(_a0 error) *mockICQService_FindByInterests_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_FindByInterests_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages, uint16) error) *mockICQService_FindByInterests_Call {
	_c.Call.Return(run)
	return _c
}

// FindByUIN provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) FindByUIN(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for FindByUIN")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_FindByUIN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByUIN'
type mockICQService_FindByUIN_Call struct {
	*mock.Call
}

// FindByUIN is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN
//   - seq uint16
func (_e *mockICQService_Expecter) FindByUIN(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_FindByUIN_Call {
	return &mockICQService_FindByUIN_Call{Call: _e.mock.On("FindByUIN", ctx, sess, req, seq)}
}

func (_c *mockICQService_FindByUIN_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, seq uint16)) *mockICQService_FindByUIN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_FindByUIN_Call) Return(_a0 error) *mockICQService_FindByUIN_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_FindByUIN_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, uint16) error) *mockICQService_FindByUIN_Call {
	_c.Call.Return(run)
	return _c
}

// FullUserInfo provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) FullUserInfo(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for FullUserInfo")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_FullUserInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FullUserInfo'
type mockICQService_FullUserInfo_Call struct {
	*mock.Call
}

// FullUserInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN
//   - seq uint16
func (_e *mockICQService_Expecter) FullUserInfo(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_FullUserInfo_Call {
	return &mockICQService_FullUserInfo_Call{Call: _e.mock.On("FullUserInfo", ctx, sess, req, seq)}
}

func (_c *mockICQService_FullUserInfo_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, seq uint16)) *mockICQService_FullUserInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_FullUserInfo_Call) Return(_a0 error) *mockICQService_FullUserInfo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_FullUserInfo_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, uint16) error) *mockICQService_FullUserInfo_Call {
	_c.Call.Return(run)
	return _c
}

// OfflineMsgReq provides a mock function with given fields: ctx, sess, seq
func (_m *mockICQService) OfflineMsgReq(ctx context.Context, sess *state.Session, seq uint16) error {
	ret := _m.Called(ctx, sess, seq)

	if len(ret) == 0 {
		panic("no return value specified for OfflineMsgReq")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, uint16) error); ok {
		r0 = rf(ctx, sess, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_OfflineMsgReq_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OfflineMsgReq'
type mockICQService_OfflineMsgReq_Call struct {
	*mock.Call
}

// OfflineMsgReq is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - seq uint16
func (_e *mockICQService_Expecter) OfflineMsgReq(ctx interface{}, sess interface{}, seq interface{}) *mockICQService_OfflineMsgReq_Call {
	return &mockICQService_OfflineMsgReq_Call{Call: _e.mock.On("OfflineMsgReq", ctx, sess, seq)}
}

func (_c *mockICQService_OfflineMsgReq_Call) Run(run func(ctx context.Context, sess *state.Session, seq uint16)) *mockICQService_OfflineMsgReq_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(uint16))
	})
	return _c
}

func (_c *mockICQService_OfflineMsgReq_Call) Return(_a0 error) *mockICQService_OfflineMsgReq_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_OfflineMsgReq_Call) RunAndReturn(run func(context.Context, *state.Session, uint16) error) *mockICQService_OfflineMsgReq_Call {
	_c.Call.Return(run)
	return _c
}

// SetAffiliations provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) SetAffiliations(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for SetAffiliations")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_SetAffiliations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetAffiliations'
type mockICQService_SetAffiliations_Call struct {
	*mock.Call
}

// SetAffiliations is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations
//   - seq uint16
func (_e *mockICQService_Expecter) SetAffiliations(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_SetAffiliations_Call {
	return &mockICQService_SetAffiliations_Call{Call: _e.mock.On("SetAffiliations", ctx, sess, req, seq)}
}

func (_c *mockICQService_SetAffiliations_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations, seq uint16)) *mockICQService_SetAffiliations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_SetAffiliations_Call) Return(_a0 error) *mockICQService_SetAffiliations_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_SetAffiliations_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations, uint16) error) *mockICQService_SetAffiliations_Call {
	_c.Call.Return(run)
	return _c
}

// SetBasicInfo provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) SetBasicInfo(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for SetBasicInfo")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_SetBasicInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetBasicInfo'
type mockICQService_SetBasicInfo_Call struct {
	*mock.Call
}

// SetBasicInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo
//   - seq uint16
func (_e *mockICQService_Expecter) SetBasicInfo(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_SetBasicInfo_Call {
	return &mockICQService_SetBasicInfo_Call{Call: _e.mock.On("SetBasicInfo", ctx, sess, req, seq)}
}

func (_c *mockICQService_SetBasicInfo_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo, seq uint16)) *mockICQService_SetBasicInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_SetBasicInfo_Call) Return(_a0 error) *mockICQService_SetBasicInfo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_SetBasicInfo_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo, uint16) error) *mockICQService_SetBasicInfo_Call {
	_c.Call.Return(run)
	return _c
}

// SetEmails provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) SetEmails(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for SetEmails")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_SetEmails_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetEmails'
type mockICQService_SetEmails_Call struct {
	*mock.Call
}

// SetEmails is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails
//   - seq uint16
func (_e *mockICQService_Expecter) SetEmails(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_SetEmails_Call {
	return &mockICQService_SetEmails_Call{Call: _e.mock.On("SetEmails", ctx, sess, req, seq)}
}

func (_c *mockICQService_SetEmails_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails, seq uint16)) *mockICQService_SetEmails_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_SetEmails_Call) Return(_a0 error) *mockICQService_SetEmails_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_SetEmails_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails, uint16) error) *mockICQService_SetEmails_Call {
	_c.Call.Return(run)
	return _c
}

// SetInterests provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) SetInterests(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for SetInterests")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_SetInterests_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetInterests'
type mockICQService_SetInterests_Call struct {
	*mock.Call
}

// SetInterests is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests
//   - seq uint16
func (_e *mockICQService_Expecter) SetInterests(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_SetInterests_Call {
	return &mockICQService_SetInterests_Call{Call: _e.mock.On("SetInterests", ctx, sess, req, seq)}
}

func (_c *mockICQService_SetInterests_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests, seq uint16)) *mockICQService_SetInterests_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_SetInterests_Call) Return(_a0 error) *mockICQService_SetInterests_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_SetInterests_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests, uint16) error) *mockICQService_SetInterests_Call {
	_c.Call.Return(run)
	return _c
}

// SetMoreInfo provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) SetMoreInfo(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for SetMoreInfo")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_SetMoreInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetMoreInfo'
type mockICQService_SetMoreInfo_Call struct {
	*mock.Call
}

// SetMoreInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo
//   - seq uint16
func (_e *mockICQService_Expecter) SetMoreInfo(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_SetMoreInfo_Call {
	return &mockICQService_SetMoreInfo_Call{Call: _e.mock.On("SetMoreInfo", ctx, sess, req, seq)}
}

func (_c *mockICQService_SetMoreInfo_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo, seq uint16)) *mockICQService_SetMoreInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_SetMoreInfo_Call) Return(_a0 error) *mockICQService_SetMoreInfo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_SetMoreInfo_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo, uint16) error) *mockICQService_SetMoreInfo_Call {
	_c.Call.Return(run)
	return _c
}

// SetPermissions provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) SetPermissions(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for SetPermissions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_SetPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetPermissions'
type mockICQService_SetPermissions_Call struct {
	*mock.Call
}

// SetPermissions is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions
//   - seq uint16
func (_e *mockICQService_Expecter) SetPermissions(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_SetPermissions_Call {
	return &mockICQService_SetPermissions_Call{Call: _e.mock.On("SetPermissions", ctx, sess, req, seq)}
}

func (_c *mockICQService_SetPermissions_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions, seq uint16)) *mockICQService_SetPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_SetPermissions_Call) Return(_a0 error) *mockICQService_SetPermissions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_SetPermissions_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions, uint16) error) *mockICQService_SetPermissions_Call {
	_c.Call.Return(run)
	return _c
}

// SetUserNotes provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) SetUserNotes(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for SetUserNotes")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_SetUserNotes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetUserNotes'
type mockICQService_SetUserNotes_Call struct {
	*mock.Call
}

// SetUserNotes is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes
//   - seq uint16
func (_e *mockICQService_Expecter) SetUserNotes(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_SetUserNotes_Call {
	return &mockICQService_SetUserNotes_Call{Call: _e.mock.On("SetUserNotes", ctx, sess, req, seq)}
}

func (_c *mockICQService_SetUserNotes_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes, seq uint16)) *mockICQService_SetUserNotes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_SetUserNotes_Call) Return(_a0 error) *mockICQService_SetUserNotes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_SetUserNotes_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes, uint16) error) *mockICQService_SetUserNotes_Call {
	_c.Call.Return(run)
	return _c
}

// SetWorkInfo provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) SetWorkInfo(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for SetWorkInfo")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_SetWorkInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetWorkInfo'
type mockICQService_SetWorkInfo_Call struct {
	*mock.Call
}

// SetWorkInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo
//   - seq uint16
func (_e *mockICQService_Expecter) SetWorkInfo(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_SetWorkInfo_Call {
	return &mockICQService_SetWorkInfo_Call{Call: _e.mock.On("SetWorkInfo", ctx, sess, req, seq)}
}

func (_c *mockICQService_SetWorkInfo_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo, seq uint16)) *mockICQService_SetWorkInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_SetWorkInfo_Call) Return(_a0 error) *mockICQService_SetWorkInfo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_SetWorkInfo_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo, uint16) error) *mockICQService_SetWorkInfo_Call {
	_c.Call.Return(run)
	return _c
}

// XMLReqData provides a mock function with given fields: ctx, sess, req, seq
func (_m *mockICQService) XMLReqData(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq, seq uint16) error {
	ret := _m.Called(ctx, sess, req, seq)

	if len(ret) == 0 {
		panic("no return value specified for XMLReqData")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq, uint16) error); ok {
		r0 = rf(ctx, sess, req, seq)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockICQService_XMLReqData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'XMLReqData'
type mockICQService_XMLReqData_Call struct {
	*mock.Call
}

// XMLReqData is a helper method to define mock.On call
//   - ctx context.Context
//   - sess *state.Session
//   - req wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq
//   - seq uint16
func (_e *mockICQService_Expecter) XMLReqData(ctx interface{}, sess interface{}, req interface{}, seq interface{}) *mockICQService_XMLReqData_Call {
	return &mockICQService_XMLReqData_Call{Call: _e.mock.On("XMLReqData", ctx, sess, req, seq)}
}

func (_c *mockICQService_XMLReqData_Call) Run(run func(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq, seq uint16)) *mockICQService_XMLReqData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Session), args[2].(wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq), args[3].(uint16))
	})
	return _c
}

func (_c *mockICQService_XMLReqData_Call) Return(_a0 error) *mockICQService_XMLReqData_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockICQService_XMLReqData_Call) RunAndReturn(run func(context.Context, *state.Session, wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq, uint16) error) *mockICQService_XMLReqData_Call {
	_c.Call.Return(run)
	return _c
}

// newMockICQService creates a new instance of mockICQService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockICQService(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockICQService {
	mock := &mockICQService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}