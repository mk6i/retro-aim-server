package foodgroup

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func TestOServiceServiceForBOS_ServiceRequest(t *testing.T) {
	chatRoom := state.NewChatRoom("the-chat-room", state.NewIdentScreenName(""), state.PrivateExchange)

	cases := []struct {
		// name is the unit test name
		name string
		// config is the application config
		cfg config.Config
		// userSession is the session of the user requesting the chat service
		// info
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// expectSNACFrame is the SNAC frame sent from the server to the recipient
		// client
		expectOutput wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name:        "request info for ICBM service, return invalid SNAC err",
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.ICBM,
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeServiceUnavailable,
				},
			},
		},
		{
			name: "request info for connecting to chat nav, return chat nav connection metadata",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				ChatNavPort: "1234",
			},
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.ChatNav,
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceServiceResponse,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, []byte("the-cookie")),
							wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.ChatNav),
							wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
							wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
						},
					},
				},
			},
			mockParams: mockParams{
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: []byte{
								0x02, 'm', 'e',
								0x0, // no client ID
							},
							cookieOut: []byte("the-cookie"),
						},
					},
				},
			},
		},
		{
			name: "request info for connecting to alert svc, return alert svc connection metadata",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				AlertPort: "1234",
			},
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.Alert,
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceServiceResponse,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, []byte("the-cookie")),
							wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.Alert),
							wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
							wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
						},
					},
				},
			},
			mockParams: mockParams{
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: []byte{
								0x02, 'm', 'e',
								0x0, // no client ID
							},
							cookieOut: []byte("the-cookie"),
						},
					},
				},
			},
		},
		{
			name: "request info for connecting to admin svc, return admin svc connection metadata",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				AdminPort: "1234",
			},
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.Admin,
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceServiceResponse,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, []byte("the-cookie")),
							wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.Admin),
							wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
							wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
						},
					},
				},
			},
			mockParams: mockParams{
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: []byte{
								0x02, 'm', 'e',
								0x0, // no client ID
							},
							cookieOut: []byte("the-cookie"),
						},
					},
				},
			},
		},
		{
			name: "request info for connecting to chat room, return chat service and chat room metadata",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  "1234",
			},
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.Chat,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
								Exchange:       chatRoom.Exchange(),
								Cookie:         chatRoom.Cookie(),
								InstanceNumber: chatRoom.InstanceNumber(),
							}),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceServiceResponse,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, []byte("the-auth-cookie")),
							wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.Chat),
							wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
							wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
						},
					},
				},
			},
			mockParams: func() mockParams {
				return mockParams{
					chatRoomRegistryParams: chatRoomRegistryParams{
						chatRoomByCookieParams: chatRoomByCookieParams{
							{
								cookie: chatRoom.Cookie(),
								room:   chatRoom,
							},
						},
					},
					cookieBakerParams: cookieBakerParams{
						cookieIssueParams: cookieIssueParams{
							{
								dataIn: []byte{
									0x11, '4', '-', '0', '-', 't', 'h', 'e', '-', 'c', 'h', 'a', 't', '-', 'r', 'o', 'o', 'm',
									0x02, 'm', 'e',
								},
								cookieOut: []byte("the-auth-cookie"),
							},
						},
					},
				}
			}(),
		},
		{
			name: "request info for connecting to non-existent chat room, return ErrChatRoomNotFound",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  "1234",
			},
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.Chat,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
								Exchange:       8,
								Cookie:         "the-chat-cookie",
								InstanceNumber: 16,
							}),
						},
					},
				},
			},
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByCookieParams: chatRoomByCookieParams{
						{
							cookie: "the-chat-cookie",
							err:    state.ErrChatRoomNotFound,
						},
					},
				},
			},
			expectErr: state.ErrChatRoomNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			chatRoomManager := newMockChatRoomRegistry(t)
			for _, params := range tc.mockParams.chatRoomByCookieParams {
				chatRoomManager.EXPECT().
					ChatRoomByCookie(params.cookie).
					Return(params.room, params.err)
			}
			cookieIssuer := newMockCookieBaker(t)
			for _, params := range tc.mockParams.cookieIssueParams {
				cookieIssuer.EXPECT().
					Issue(params.dataIn).
					Return(params.cookieOut, params.err)
			}
			//
			// send input SNAC
			//
			svc := NewOServiceServiceForBOS(tc.cfg, nil, slog.Default(), cookieIssuer, chatRoomManager, nil, nil)

			outputSNAC, err := svc.ServiceRequest(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x01_0x04_OServiceServiceRequest))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestSetUserInfoFields(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user whose info is being set
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// expectOutput is the SNAC reply sent from the server back to the
		// client
		expectOutput wire.SNACMessage
		// broadcastMessage is the arrival/departure message sent to buddies
		broadcastMessage []struct {
			recipients []string
			msg        wire.SNACMessage
		}
		// interestedUserLookups contains all the users who have this user on
		// their buddy list
		interestedUserLookups map[string][]string
		// expectErr is the expected error returned
		expectErr error
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "set user status to visible",
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.OServiceUserInfoStatus, uint32(0x0000)),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceUserInfoUpdate,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: newTestSession("me").TLVUserInfo(),
				},
			},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("me"),
						},
					},
				},
			},
		},
		{
			name:        "set user status to invisible",
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.OServiceUserInfoStatus, uint32(0x0100)),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceUserInfoUpdate,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: newTestSession("me", sessOptInvisible).TLVUserInfo(),
				},
			},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyDepartedParams: broadcastBuddyDepartedParams{
						{
							screenName: state.NewIdentScreenName("me"),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buddyUpdateBroadcaster := newMockbuddyBroadcaster(t)
			for _, params := range tc.mockParams.broadcastBuddyArrivedParams {
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, matchSession(params.screenName)).
					Return(params.err)
			}
			for _, params := range tc.mockParams.broadcastBuddyDepartedParams {
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyDeparted(mock.Anything, matchSession(params.screenName)).
					Return(params.err)
			}
			svc := OServiceService{
				cfg:              config.Config{},
				logger:           slog.Default(),
				buddyBroadcaster: buddyUpdateBroadcaster,
			}
			outputSNAC, err := svc.SetUserInfoFields(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestOServiceService_RateParamsQuery(t *testing.T) {
	expectRateGroups := []struct {
		ID    uint16
		Pairs []struct {
			FoodGroup uint16
			SubGroup  uint16
		} `oscar:"count_prefix=uint16"`
	}{
		{
			ID: 1,
			Pairs: []struct {
				FoodGroup uint16
				SubGroup  uint16
			}{
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceErr,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceClientOnline,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceHostOnline,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceServiceRequest,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceServiceResponse,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceRateParamsQuery,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceRateParamsReply,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceRateParamsSubAdd,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceRateDelParamSub,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceRateParamChange,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServicePauseReq,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServicePauseAck,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceResume,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceUserInfoQuery,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceUserInfoUpdate,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceEvilNotification,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceIdleNotification,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceMigrateGroups,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceMotd,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceSetPrivacyFlags,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceWellKnownUrls,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceNoop,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceClientVersions,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceHostVersions,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceMaxConfigQuery,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceMaxConfigReply,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceStoreConfig,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceConfigQuery,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceConfigReply,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceSetUserInfoFields,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceProbeReq,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceProbeAck,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceBartReply,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceBartQuery2,
				},
				{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceBartReply2,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateErr,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateRightsQuery,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateRightsReply,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateSetInfo,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateUserInfoQuery,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateUserInfoReply,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateWatcherSubRequest,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateWatcherNotification,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateSetDirInfo,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateSetDirReply,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateGetDirInfo,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateGetDirReply,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateGroupCapabilityQuery,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateGroupCapabilityReply,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateSetKeywordInfo,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateSetKeywordReply,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateGetKeywordInfo,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateGetKeywordReply,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateFindListByEmail,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateFindListReply,
				},
				{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateUserInfoQuery2,
				},

				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyErr,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyRightsQuery,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyRightsReply,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyAddBuddies,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyDelBuddies,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyWatcherListQuery,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyWatcherListResponse,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyWatcherSubRequest,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyWatcherNotification,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyRejectNotification,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyArrived,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyDeparted,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyAddTempBuddies,
				},
				{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyDelTempBuddies,
				},

				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMAddParameters,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMDelParameters,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMParameterQuery,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMParameterReply,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMChannelMsgToHost,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMChannelMsgToClient,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMEvilRequest,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMEvilReply,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMMissedCalls,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMClientErr,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMHostAck,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMSinStored,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMSinListQuery,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMSinListReply,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMSinRetrieve,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMSinDelete,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMNotifyRequest,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMNotifyReply,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMClientEvent,
				},
				{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMSinReply,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavErr,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavRequestChatRights,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavRequestExchangeInfo,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavRequestRoomInfo,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavRequestMoreRoomInfo,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavRequestOccupantList,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavSearchForRoom,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavCreateRoom,
				},
				{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavNavInfo,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatErr,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatRoomInfoUpdate,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatUsersJoined,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatUsersLeft,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatChannelMsgToHost,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatChannelMsgToClient,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatEvilRequest,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatEvilReply,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatClientErr,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatPauseRoomReq,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatPauseRoomAck,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatResumeRoom,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatShowMyRow,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatShowRowByUsername,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatShowRowByNumber,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatShowRowByName,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatRowInfo,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatListRows,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatRowListInfo,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatMoreRows,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatMoveToRow,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatToggleChat,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatSendQuestion,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatSendComment,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatTallyVote,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatAcceptBid,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatSendInvite,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatDeclineInvite,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatAcceptInvite,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatNotifyMessage,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatGotoRow,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatStageUserJoin,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatStageUserLeft,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatUnnamedSnac22,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatClose,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatUserBan,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatUserUnban,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatJoined,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatUnnamedSnac27,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatUnnamedSnac28,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatUnnamedSnac29,
				},
				{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatRoomInfoOwner,
				},

				{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTErr,
				},
				{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTUploadQuery,
				},
				{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTUploadReply,
				},
				{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTDownloadQuery,
				},
				{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTDownloadReply,
				},
				{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTDownload2Query,
				},
				{
					FoodGroup: wire.BART,
					SubGroup:  wire.BARTDownload2Reply,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagErr,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRightsQuery,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRightsReply,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagQuery,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagQueryIfModified,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagReply,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagUse,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagInsertItem,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagUpdateItem,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagDeleteItem,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagInsertClass,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagUpdateClass,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagDeleteClass,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagReplyNotModified,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagDeleteUser,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStartCluster,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagEndCluster,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagAuthorizeBuddy,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagPreAuthorizeBuddy,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagPreAuthorizedBuddy,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRemoveMe,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRemoveMe2,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRequestAuthorizeToHost,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRequestAuthorizeToClient,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRespondAuthorizeToHost,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRespondAuthorizeToClient,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagBuddyAdded,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRequestAuthorizeToBadog,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRespondAuthorizeToBadog,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagBuddyAddedToBadog,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagTestSnac,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagForwardMsg,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagIsAuthRequiredQuery,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagIsAuthRequiredReply,
				},
				{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagRecentBuddyUpdate,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPErr,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginRequest,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginResponse,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPRegisterRequest,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPChallengeRequest,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPChallengeResponse,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPAsasnRequest,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPSecuridRequest,
				},
				{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPRegistrationImageRequest,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertErr,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertSetAlertRequest,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertSetAlertReply,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertGetSubsRequest,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertGetSubsResponse,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertNotifyCapabilities,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertNotify,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertGetRuleRequest,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertGetRuleReply,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertGetFeedRequest,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertGetFeedReply,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertRefreshFeed,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertEvent,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertQogSnac,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertRefreshFeedStock,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertNotifyTransport,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertSetAlertRequestV2,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertSetAlertReplyV2,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertTransitReply,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertNotifyAck,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertNotifyDisplayCapabilities,
				},
				{
					FoodGroup: wire.Alert,
					SubGroup:  wire.AlertUserOnline,
				},
				{
					FoodGroup: wire.ICQ,
					SubGroup:  wire.ICQErr,
				},
				{
					FoodGroup: wire.ICQ,
					SubGroup:  wire.ICQDBQuery,
				},
				{
					FoodGroup: wire.ICQ,
					SubGroup:  wire.ICQDBReply,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyErr,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyRightsQuery,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyRightsReply,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenySetGroupPermitMask,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyAddPermListEntries,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyDelPermListEntries,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyAddDenyListEntries,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyDelDenyListEntries,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyBosErr,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyAddTempPermitListEntries,
				},
				{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyDelTempPermitListEntries,
				},
				{
					FoodGroup: wire.ODir,
					SubGroup:  wire.ODirErr,
				},
				{
					FoodGroup: wire.ODir,
					SubGroup:  wire.ODirInfoQuery,
				},
				{
					FoodGroup: wire.ODir,
					SubGroup:  wire.ODirInfoReply,
				},
				{
					FoodGroup: wire.ODir,
					SubGroup:  wire.ODirKeywordListQuery,
				},
				{
					FoodGroup: wire.ODir,
					SubGroup:  wire.ODirKeywordListReply,
				},
				{
					FoodGroup: wire.UserLookup,
					SubGroup:  wire.UserLookupFindByEmail,
				},
			},
		},
	}

	cases := []struct {
		// name is the unit test name
		name string
		// config is the application config
		cfg config.Config
		// userSession is the session of the user requesting the chat service
		// info
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// expectSNACFrame is the SNAC frame sent from the server to the recipient
		// client
		expectOutput wire.SNACMessage
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name:        "get rate limits for non-AIM 1.x client",
			userSession: newTestSession("me"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{RequestID: 1234},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceRateParamsReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x07_OServiceRateParamsReply{
					RateClasses: []struct {
						ID              uint16
						WindowSize      uint32
						ClearLevel      uint32
						AlertLevel      uint32
						LimitLevel      uint32
						DisconnectLevel uint32
						CurrentLevel    uint32
						MaxLevel        uint32
						V2Params        *struct {
							LastTime     uint32
							CurrentState uint8
						} `oscar:"optional"`
					}{
						{
							ID:              0x0001,
							WindowSize:      0x00000050,
							ClearLevel:      0x000009C4,
							AlertLevel:      0x000007D0,
							LimitLevel:      0x000005DC,
							DisconnectLevel: 0x00000320,
							CurrentLevel:    0x00000D69,
							MaxLevel:        0x00001770,
							V2Params: &struct {
								LastTime     uint32
								CurrentState uint8
							}{
								LastTime:     0x00000000,
								CurrentState: 0x00,
							},
						},
					},
					RateGroups: expectRateGroups,
				},
			},
		},
		{
			name:        "get rate limits for AIM 1.x client",
			userSession: newTestSession("me", sessClientID("AOL Instant Messenger (TM), version 1.")),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{RequestID: 1234},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceRateParamsReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x07_OServiceRateParamsReply{
					RateClasses: []struct {
						ID              uint16
						WindowSize      uint32
						ClearLevel      uint32
						AlertLevel      uint32
						LimitLevel      uint32
						DisconnectLevel uint32
						CurrentLevel    uint32
						MaxLevel        uint32
						V2Params        *struct {
							LastTime     uint32
							CurrentState uint8
						} `oscar:"optional"`
					}{
						{
							ID:              0x0001,
							WindowSize:      0x00000050,
							ClearLevel:      0x000009C4,
							AlertLevel:      0x000007D0,
							LimitLevel:      0x000005DC,
							DisconnectLevel: 0x00000320,
							CurrentLevel:    0x00000D69,
							MaxLevel:        0x00001770,
						},
					},
					RateGroups: expectRateGroups,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := OServiceService{
				cfg:    config.Config{},
				logger: slog.Default(),
			}
			have := svc.RateParamsQuery(nil, tc.userSession, tc.inputSNAC.Frame)
			assert.Equal(t, tc.expectOutput, have)
		})
	}
}

