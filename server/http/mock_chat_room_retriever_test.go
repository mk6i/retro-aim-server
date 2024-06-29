// Code generated by mockery v2.43.2. DO NOT EDIT.

package http

import (
	state "github.com/mk6i/retro-aim-server/state"
	mock "github.com/stretchr/testify/mock"
)

// mockChatRoomRetriever is an autogenerated mock type for the ChatRoomRetriever type
type mockChatRoomRetriever struct {
	mock.Mock
}

type mockChatRoomRetriever_Expecter struct {
	mock *mock.Mock
}

func (_m *mockChatRoomRetriever) EXPECT() *mockChatRoomRetriever_Expecter {
	return &mockChatRoomRetriever_Expecter{mock: &_m.Mock}
}

// AllChatRooms provides a mock function with given fields: exchange
func (_m *mockChatRoomRetriever) AllChatRooms(exchange uint16) ([]state.ChatRoom, error) {
	ret := _m.Called(exchange)

	if len(ret) == 0 {
		panic("no return value specified for AllChatRooms")
	}

	var r0 []state.ChatRoom
	var r1 error
	if rf, ok := ret.Get(0).(func(uint16) ([]state.ChatRoom, error)); ok {
		return rf(exchange)
	}
	if rf, ok := ret.Get(0).(func(uint16) []state.ChatRoom); ok {
		r0 = rf(exchange)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]state.ChatRoom)
		}
	}

	if rf, ok := ret.Get(1).(func(uint16) error); ok {
		r1 = rf(exchange)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockChatRoomRetriever_AllChatRooms_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AllChatRooms'
type mockChatRoomRetriever_AllChatRooms_Call struct {
	*mock.Call
}

// AllChatRooms is a helper method to define mock.On call
//   - exchange uint16
func (_e *mockChatRoomRetriever_Expecter) AllChatRooms(exchange interface{}) *mockChatRoomRetriever_AllChatRooms_Call {
	return &mockChatRoomRetriever_AllChatRooms_Call{Call: _e.mock.On("AllChatRooms", exchange)}
}

func (_c *mockChatRoomRetriever_AllChatRooms_Call) Run(run func(exchange uint16)) *mockChatRoomRetriever_AllChatRooms_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uint16))
	})
	return _c
}

func (_c *mockChatRoomRetriever_AllChatRooms_Call) Return(_a0 []state.ChatRoom, _a1 error) *mockChatRoomRetriever_AllChatRooms_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockChatRoomRetriever_AllChatRooms_Call) RunAndReturn(run func(uint16) ([]state.ChatRoom, error)) *mockChatRoomRetriever_AllChatRooms_Call {
	_c.Call.Return(run)
	return _c
}

// newMockChatRoomRetriever creates a new instance of mockChatRoomRetriever. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockChatRoomRetriever(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockChatRoomRetriever {
	mock := &mockChatRoomRetriever{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}