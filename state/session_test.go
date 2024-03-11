package state

import (
	"sync"
	"testing"
	"time"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestSession_SetAndGetAwayMessage(t *testing.T) {
	s := NewSession()
	assert.Empty(t, s.AwayMessage())

	msg := "here's my message"
	s.SetAwayMessage(msg)
	assert.Equal(t, msg, s.AwayMessage())
}

func TestSession_SetAndGetID(t *testing.T) {
	s := NewSession()
	// make sure NewSession creates a default ID
	assert.NotEmpty(t, s.SetID)
	newID := "new-id"
	s.SetID(newID)
	assert.Equal(t, newID, s.ID())
}

func TestSession_IncrementAndGetWarning(t *testing.T) {
	s := NewSession()
	assert.Zero(t, s.Warning())
	s.IncrementWarning(1)
	s.IncrementWarning(2)
	assert.Equal(t, uint16(3), s.Warning())
}

func TestSession_SetAndGetInvisible(t *testing.T) {
	s := NewSession()
	assert.False(t, s.Invisible())
	s.SetInvisible(true)
	assert.True(t, s.Invisible())
	s.SetInvisible(false)
	assert.False(t, s.Invisible())
}

func TestSession_SetAndGetScreenName(t *testing.T) {
	s := NewSession()
	assert.Empty(t, s.ScreenName())
	sn := "user-screen-name"
	s.SetScreenName(sn)
	assert.Equal(t, sn, s.ScreenName())
}

func TestSession_SetAndGetChatRoomCookie(t *testing.T) {
	s := NewSession()
	assert.Empty(t, s.ChatRoomCookie())
	sn := "the-chat-cookie"
	s.SetChatRoomCookie(sn)
	assert.Equal(t, sn, s.ChatRoomCookie())
}

func TestSession_TLVUserInfo(t *testing.T) {
	tests := []struct {
		name           string
		givenSessionFn func() *Session
		want           wire.TLVUserInfo
	}{
		{
			name: "user is active and visible",
			givenSessionFn: func() *Session {
				s := NewSession()
				s.SetSignonTime(time.Unix(1, 0))
				s.SetScreenName("xXAIMUSERXx")
				s.IncrementWarning(10)
				return s
			},
			want: wire.TLVUserInfo{
				ScreenName:   "xXAIMUSERXx",
				WarningLevel: 10,
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceUserInfoSignonTOD, uint32(1)),
						wire.NewTLV(wire.OServiceUserInfoUserFlags, uint16(0x0010)),
						wire.NewTLV(wire.OServiceUserInfoStatus, uint16(0x0000)),
						wire.NewTLV(wire.OServiceUserInfoIdleTime, uint16(0)),
					},
				},
			},
		},
		{
			name: "user has away message set",
			givenSessionFn: func() *Session {
				s := NewSession()
				s.SetSignonTime(time.Unix(1, 0))
				s.SetAwayMessage("here's my away message")
				return s
			},
			want: wire.TLVUserInfo{
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceUserInfoSignonTOD, uint32(1)),
						wire.NewTLV(wire.OServiceUserInfoUserFlags, uint16(0x30)),
						wire.NewTLV(wire.OServiceUserInfoStatus, uint16(0x0000)),
						wire.NewTLV(wire.OServiceUserInfoIdleTime, uint16(0)),
					},
				},
			},
		},
		{
			name: "user is invisible",
			givenSessionFn: func() *Session {
				s := NewSession()
				s.SetSignonTime(time.Unix(1, 0))
				s.SetInvisible(true)
				return s
			},
			want: wire.TLVUserInfo{
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceUserInfoSignonTOD, uint32(1)),
						wire.NewTLV(wire.OServiceUserInfoUserFlags, uint16(0x0010)),
						wire.NewTLV(wire.OServiceUserInfoStatus, uint16(0x0100)),
						wire.NewTLV(wire.OServiceUserInfoIdleTime, uint16(0)),
					},
				},
			},
		},
		{
			name: "user is idle",
			givenSessionFn: func() *Session {
				s := NewSession()
				s.SetSignonTime(time.Unix(1, 0))
				// now() returns T=1000 when SetIdle() is called
				s.nowFn = func() time.Time { return time.Unix(1000, 0) }
				s.SetIdle(1 * time.Second)
				// now() returns T=2000 when TLVUserInfo() is called
				s.nowFn = func() time.Time { return time.Unix(2000, 0) }
				return s
			},
			want: wire.TLVUserInfo{
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceUserInfoSignonTOD, uint32(1)),
						wire.NewTLV(wire.OServiceUserInfoUserFlags, uint16(0x0010)),
						wire.NewTLV(wire.OServiceUserInfoStatus, uint16(0x0000)),
						wire.NewTLV(wire.OServiceUserInfoIdleTime, uint16(1001)),
					},
				},
			},
		},
		{
			name: "user goes idle then returns",
			givenSessionFn: func() *Session {
				s := NewSession()
				s.SetSignonTime(time.Unix(1, 0))
				s.SetIdle(1 * time.Second)
				s.UnsetIdle()
				return s
			},
			want: wire.TLVUserInfo{
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceUserInfoSignonTOD, uint32(1)),
						wire.NewTLV(wire.OServiceUserInfoUserFlags, uint16(0x0010)),
						wire.NewTLV(wire.OServiceUserInfoStatus, uint16(0x0000)),
						wire.NewTLV(wire.OServiceUserInfoIdleTime, uint16(0)),
					},
				},
			},
		},
		{
			name: "user has capabilities",
			givenSessionFn: func() *Session {
				s := NewSession()
				s.SetSignonTime(time.Unix(1, 0))
				s.SetCaps([][16]byte{
					{
						// chat: "748F2420-6287-11D1-8222-444553540000"
						0x74, 0x8f, 0x24, 0x20, 0x62, 0x87, 0x11, 0xd1,
						0x82, 0x22, 0x44, 0x45, 0x53, 0x54, 0x00, 0x00,
					},
					{
						// chat2: "748F2420-6287-11D1-8222-444553540000"
						0x75, 0x8f, 0x24, 0x20, 0x62, 0x87, 0x11, 0xd1,
						0x82, 0x22, 0x44, 0x45, 0x53, 0x54, 0x00, 0x01,
					},
				})
				return s
			},
			want: wire.TLVUserInfo{
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceUserInfoSignonTOD, uint32(1)),
						wire.NewTLV(wire.OServiceUserInfoUserFlags, uint16(0x0010)),
						wire.NewTLV(wire.OServiceUserInfoStatus, uint16(0x0000)),
						wire.NewTLV(wire.OServiceUserInfoIdleTime, uint16(0)),
						wire.NewTLV(wire.OServiceUserInfoOscarCaps, []byte{
							// chat: "748F2420-6287-11D1-8222-444553540000"
							0x74, 0x8f, 0x24, 0x20, 0x62, 0x87, 0x11, 0xd1,
							0x82, 0x22, 0x44, 0x45, 0x53, 0x54, 0x00, 0x00,
							// chat: "748F2420-6287-11D1-8222-444553540000"
							0x75, 0x8f, 0x24, 0x20, 0x62, 0x87, 0x11, 0xd1,
							0x82, 0x22, 0x44, 0x45, 0x53, 0x54, 0x00, 0x01,
						}),
					},
				},
			},
		},
		{
			name: "user has buddy icon",
			givenSessionFn: func() *Session {
				s := NewSession()
				s.SetSignonTime(time.Unix(1, 0))
				return s
			},
			want: wire.TLVUserInfo{
				WarningLevel: 0,
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.OServiceUserInfoSignonTOD, uint32(1)),
						wire.NewTLV(wire.OServiceUserInfoUserFlags, uint16(0x0010)),
						wire.NewTLV(wire.OServiceUserInfoStatus, uint16(0x0000)),
						wire.NewTLV(wire.OServiceUserInfoIdleTime, uint16(0)),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.givenSessionFn()
			assert.Equal(t, tt.want, s.TLVUserInfo())
		})
	}
}