func TestOServiceServiceForBOS_OServiceHostOnline(t *testing.T) {
	cookieIssuer := newMockCookieBaker(t)
	svc := NewOServiceServiceForBOS(config.Config{}, nil, slog.Default(), cookieIssuer, nil, nil, nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.Alert,
				wire.BART,
				wire.Buddy,
				wire.Feedbag,
				wire.ICBM,
				wire.ICQ,
				wire.Locate,
				wire.OService,
				wire.PermitDeny,
				wire.UserLookup,
			},
		},
	}

	have := svc.HostOnline()
	assert.Equal(t, want, have)
}

func TestOServiceServiceForChat_OServiceHostOnline(t *testing.T) {
	svc := NewOServiceServiceForChat(config.Config{}, slog.Default(), nil, nil, nil, nil, nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.OService,
				wire.Chat,
			},
		},
	}

	have := svc.HostOnline()
	assert.Equal(t, want, have)
}

func TestOServiceService_ClientVersions(t *testing.T) {
	svc := OServiceService{
		cfg:    config.Config{},
		logger: slog.Default(),
	}

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostVersions,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: []uint16{5, 6, 7, 8},
		},
	}

	have := svc.ClientVersions(nil, wire.SNACFrame{
		RequestID: 1234,
	}, wire.SNAC_0x01_0x17_OServiceClientVersions{
		Versions: []uint16{5, 6, 7, 8},
	})

	assert.Equal(t, want, have)
}

