package foodgroup

import (
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func TestAdminService_ConfirmRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg config.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// userSession is the session of the user
		userSession *state.Session
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// expectErr is the expected error returned
		expectErr error
	}{
		{
			name:        "unconfirmed account sends confirmation request",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminAcctConfirmRequest,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x07_0x06_AdminConfirmRequest{},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminAcctConfirmReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x07_0x07_AdminConfirmReply{
					Status: wire.AdminAcctConfirmStatusEmailSent,
				},
			},
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerEmailAddressByNameParams: accountManagerEmailAddressByNameParams{
						{

							screenName: state.NewIdentScreenName("chattingchuck"),
							emailAddress: &mail.Address{
								Address: "chuck@aol.com",
							},
							err: nil,
						},
					},
					accountManagerConfirmStatusByNameParams: accountManagerConfirmStatusByNameParams{
						{
							screenName:    state.NewIdentScreenName("chattingchuck"),
							confirmStatus: false,
							err:           nil,
						},
					},
					accountManagerUpdateConfirmStatusParams: accountManagerUpdateConfirmStatusParams{
						{
							confirmStatus: true,
							screenName:    state.NewIdentScreenName("chattingchuck"),
							err:           nil,
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("Chatting Chuck"),
						},
					},
				},
			},
		},
		{
			name:        "already confirmed account sends confirmation request",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminAcctConfirmRequest,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x07_0x06_AdminConfirmRequest{},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminAcctConfirmReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x07_0x07_AdminConfirmReply{
					Status: wire.AdminAcctConfirmStatusAlreadyConfirmed,
				},
			},
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerEmailAddressByNameParams: accountManagerEmailAddressByNameParams{
						{

							screenName: state.NewIdentScreenName("chattingchuck"),
							emailAddress: &mail.Address{
								Address: "chuck@aol.com",
							},
							err: nil,
						},
					},
					accountManagerConfirmStatusByNameParams: accountManagerConfirmStatusByNameParams{
						{
							screenName:    state.NewIdentScreenName("chattingchuck"),
							confirmStatus: true,
							err:           nil,
						},
					},
				},
			},
		},
		{
			name:        "acccount with no email address sends confirmation request",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminAcctConfirmRequest,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x07_0x06_AdminConfirmRequest{},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminAcctConfirmReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x07_0x07_AdminConfirmReply{
					Status: wire.AdminAcctConfirmStatusServerError,
				},
			},
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerEmailAddressByNameParams: accountManagerEmailAddressByNameParams{
						{

							screenName:   state.NewIdentScreenName("chattingchuck"),
							emailAddress: nil,
							err:          state.ErrNoEmailAddress,
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			accountManager := newMockAccountManager(t)
			buddyBroadcaster := newMockbuddyBroadcaster(t)

			for _, params := range tc.mockParams.accountManagerParams.accountManagerEmailAddressByNameParams {
				accountManager.EXPECT().
					EmailAddressByName(params.screenName).
					Return(params.emailAddress, params.err)
			}
			for _, params := range tc.mockParams.accountManagerParams.accountManagerConfirmStatusByNameParams {
				accountManager.EXPECT().
					ConfirmStatusByName(params.screenName).
					Return(params.confirmStatus, params.err)
			}
			for _, params := range tc.mockParams.accountManagerParams.accountManagerUpdateConfirmStatusParams {
				accountManager.EXPECT().
					UpdateConfirmStatus(params.confirmStatus, params.screenName).
					Return(params.err)
			}
			for _, params := range tc.mockParams.broadcastBuddyArrivedParams {
				buddyBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, tc.userSession).
					Return(params.err)
			}
			svc := AdminService{
				messageRelayer:   messageRelayer,
				accountManager:   accountManager,
				buddyBroadcaster: buddyBroadcaster,
			}
			outputSNAC, err := svc.ConfirmRequest(nil, tc.userSession, tc.inputSNAC.Frame)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAdminService_InfoQuery(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg config.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// userSession is the session of the user
		userSession *state.Session
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// expectErr is the expected error returned
		expectErr error
	}{
		{
			name:        "user requests account registration status",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoQuery,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x02_AdminInfoQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVRegistrationStatus, uint16(0x00))},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x03_AdminInfoReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVRegistrationStatus, wire.AdminInfoRegStatusLimitDisclosure),
						},
					},
				},
			},
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerRegStatusByNameParams: accountManagerRegStatusByNameParams{
						{
							screenName: state.NewIdentScreenName("chattingchuck"),
							regStatus:  wire.AdminInfoRegStatusLimitDisclosure,
							err:        nil,
						},
					},
				},
			},
		},
		{
			name:        "user requests account email address",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoQuery,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x02_AdminInfoQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVEmailAddress, uint16(0x00))},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x03_AdminInfoReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVEmailAddress, "chattingchuck@aol.com"),
						},
					},
				},
			},
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerEmailAddressByNameParams: accountManagerEmailAddressByNameParams{
						{
							screenName: state.NewIdentScreenName("chattingchuck"),
							emailAddress: &mail.Address{
								Address: "chattingchuck@aol.com",
							},
							err: nil,
						},
					},
				},
			},
		},
		{
			name:        "user requests account email address but not set",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoQuery,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x02_AdminInfoQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVEmailAddress, uint16(0x00))},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x03_AdminInfoReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVEmailAddress, ""),
						},
					},
				},
			},
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerEmailAddressByNameParams: accountManagerEmailAddressByNameParams{
						{
							screenName:   state.NewIdentScreenName("chattingchuck"),
							emailAddress: nil,
							err:          state.ErrNoEmailAddress,
						},
					},
				},
			},
		},
		{
			name:        "user requests formatted screenname",
			userSession: newTestSession("ChattingChuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoQuery,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x02_AdminInfoQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVScreenNameFormatted, uint16(0x00))},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x03_AdminInfoReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVScreenNameFormatted, "ChattingChuck"),
						},
					},
				},
			},
		},
		{
			name:        "user requests invalid TLV",
			userSession: newTestSession("ChattingChuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoQuery,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x02_AdminInfoQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(uint16(0x99), uint16(0x00))},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminErr,
					RequestID: 1337,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotSupportedByHost,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			accountManager := newMockAccountManager(t)
			buddyBroadcaster := newMockbuddyBroadcaster(t)

			for _, params := range tc.mockParams.accountManagerParams.accountManagerRegStatusByNameParams {
				accountManager.EXPECT().
					RegStatusByName(params.screenName).
					Return(params.regStatus, params.err)
			}

			for _, params := range tc.mockParams.accountManagerParams.accountManagerEmailAddressByNameParams {
				accountManager.EXPECT().
					EmailAddressByName(params.screenName).
					Return(params.emailAddress, params.err)
			}

			svc := AdminService{
				messageRelayer:   messageRelayer,
				accountManager:   accountManager,
				buddyBroadcaster: buddyBroadcaster,
			}
			outputSNAC, err := svc.InfoQuery(nil, tc.userSession, tc.inputSNAC.Frame, tc.inputSNAC.Body.(wire.SNAC_0x07_0x02_AdminInfoQuery))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAdminService_InfoChangeRequest_ScreenName(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg config.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// userSession is the session of the user
		userSession *state.Session
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// expectErr is the expected error returned
		expectErr error
	}{
		{
			name:        "user changes screen name format successfully",
			userSession: newTestSession("chattingchuck"),
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerUpdateDisplayScreenNameParams: accountManagerUpdateDisplayScreenNameParams{
						{
							displayScreenName: state.DisplayScreenName("Chatting Chuck"),
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("Chatting Chuck"),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("Chatting Chuck"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.OService,
									SubGroup:  wire.OServiceUserInfoUpdate,
								},
								Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
									TLVUserInfo: newTestSession("Chatting Chuck").TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVScreenNameFormatted, "Chatting Chuck"),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVScreenNameFormatted, "Chatting Chuck"),
						},
					},
				},
			},
		},
		{
			name:        "proposed screen name is too long",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVScreenNameFormatted, "c  h  a  t  t  i  n  g  c  h  u  c  k")},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVErrorCode, wire.AdminInfoErrorInvalidNickNameLength),
							wire.NewTLVBE(wire.AdminTLVUrl, ""),
						},
					},
				},
			},
		},
		{
			name:        "proposed screen name does not match session screen name",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVScreenNameFormatted, "QuietQuinton")},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVErrorCode, wire.AdminInfoErrorValidateNickName),
							wire.NewTLVBE(wire.AdminTLVUrl, ""),
						},
					},
				},
			},
		},
		{
			name:        "proposed screen name ends in a space",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVScreenNameFormatted, "ChattingChuck ")},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVErrorCode, wire.AdminInfoErrorInvalidNickName),
							wire.NewTLVBE(wire.AdminTLVUrl, ""),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			accountManager := newMockAccountManager(t)
			mockBuddyBroadcaster := newMockbuddyBroadcaster(t)

			for _, params := range tc.mockParams.accountManagerParams.accountManagerUpdateDisplayScreenNameParams {
				accountManager.EXPECT().
					UpdateDisplayScreenName(params.displayScreenName).
					Return(params.err)
			}

			for _, params := range tc.mockParams.broadcastBuddyArrivedParams {
				mockBuddyBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, matchSession(params.screenName)).
					Return(params.err)
			}

			for _, params := range tc.mockParams.messageRelayerParams.relayToScreenNameParams {
				p := params
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, p.screenName, p.message)
			}

			svc := AdminService{
				accountManager:   accountManager,
				buddyBroadcaster: mockBuddyBroadcaster,
				messageRelayer:   messageRelayer,
			}
			outputSNAC, err := svc.InfoChangeRequest(nil, tc.userSession, tc.inputSNAC.Frame, tc.inputSNAC.Body.(wire.SNAC_0x07_0x04_AdminInfoChangeRequest))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAdminService_InfoChangeRequest_EmailAddress(t *testing.T) {
	// One case needs a 320 character long email address
	longEmailAddress := "longemailaddress@"
	for i := 0; i < 50; i++ {
		longEmailAddress += "domain"
	}
	longEmailAddress += ".com"
	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg config.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// userSession is the session of the user
		userSession *state.Session
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// expectErr is the expected error returned
		expectErr error
	}{
		{
			name:        "user changes email address successfully",
			userSession: newTestSession("chattingchuck"),
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerUpdateEmailAddressParams: accountManagerUpdateEmailAddressParams{
						{
							screenName: state.NewIdentScreenName("chattingchuck"),
							emailAddress: &mail.Address{
								Address: "chattingchuck@aol.com",
							},
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVEmailAddress, "chattingchuck@aol.com"),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVEmailAddress, "chattingchuck@aol.com"),
						},
					},
				},
			},
		},
		{
			name:        "proposed email address invalid rfc 5322 format",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVEmailAddress, "chattingchuck@@@@@@@aol.com"),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVErrorCode, wire.AdminInfoErrorInvalidEmail),
							wire.NewTLVBE(wire.AdminTLVUrl, ""),
						},
					},
				},
			},
		},
		{
			name:        "proposed email address too long",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVEmailAddress, longEmailAddress),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVErrorCode, wire.AdminInfoErrorInvalidEmailLength),
							wire.NewTLVBE(wire.AdminTLVUrl, ""),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			accountManager := newMockAccountManager(t)
			buddyBroadcaster := newMockbuddyBroadcaster(t)

			for _, params := range tc.mockParams.accountManagerParams.accountManagerUpdateEmailAddressParams {
				accountManager.EXPECT().
					UpdateEmailAddress(params.emailAddress, params.screenName).
					Return(params.err)
			}

			svc := AdminService{
				accountManager:   accountManager,
				buddyBroadcaster: buddyBroadcaster,
				messageRelayer:   messageRelayer,
			}
			outputSNAC, err := svc.InfoChangeRequest(nil, tc.userSession, tc.inputSNAC.Frame, tc.inputSNAC.Body.(wire.SNAC_0x07_0x04_AdminInfoChangeRequest))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAdminService_InfoChangeRequest_RegStatus(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg config.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// userSession is the session of the user
		userSession *state.Session
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// expectErr is the expected error returned
		expectErr error
	}{
		{
			name:        "user changes reg preference successfully",
			userSession: newTestSession("chattingchuck"),
			mockParams: mockParams{
				accountManagerParams: accountManagerParams{
					accountManagerUpdateRegStatusParams: accountManagerUpdateRegStatusParams{
						{
							screenName: state.NewIdentScreenName("chattingchuck"),
							regStatus:  wire.AdminInfoRegStatusNoDisclosure,
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVRegistrationStatus, wire.AdminInfoRegStatusNoDisclosure),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVRegistrationStatus, wire.AdminInfoRegStatusNoDisclosure),
						},
					},
				},
			},
		},
		{
			name:        "proposed reg preference invalid",
			userSession: newTestSession("chattingchuck"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeRequest,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVRegistrationStatus, uint16(0x1337)),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminInfoChangeReply,
					RequestID: 1337,
				},
				Body: wire.SNAC_0x07_0x05_AdminChangeReply{
					Permissions: wire.AdminInfoPermissionsReadWrite,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.AdminTLVErrorCode, wire.AdminInfoErrorInvalidRegistrationPreference),
							wire.NewTLVBE(wire.AdminTLVUrl, ""),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			accountManager := newMockAccountManager(t)
			buddyBroadcaster := newMockbuddyBroadcaster(t)

			for _, params := range tc.mockParams.accountManagerParams.accountManagerUpdateRegStatusParams {
				accountManager.EXPECT().
					UpdateRegStatus(params.regStatus, params.screenName).
					Return(params.err)
			}

			svc := AdminService{
				accountManager:   accountManager,
				buddyBroadcaster: buddyBroadcaster,
				messageRelayer:   messageRelayer,
			}
			outputSNAC, err := svc.InfoChangeRequest(nil, tc.userSession, tc.inputSNAC.Frame, tc.inputSNAC.Body.(wire.SNAC_0x07_0x04_AdminInfoChangeRequest))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
