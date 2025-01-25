package toc

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"github.com/stretchr/testify/assert"
)

func TestOSCARProxy_RecvBOS_UpdateBuddyArrival(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenMsg is the incoming SNAC
		givenMsg wire.SNACMessage
		// wantCmd is the expected TOC response
		wantCmd []byte
	}{
		{
			name: "send buddy arrival - buddy online",
			me:   newTestSession("me"),
			givenMsg: wire.SNACMessage{
				Body: wire.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: wire.TLVUserInfo{
						ScreenName:   "me",
						WarningLevel: 10,
						TLVBlock: wire.TLVBlock{
							TLVList: wire.TLVList{
								wire.NewTLVBE(wire.OServiceUserInfoSignonTOD, uint32(1234)),
								wire.NewTLVBE(wire.OServiceUserInfoIdleTime, uint16(5678)),
							},
						},
					},
				},
			},
			wantCmd: []byte("UPDATE_BUDDY:me:T:10:1234:5678: O "),
		},
		{
			name: "send buddy arrival - buddy away",
			me:   newTestSession("me"),
			givenMsg: wire.SNACMessage{
				Body: wire.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: wire.TLVUserInfo{
						ScreenName:   "me",
						WarningLevel: 10,
						TLVBlock: wire.TLVBlock{
							TLVList: wire.TLVList{
								wire.NewTLVBE(wire.OServiceUserInfoSignonTOD, uint32(1234)),
								wire.NewTLVBE(wire.OServiceUserInfoIdleTime, uint16(5678)),
								wire.NewTLVBE(wire.OServiceUserInfoUserFlags, wire.OServiceUserFlagUnavailable),
							},
						},
					},
				},
			},
			wantCmd: []byte("UPDATE_BUDDY:me:T:10:1234:5678: OU"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			svc := OSCARProxy{
				Logger: slog.Default(),
			}

			ch := make(chan []byte)
			wg := &sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer wg.Done()
				err := svc.RecvBOS(ctx, tc.me, nil, ch)
				assert.NoError(t, err)
			}()

			status := tc.me.RelayMessage(tc.givenMsg)
			assert.Equal(t, state.SessSendOK, status)

			gotCmd := <-ch
			assert.Equal(t, string(tc.wantCmd), string(gotCmd))

			cancel()
			wg.Wait()
		})
	}
}

func TestOSCARProxy_RecvBOS_UpdateBuddyDeparted(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenMsg is the incoming SNAC
		givenMsg wire.SNACMessage
		// wantCmd is the expected TOC response
		wantCmd []byte
	}{
		{
			name: "send buddy departure",
			me:   newTestSession("me"),
			givenMsg: wire.SNACMessage{
				Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
					TLVUserInfo: wire.TLVUserInfo{
						ScreenName: "me",
					},
				},
			},
			wantCmd: []byte("UPDATE_BUDDY:me:F:0:0:0:   "),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			svc := OSCARProxy{
				Logger: slog.Default(),
			}

			ch := make(chan []byte)
			wg := &sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer wg.Done()
				err := svc.RecvBOS(ctx, tc.me, nil, ch)
				assert.NoError(t, err)
			}()

			status := tc.me.RelayMessage(tc.givenMsg)
			assert.Equal(t, state.SessSendOK, status)

			gotCmd := <-ch
			assert.Equal(t, string(tc.wantCmd), string(gotCmd))

			cancel()
			wg.Wait()
		})
	}
}