func TestOServiceService_UserInfoQuery(t *testing.T) {
	svc := OServiceService{
		cfg:    config.Config{},
		logger: slog.Default(),
	}
	sess := newTestSession("me")

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoUpdate,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}

	have := svc.UserInfoQuery(nil, sess, wire.SNACFrame{RequestID: 1234})

	assert.Equal(t, want, have)
}

func TestOServiceService_IdleNotification(t *testing.T) {
	tests := []struct {
		name   string
		sess   *state.Session
		bodyIn wire.SNAC_0x01_0x11_OServiceIdleNotification
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		wantErr    error
	}{
		{
			name: "set idle from active",
			sess: newTestSession("me"),
			bodyIn: wire.SNAC_0x01_0x11_OServiceIdleNotification{
				IdleTime: 90,
			},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("me"),
						},
					},
				},
			},
		},
		{
			name: "set active from idle",
			sess: newTestSession("me", sessOptIdle(90*time.Second)),
			bodyIn: wire.SNAC_0x01_0x11_OServiceIdleNotification{
				IdleTime: 0,
			},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("me"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buddyUpdateBroadcaster := newMockbuddyBroadcaster(t)
			for _, params := range tt.mockParams.broadcastBuddyArrivedParams {
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, matchSession(params.screenName)).
					Return(params.err)
			}
			svc := OServiceService{
				cfg:              config.Config{},
				logger:           slog.Default(),
				buddyBroadcaster: buddyUpdateBroadcaster,
			}
			haveErr := svc.IdleNotification(nil, tt.sess, tt.bodyIn)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}

