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
	have1 := sm.RetrieveByScreenName("user-screen-name")
	assert.Same(t, want1, have1)

	want2 := sm.AddSession("user-screen-name")
	have2 := sm.RetrieveByScreenName("user-screen-name")
	assert.Same(t, want2, have2)

	// ensure that the second session created with the same screen name as the
	// first session clobbers the previous session in the session manager store
	assert.NotSame(t, have1, have2)
}

func TestInMemorySessionManager_Remove(t *testing.T) {
	tests := []struct {
		name   string
		given  []*Session
		remove string
		want   []string
	}{
		{
			name: "remove user that exists",
			given: []*Session{
				{
					screenName: "user-screen-name-1",
				},
				{
					screenName: "user-screen-name-2",
				},
			},
			remove: "user-screen-name-1",
			want: []string{
				"user-screen-name-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewInMemorySessionManager(slog.Default())

			for _, sess := range tt.given {
				sm.AddSession(sess.screenName)
			}

			sm.RemoveSession(sm.RetrieveByScreenName(tt.remove))

			for i, sess := range sm.AllSessions() {
				assert.Equal(t, tt.want[i], sess.screenName)
			}
		})
	}
}

func TestInMemorySessionManager_Empty(t *testing.T) {
	tests := []struct {
		name   string
		given  []*Session
		remove string
		want   bool
	}{
		{
			name: "session manager is not empty",
			given: []*Session{
				{
					screenName: "user-screen-name-1",
				},
			},
			want: false,
		},
		{
			name:  "session manager is empty",
			given: []*Session{},
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewInMemorySessionManager(slog.Default())

			for _, sess := range tt.given {
				sm.AddSession(sess.screenName)
			}

			have := sm.Empty()
			assert.Equal(t, tt.want, have)
		})
	}
}

func TestInMemorySessionManager_Retrieve(t *testing.T) {
	tests := []struct {
		name             string
		given            []*Session
		lookupScreenName string
		remove           string
		wantScreenName   string
	}{
		{
			name: "lookup finds match",
			given: []*Session{
				{
					screenName: "user-screen-name-1",
				},
				{
					screenName: "user-screen-name-2",
				},
			},
			lookupScreenName: "user-screen-name-2",
			wantScreenName:   "user-screen-name-2",
		},
		{
			name:             "lookup does not find match",
			given:            []*Session{},
			lookupScreenName: "user-screen-name-3",
			wantScreenName:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewInMemorySessionManager(slog.Default())

			for _, sess := range tt.given {
				sm.AddSession(sess.screenName)
			}

			have := sm.RetrieveSession(tt.lookupScreenName)
			if have == nil {
				assert.Empty(t, tt.wantScreenName)
			} else {
				assert.Equal(t, tt.wantScreenName, have.ScreenName())
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

	recips := []string{"user-screen-name-1", "user-screen-name-2"}
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

	recip := "user-screen-name-1"
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

	recip := "user-screen-name-2"
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

	recip := "user-screen-name-1"
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

func TestInMemorySessionManager_RelayToAllExcept(t *testing.T) {
	sm := NewInMemorySessionManager(slog.Default())

	user1 := sm.AddSession("user-screen-name-1")
	user2 := sm.AddSession("user-screen-name-2")
	user3 := sm.AddSession("user-screen-name-3")

	want := wire.SNACMessage{Frame: wire.SNACFrame{FoodGroup: wire.ICBM}}

	sm.RelayToAllExcept(context.Background(), user2, want)

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
