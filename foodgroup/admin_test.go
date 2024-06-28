package foodgroup

import (
	"testing"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// expectErr is the expected error returned
		expectErr error
	}{
		{
			name: "user confirms their account",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminAcctConfirmRequest,
				},
				Body: wire.SNAC_0x07_0x06_AdminConfirmRequest{},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminAcctConfirmReply,
				},
				Body: wire.SNAC_0x07_0x07_AdminConfirmReply{
					Status: wire.AdminAcctConfirmStatusEmailSent,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionManager := newMockSessionManager(t)
			accountManager := newMockAccountManager(t)
			buddyBroadcaster := newMockbuddyBroadcaster(t)
			svc := AdminService{
				sessionManager:         sessionManager,
				accountManager:         accountManager,
				buddyUpdateBroadcaster: buddyBroadcaster,
			}
			outputSNAC, err := svc.ConfirmRequest(nil, tc.inputSNAC.Frame)
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
							wire.NewTLV(wire.AdminTLVRegistrationStatus, uint16(0x00))},
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
							wire.NewTLV(wire.AdminTLVRegistrationStatus, wire.AdminInfoRegStatusFullDisclosure),
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
							wire.NewTLV(wire.AdminTLVEmailAddress, uint16(0x00))},
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
							wire.NewTLV(wire.AdminTLVEmailAddress, "chattingchuck@aol.com"), // todo: get from session
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
							wire.NewTLV(wire.AdminTLVScreenNameFormatted, uint16(0x00))},
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
							wire.NewTLV(wire.AdminTLVScreenNameFormatted, "ChattingChuck"),
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
							wire.NewTLV(uint16(0x99), uint16(0x00))},
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
			sessionManager := newMockSessionManager(t)
			accountManager := newMockAccountManager(t)
			buddyBroadcaster := newMockbuddyBroadcaster(t)
			svc := AdminService{
				sessionManager:         sessionManager,
				accountManager:         accountManager,
				buddyUpdateBroadcaster: buddyBroadcaster,
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

func TestAdminService_InfoChangeRequest(t *testing.T) {
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
							wire.NewTLV(wire.AdminTLVScreenNameFormatted, "Chatting Chuck"),
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
							wire.NewTLV(wire.AdminTLVScreenNameFormatted, "Chatting Chuck"),
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
							wire.NewTLV(wire.AdminTLVScreenNameFormatted, "c  h  a  t  t  i  n  g  c  h  u  c  k")},
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
					Code: wire.AdminInfoErrorInvalidNickNameLength,
				},
			},
		},
		{
			name:        "proposed screen name does not match session's screen name (malicous client)",
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
							wire.NewTLV(wire.AdminTLVScreenNameFormatted, "QuietQuinton")},
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
					Code: wire.AdminInfoErrorInvalidNickName,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionManager := newMockSessionManager(t)
			accountManager := newMockAccountManager(t)
			buddyBroadcaster := newMockbuddyBroadcaster(t)

			for _, params := range tc.mockParams.accountManagerParams.accountManagerUpdateDisplayScreenNameParams {
				accountManager.EXPECT().
					UpdateDisplayScreenName(params.displayScreenName).
					Return(params.err)
			}

			for _, params := range tc.mockParams.broadcastBuddyArrivedParams {
				p := params
				buddyBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, mock.MatchedBy(func(s *state.Session) bool {
						return s.IdentScreenName() == p.screenName
					})).
					Return(nil)
			}

			svc := AdminService{
				sessionManager:         sessionManager,
				accountManager:         accountManager,
				buddyUpdateBroadcaster: buddyBroadcaster,
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