func TestOServiceServiceForBOS_ClientOnline(t *testing.T) {
	tests := []struct {
		// name is the name of the test
		name string
		// joiningChatter is the session of the arriving user
		sess *state.Session
		// bodyIn is the SNAC body sent from the arriving user's client to the
		// server
		bodyIn wire.SNAC_0x01_0x02_OServiceClientOnline
		// wantErr is the expected error from the handler
		wantErr error
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// wantSess is the expected session state after the method is called
		wantSess *state.Session
	}{
		{
			name:   "notify that user is online",
			sess:   newTestSession("me", sessOptCannedSignonTime),
			bodyIn: wire.SNAC_0x01_0x02_OServiceClientOnline{},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastVisibilityParams: broadcastVisibilityParams{
						{
							from:             state.NewIdentScreenName("me"),
							filter:           nil,
							doSendDepartures: false,
						},
					},
				},
			},
			wantSess: newTestSession("me", sessOptCannedSignonTime, sessOptSignonComplete),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buddyUpdateBroadcaster := newMockbuddyBroadcaster(t)
			for _, params := range tt.mockParams.broadcastVisibilityParams {
				buddyUpdateBroadcaster.EXPECT().
					BroadcastVisibility(mock.Anything, matchSession(params.from), params.filter, params.doSendDepartures).
					Return(params.err)
			}

			svc := NewOServiceServiceForBOS(config.Config{}, nil, slog.Default(), nil, nil, nil, nil)
			svc.buddyBroadcaster = buddyUpdateBroadcaster
			haveErr := svc.ClientOnline(nil, tt.bodyIn, tt.sess)
			assert.ErrorIs(t, tt.wantErr, haveErr)
			assert.Equal(t, tt.wantSess.SignonComplete(), tt.sess.SignonComplete())
		})
	}
}

