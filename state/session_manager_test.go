package state

import (
	"context"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestInMemorySessionManager_AddSession(t *testing.T) {
	sm := NewInMemorySessionManager(slog.Default())

	want1 := sm.AddSession("user-screen-name")
	have1 := sm.RetrieveByScreenName(NewIdentScreenName("user-screen-name"))
	assert.Same(t, want1, have1)

	want2 := sm.AddSession("user-screen-name")
	have2 := sm.RetrieveByScreenName(NewIdentScreenName("user-screen-name"))
	assert.Same(t, want2, have2)

	// ensure that the second session created with the same screen name as the
	// first session clobbers the previous session in the session manager store
	assert.NotSame(t, have1, have2)
}

func TestInMemorySessionManager_Remove(t *testing.T) {
	tests := []struct {
		name   string
		given  []DisplayScreenName
		remove IdentScreenName
		want   []IdentScreenName
	}{
		{
			name: "remove user that exists",
			given: []DisplayScreenName{
				"user-screen-name-1",
				"user-screen-name-2",
			},
			remove: NewIdentScreenName("user-screen-name-1"),
			want: []IdentScreenName{
				NewIdentScreenName("user-screen-name-2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewInMemorySessionManager(slog.Default())

			for _, screenName := range tt.given {
				sm.AddSession(screenName)
			}

			sm.RemoveSession(sm.RetrieveByScreenName(tt.remove))

			for i, sess := range sm.AllSessions() {
				assert.Equal(t, tt.want[i], sess.identScreenName)
			}
		})
	}
}

func TestInMemorySessionManager_Empty(t *testing.T) {
	tests := []struct {
		name  string
		given []DisplayScreenName
		want  bool
	}{
		{
			name: "session manager is not empty",
			given: []DisplayScreenName{
				"user-screen-name-1",
			},
			want: false,
		},
		{
			name:  "session manager is empty",
			given: []DisplayScreenName{},
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewInMemorySessionManager(slog.Default())

			for _, screenName := range tt.given {
				sm.AddSession(screenName)
			}

			have := sm.Empty()
			assert.Equal(t, tt.want, have)
		})
	}
}

func TestInMemorySessionManager_Retrieve(t *testing.T) {
	tests := []struct {
		name             string
		given            []DisplayScreenName
		lookupScreenName IdentScreenName
		wantScreenName   IdentScreenName
	}{
		{
			name: "lookup finds match",
			given: []DisplayScreenName{
				"user-screen-name-1",
				"user-screen-name-2",
			},
			lookupScreenName: NewIdentScreenName("user-screen-name-2"),
			wantScreenName:   NewIdentScreenName("user-screen-name-2"),
		},
		{
			name:             "lookup does not find match",
			given:            []DisplayScreenName{},
			lookupScreenName: NewIdentScreenName("user-screen-name-3"),
			wantScreenName:   NewIdentScreenName(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewInMemorySessionManager(slog.Default())

			for _, screenName := range tt.given {
				sm.AddSession(screenName)
			}

			have := sm.RetrieveSession(tt.lookupScreenName)
			if have == nil {
				assert.Empty(t, tt.wantScreenName)
			} else {
				assert.Equal(t, tt.wantScreenName, have.IdentScreenName())
			}
		})
	}
}

func TestInMemorySessionManager_RelayToScreenNames(t *testing.T) {
	sm := NewInMemorySessionManager(slog.Default())

	user1 := sm.AddSession("user-screen-name-1")
	user2 := sm.AddSession("user-screen-name-2")
	user3 := sm.AddSession("user-screen-name-3")

	want := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	recips := []IdentScreenName{
		NewIdentScreenName("user-screen-name-1"),
		NewIdentScreenName("user-screen-name-2"),
	}
	sm.RelayToScreenNames(context.Background(), recips, want)

	select {
	case have := <-user1.ReceiveMessage():
		assert.Equal(t, want, have)
	}

	select {
	case have := <-user2.ReceiveMessage():
		assert.Equal(t, want, have)
	}

	select {
	case <-user3.ReceiveMessage():
		assert.Fail(t, "user 3 should not receive a message")
	default:
	}
}

func TestInMemorySessionManager_Broadcast(t *testing.T) {
	sm := NewInMemorySessionManager(slog.Default())

	user1 := sm.AddSession("user-screen-name-1")
	user2 := sm.AddSession("user-screen-name-2")

	want := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	sm.RelayToAll(context.Background(), want)

	select {
	case have := <-user1.ReceiveMessage():
		assert.Equal(t, want, have)
	}

	select {
	case have := <-user2.ReceiveMessage():
		assert.Equal(t, want, have)
	}
}

func TestInMemorySessionManager_Broadcast_SkipClosedSession(t *testing.T) {
	sm := NewInMemorySessionManager(slog.Default())

	user1 := sm.AddSession("user-screen-name-1")
	user2 := sm.AddSession("user-screen-name-2")
	user2.Close()

	want := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	sm.RelayToAll(context.Background(), want)

	select {
	case have := <-user1.ReceiveMessage():
		assert.Equal(t, want, have)
	}

	select {
	case <-user2.ReceiveMessage():
		assert.Fail(t, "user 2 should not receive a message")
	default:
	}
}

func TestInMemorySessionManager_RelayToScreenName_SessionExists(t *testing.T) {
	sm := NewInMemorySessionManager(slog.Default())

	user1 := sm.AddSession("user-screen-name-1")
	user2 := sm.AddSession("user-screen-name-2")

	want := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	recip := NewIdentScreenName("user-screen-name-1")
	sm.RelayToScreenName(context.Background(), recip, want)

	select {
	case have := <-user1.ReceiveMessage():
		assert.Equal(t, want, have)
	}

	select {
	case <-user2.ReceiveMessage():
		assert.Fail(t, "user 2 should not receive a message")
	default:
	}
}

func TestInMemorySessionManager_RelayToScreenName_SessionNotExist(t *testing.T) {
	sm := NewInMemorySessionManager(slog.Default())

	user1 := sm.AddSession("user-screen-name-1")

	want := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	recip := NewIdentScreenName("user-screen-name-2")
	sm.RelayToScreenName(context.Background(), recip, want)

	select {
	case <-user1.ReceiveMessage():
		assert.Fail(t, "user 1 should not receive a message")
	default:
	}
}

func TestInMemorySessionManager_RelayToScreenName_SkipFullSession(t *testing.T) {
	sm := NewInMemorySessionManager(slog.Default())

	user1 := sm.AddSession("user-screen-name-1")
	msg := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	wantCount := 0
	for {
		if user1.RelayMessage(msg) == SessQueueFull {
			break
		}
		wantCount++
	}

	recip := NewIdentScreenName("user-screen-name-1")
	sm.RelayToScreenName(context.Background(), recip, msg)

	haveCount := 0
loop:
	for {
		select {
		case <-user1.ReceiveMessage():
			haveCount++
		default:
			break loop
		}
	}

	assert.Equal(t, wantCount, haveCount)
}

func TestInMemoryChatSessionManager_RelayToAllExcept_HappyPath(t *testing.T) {
	sm := NewInMemoryChatSessionManager(slog.Default())

	cookie := "the-cookie"
	user1 := sm.AddSession(cookie, "user-screen-name-1")
	user2 := sm.AddSession(cookie, "user-screen-name-2")
	user3 := sm.AddSession(cookie, "user-screen-name-3")

	want := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	sm.RelayToAllExcept(context.Background(), cookie, user2.IdentScreenName(), want)

	select {
	case have := <-user1.ReceiveMessage():
		assert.Equal(t, want, have)
	}

	select {
	case <-user2.ReceiveMessage():
		assert.Fail(t, "user 2 should not receive a message")
	default:
	}

	select {
	case have := <-user3.ReceiveMessage():
		assert.Equal(t, want, have)
	}
}

func TestInMemoryChatSessionManager_AllSessions_RoomExists(t *testing.T) {
	sm := NewInMemoryChatSessionManager(slog.Default())

	user1 := sm.AddSession("the-cookie", "user-screen-name-1")
	user2 := sm.AddSession("the-cookie", "user-screen-name-2")

	sessions := sm.AllSessions("the-cookie")
	assert.Len(t, sessions, 2)

	lookup := make(map[*Session]bool)
	for _, session := range sessions {
		lookup[session] = true
	}

	assert.True(t, lookup[user1])
	assert.True(t, lookup[user2])
}

func TestInMemoryChatSessionManager_RelayToScreenName_SessionAndChatRoomExist(t *testing.T) {
	sm := NewInMemoryChatSessionManager(slog.Default())

	user1 := sm.AddSession("chat-room-1", "user-screen-name-1")
	user2 := sm.AddSession("chat-room-1", "user-screen-name-2")

	want := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	recip := NewIdentScreenName("user-screen-name-1")
	sm.RelayToScreenName(context.Background(), "chat-room-1", recip, want)

	select {
	case have := <-user1.ReceiveMessage():
		assert.Equal(t, want, have)
	}

	select {
	case <-user2.ReceiveMessage():
		assert.Fail(t, "user 2 should not receive a message")
	default:
	}
}

func TestInMemoryChatSessionManager_RemoveSession(t *testing.T) {
	sm := NewInMemoryChatSessionManager(slog.Default())

	user1 := sm.AddSession("chat-room-1", "user-screen-name-1")
	user2 := sm.AddSession("chat-room-1", "user-screen-name-2")

	assert.Len(t, sm.AllSessions("chat-room-1"), 2)

	sm.RemoveSession(user1)
	sm.RemoveSession(user2)

	assert.Empty(t, sm.AllSessions("chat-room-1"))
}
