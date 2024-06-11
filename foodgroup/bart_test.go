package foodgroup

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func TestBARTService_UpsertItem(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
	}{
		{
			name:        "upsert item",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x10_0x02_BARTUploadQuery{
					Type: 1,
					Data: []byte{'i', 't', 'e', 'm', 'd', 'a', 't', 'a'},
				},
			},
			mockParams: mockParams{
				bartManagerParams: bartManagerParams{
					bartManagerUpsertParams: bartManagerUpsertParams{
						{
							itemHash: []byte{0x4e, 0xd9, 0xc1, 0x96, 0x45, 0xdb, 0x5a, 0xec, 0xdb, 0xf5, 0xc7, 0xa2, 0x4e, 0x8e, 0xa0, 0xed},
							payload:  []byte{'i', 't', 'e', 'm', 'd', 'a', 't', 'a'},
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTUploadReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x10_0x03_BARTUploadReply{
					Code: wire.BARTReplyCodesSuccess,
					ID: wire.BARTID{
						Type: wire.BARTTypesBuddyIcon,
						BARTInfo: wire.BARTInfo{
							Flags: wire.BARTFlagsKnown,
							Hash:  []byte{0x4e, 0xd9, 0xc1, 0x96, 0x45, 0xdb, 0x5a, 0xec, 0xdb, 0xf5, 0xc7, 0xa2, 0x4e, 0x8e, 0xa0, 0xed},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bartManager := newMockBARTManager(t)
			for _, params := range tc.mockParams.bartManagerUpsertParams {
				bartManager.EXPECT().
					BARTUpsert(params.itemHash, params.payload).
					Return(nil)
			}
			buddyUpdateBroadcaster := newMockBuddyBroadcaster(t)
			for _, params := range tc.mockParams.broadcastBuddyArrivedParams {
				p := params
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, mock.MatchedBy(func(s *state.Session) bool {
						return s.IdentScreenName() == p.screenName
					})).
					Return(nil)
			}
			svc := NewBARTService(slog.Default(), bartManager, buddyUpdateBroadcaster)

			output, err := svc.UpsertItem(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x10_0x02_BARTUploadQuery))

			assert.NoError(t, err)
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}

func TestBARTService_RetrieveItem(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// expectErr is the expected error
		expectErr error
	}{
		{
			name:        "retrieve buddy icon",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x10_0x04_BARTDownloadQuery{
					ScreenName: "user_screen_name",
					Command:    1,
					BARTID: wire.BARTID{
						Type: wire.BARTTypesBuddyIcon,
						BARTInfo: wire.BARTInfo{
							Flags: wire.BARTFlagsKnown,
							Hash:  []byte{0x4e, 0xd9, 0xc1, 0x96, 0x45, 0xdb, 0x5a, 0xec, 0xdb, 0xf5, 0xc7, 0xa2, 0x4e, 0x8e, 0xa0, 0xed},
						},
					},
				},
			},
			mockParams: mockParams{
				bartManagerParams: bartManagerParams{
					bartManagerRetrieveParams: bartManagerRetrieveParams{
						{
							itemHash: []byte{0x4e, 0xd9, 0xc1, 0x96, 0x45, 0xdb, 0x5a, 0xec, 0xdb, 0xf5, 0xc7, 0xa2, 0x4e, 0x8e, 0xa0, 0xed},
							result:   []byte{'i', 't', 'e', 'm', 'd', 'a', 't', 'a'},
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTDownloadReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x10_0x05_BARTDownloadReply{
					ScreenName: "user_screen_name",
					BARTID: wire.BARTID{
						Type: wire.BARTTypesBuddyIcon,
						BARTInfo: wire.BARTInfo{
							Flags: wire.BARTFlagsKnown,
							Hash:  []byte{0x4e, 0xd9, 0xc1, 0x96, 0x45, 0xdb, 0x5a, 0xec, 0xdb, 0xf5, 0xc7, 0xa2, 0x4e, 0x8e, 0xa0, 0xed},
						},
					},
					Data: []byte{'i', 't', 'e', 'm', 'd', 'a', 't', 'a'},
				},
			},
		},
		{
			name:        "retrieve blank icon used for clearing buddy icon",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x10_0x04_BARTDownloadQuery{
					ScreenName: "user_screen_name",
					Command:    1,
					BARTID: wire.BARTID{
						Type: wire.BARTTypesBuddyIcon,
						BARTInfo: wire.BARTInfo{
							Flags: wire.BARTFlagsKnown,
							Hash:  wire.GetClearIconHash(),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTDownloadReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x10_0x05_BARTDownloadReply{
					ScreenName: "user_screen_name",
					BARTID: wire.BARTID{
						Type: wire.BARTTypesBuddyIcon,
						BARTInfo: wire.BARTInfo{
							Flags: wire.BARTFlagsKnown,
							Hash:  wire.GetClearIconHash(),
						},
					},
					Data: blankGIF,
				},
			},
		},
		{
			name:        "retrieve unknown icon, expect err",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x10_0x04_BARTDownloadQuery{
					ScreenName: "user_screen_name",
					Command:    1,
					BARTID: wire.BARTID{
						Type: wire.BARTTypesBuddyIcon,
						BARTInfo: wire.BARTInfo{
							Flags: wire.BARTFlagsUnknown,
							Hash:  wire.GetClearIconHash(),
						},
					},
				},
			},
			expectErr: errKnownIconsOnly,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bartManager := newMockBARTManager(t)
			for _, params := range tc.mockParams.bartManagerParams.bartManagerRetrieveParams {
				bartManager.EXPECT().
					BARTRetrieve(params.itemHash).
					Return(params.result, nil)
			}

			svc := NewBARTService(slog.Default(), bartManager, nil)

			output, err := svc.RetrieveItem(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x10_0x04_BARTDownloadQuery))

			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}