func TestOServiceServiceForChat_ClientOnline(t *testing.T) {
	chatRoom := state.NewChatRoom("the-chat-room", state.NewIdentScreenName("creator"), state.PrivateExchange)
	chatter1 := newTestSession("chatter-1", sessOptChatRoomCookie(chatRoom.Cookie()))
	chatter2 := newTestSession("chatter-2", sessOptChatRoomCookie(chatRoom.Cookie()))

	tests := []struct {
		// name is the name of the test
		name string
		// joiningChatter is the user joining the chat room
		joiningChatter *state.Session
		// bodyIn is the SNAC body sent from the arriving user's client to the
		// server
		bodyIn  wire.SNAC_0x01_0x02_OServiceClientOnline
		wantErr error
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:           "upon joining, send chat room metadata and participant list to joining user; alert arrival to existing participants",
			joiningChatter: chatter1,
			bodyIn:         wire.SNAC_0x01_0x02_OServiceClientOnline{},
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					chatRelayToAllExceptParams: chatRelayToAllExceptParams{
						{
							screenName: state.NewIdentScreenName("chatter-1"),
							cookie:     chatRoom.Cookie(),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatUsersJoined,
								},
								Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
									Users: []wire.TLVUserInfo{
										chatter1.TLVUserInfo(),
									},
								},
							},
						},
					},
					chatAllSessionsParams: chatAllSessionsParams{
						{
							cookie: chatRoom.Cookie(),
							sessions: []*state.Session{
								chatter1,
								chatter2,
							},
						},
					},
					chatRelayToScreenNameParams: chatRelayToScreenNameParams{
						{
							cookie:     chatRoom.Cookie(),
							screenName: chatter1.IdentScreenName(),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatRoomInfoUpdate,
								},
								Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
									Exchange:       chatRoom.Exchange(),
									Cookie:         chatRoom.Cookie(),
									InstanceNumber: chatRoom.InstanceNumber(),
									DetailLevel:    chatRoom.DetailLevel(),
									TLVBlock: wire.TLVBlock{
										TLVList: chatRoom.TLVList(),
									},
								},
							},
						},
						{
							cookie:     chatRoom.Cookie(),
							screenName: chatter1.IdentScreenName(),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatUsersJoined,
								},
								Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
									Users: []wire.TLVUserInfo{
										chatter1.TLVUserInfo(),
										chatter2.TLVUserInfo(),
									},
								},
							},
						},
					},
				},
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByCookieParams: chatRoomByCookieParams{
						{
							cookie: chatRoom.Cookie(),
							room:   chatRoom,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRoomManager := newMockChatRoomRegistry(t)
			for _, params := range tt.mockParams.chatRoomByCookieParams {
				chatRoomManager.EXPECT().
					ChatRoomByCookie(params.cookie).
					Return(params.room, params.err)
			}
			chatMessageRelayer := newMockChatMessageRelayer(t)
			for _, params := range tt.mockParams.chatRelayToAllExceptParams {
				chatMessageRelayer.EXPECT().
					RelayToAllExcept(mock.Anything, params.cookie, params.screenName, params.message)
			}
			for _, params := range tt.mockParams.chatAllSessionsParams {
				chatMessageRelayer.EXPECT().
					AllSessions(params.cookie).
					Return(params.sessions)
			}
			for _, params := range tt.mockParams.chatRelayToScreenNameParams {
				chatMessageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.cookie, params.screenName, params.message)
			}

			svc := NewOServiceServiceForChat(config.Config{}, slog.Default(), nil, chatRoomManager, chatMessageRelayer, nil, nil)

			haveErr := svc.ClientOnline(nil, wire.SNAC_0x01_0x02_OServiceClientOnline{}, tt.joiningChatter)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}

func TestOServiceServiceForChatNav_HostOnline(t *testing.T) {
	svc := NewOServiceServiceForChatNav(config.Config{}, slog.Default(), nil, nil, nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.ChatNav,
				wire.OService,
			},
		},
	}

	have := svc.HostOnline()
	assert.Equal(t, want, have)
}

func TestOServiceServiceForAlert_HostOnline(t *testing.T) {
	svc := NewOServiceServiceForAlert(config.Config{}, slog.Default(), nil, nil, nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.Alert,
				wire.OService,
			},
		},
	}

	have := svc.HostOnline()
	assert.Equal(t, want, have)
}

func TestOServiceService_SetPrivacyFlags(t *testing.T) {
	svc := OServiceService{
		cfg:    config.Config{},
		logger: slog.Default(),
	}
	body := wire.SNAC_0x01_0x14_OServiceSetPrivacyFlags{
		PrivacyFlags: wire.OServicePrivacyFlagMember | wire.OServicePrivacyFlagIdle,
	}
	svc.SetPrivacyFlags(context.Background(), body)
}
