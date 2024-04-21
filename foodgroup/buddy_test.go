package foodgroup

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestBuddyService_RightsQuery(t *testing.T) {
	svc := NewBuddyService(nil, nil, nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyRightsReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x03_0x03_BuddyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.BuddyTLVTagsParmMaxBuddies, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxWatchers, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxIcqBroad, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxTempBuddies, uint16(100)),
				},
			},
		},
	}
	have := svc.RightsQuery(nil, wire.SNACFrame{RequestID: 1234})

	assert.Equal(t, want, have)
}

func TestBuddyService_AddBuddies(t *testing.T) {
	tests := []struct {
		// name is the name of the test
		name string
		// sess is the client session
		sess *state.Session
		// bodyIn is the input SNAC
		bodyIn wire.SNAC_0x03_0x04_BuddyAddBuddies
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// wantErr is the expected error
		wantErr error
	}{
		{
			name: "add 2 buddies, sign-on complete",
			sess: newTestSession("user_screen_name", sessOptSignonComplete),
			bodyIn: wire.SNAC_0x03_0x04_BuddyAddBuddies{
				Buddies: []struct {
					ScreenName string `len_prefix:"uint8"`
				}{
					{
						ScreenName: "buddy_1_online",
					},
					{
						ScreenName: "buddy_2_offline",
					},
				},
			},
			mockParams: mockParams{
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					addBuddyParams: addBuddyParams{
						{
							userScreenName:  "user_screen_name",
							buddyScreenName: "buddy_1_online",
						},
						{
							userScreenName:  "user_screen_name",
							buddyScreenName: "buddy_2_offline",
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_1_online",
							sess:       newTestSession("buddy_1_online", sessOptCannedSignonTime),
						},
						{
							screenName: "buddy_2_offline",
							sess:       nil,
						},
					},
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: "user_screen_name",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_1_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: "buddy_1_online",
						},
					},
				},
			},
		},
		{
			name: "add 2 buddies, sign-on not complete",
			sess: newTestSession("user_screen_name"),
			bodyIn: wire.SNAC_0x03_0x04_BuddyAddBuddies{
				Buddies: []struct {
					ScreenName string `len_prefix:"uint8"`
				}{
					{
						ScreenName: "buddy_1_online",
					},
					{
						ScreenName: "buddy_2_offline",
					},
				},
			},
			mockParams: mockParams{
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					addBuddyParams: addBuddyParams{
						{
							userScreenName:  "user_screen_name",
							buddyScreenName: "buddy_1_online",
						},
						{
							userScreenName:  "user_screen_name",
							buddyScreenName: "buddy_2_offline",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess)
			}
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tt.mockParams.feedbagParams {
				feedbagManager.EXPECT().
					Feedbag(params.screenName).
					Return(params.results, nil)
			}

			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tt.mockParams.addBuddyParams {
				legacyBuddyListManager.EXPECT().
					AddBuddy(params.userScreenName, params.buddyScreenName)
			}

			svc := NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager)

			haveErr := svc.AddBuddies(nil, tt.sess, tt.bodyIn)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}

func TestBuddyService_DelBuddies(t *testing.T) {
	tests := []struct {
		// name is the name of the test
		name string
		// sess is the client session
		sess *state.Session
		// bodyIn is the input SNAC
		bodyIn wire.SNAC_0x03_0x05_BuddyDelBuddies
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// wantErr is the expected error
		wantErr error
	}{
		{
			name: "delete 2 buddies",
			sess: newTestSession("user_screen_name", sessOptSignonComplete),
			bodyIn: wire.SNAC_0x03_0x05_BuddyDelBuddies{
				Buddies: []struct {
					ScreenName string `len_prefix:"uint8"`
				}{
					{
						ScreenName: "buddy_1_online",
					},
					{
						ScreenName: "buddy_2_offline",
					},
				},
			},
			mockParams: mockParams{
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					deleteBuddyParams: deleteBuddyParams{
						{
							userScreenName:  "user_screen_name",
							buddyScreenName: "buddy_1_online",
						},
						{
							userScreenName:  "user_screen_name",
							buddyScreenName: "buddy_2_offline",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tt.mockParams.deleteBuddyParams {
				legacyBuddyListManager.EXPECT().
					DeleteBuddy(params.userScreenName, params.buddyScreenName)
			}

			svc := NewBuddyService(nil, nil, legacyBuddyListManager)

			svc.DelBuddies(nil, tt.sess, tt.bodyIn)
		})
	}
}