func TestSession_SendAndRecvMessage_ExpectSessSendOK(t *testing.T) {
	s := NewSession()

	msg := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer s.Close()
		status := s.RelayMessage(msg)
		assert.Equal(t, SessSendOK, status)
	}()

loop:
	for {
		select {
		case m := <-s.ReceiveMessage():
			assert.Equal(t, msg, m)
		case <-s.Closed():
			break loop
		}
	}

	wg.Wait()
}

func TestSession_SendMessage_SessSendClosed(t *testing.T) {
	s := Session{
		msgCh:  make(chan wire.SNACMessage, 1),
		stopCh: make(chan struct{}),
	}
	s.Close()
	if res := s.RelayMessage(wire.SNACMessage{}); res != SessSendClosed {
		t.Fatalf("expected SessSendClosed, got %+v", res)
	}
}

func TestSession_SendMessage_SessQueueFull(t *testing.T) {
	bufSize := 10
	s := Session{
		msgCh:  make(chan wire.SNACMessage, bufSize),
		stopCh: make(chan struct{}),
	}
	for i := 0; i < bufSize; i++ {
		assert.Equal(t, SessSendOK, s.RelayMessage(wire.SNACMessage{}))
	}
	assert.Equal(t, SessQueueFull, s.RelayMessage(wire.SNACMessage{}))
}

func TestSession_Close_Twice(t *testing.T) {
	s := Session{
		stopCh: make(chan struct{}),
	}
	s.Close()
	s.Close() // make sure close is idempotent
	if !s.closed {
		t.Fatal("expected session to be closed")
	}
	select {
	case <-s.Closed():
	case <-time.After(1 * time.Second):
		t.Fatalf("channel is not closed")
	}
}

func TestSession_Close(t *testing.T) {
	s := NewSession()
	select {
	case <-s.Closed():
		assert.Fail(t, "channel is closed")
	default:
		// channel is open by default
	}
	s.Close()
	<-s.Closed()
}
