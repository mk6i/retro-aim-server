package handler

import (
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_BUCPLoginRequestHandler(t *testing.T) {
	sessUUID := uuid.UUID{1, 2, 3}
	user := state.User{
		ScreenName: "screen_name",
		AuthKey:    "auth_key",
	}
	assert.NoError(t, user.HashPassword("the_password"))
	userSession := newTestSession(user.ScreenName, sessOptID(sessUUID.String()))

	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg server.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNAC_0x17_0x02_BUCPLoginRequest
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// newUserFn is the function that registers a new user account
		newUserFn func(screenName string) (state.User, error)
		// expectOutput is the SNAC sent from the server to client
		expectOutput oscar.SNACMessage
		// wantErr is the error we expect from the method
		wantErr error
	}{
		{
			name: "user provides valid credentials and logs in successfully",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, user.PassHash),
						oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.ScreenName,
							result:     &user,
						},
					},
				},
				sessionManagerParams: sessionManagerParams{
					newSessionWithSNParams: newSessionWithSNParams{
						{
							sessID:     userSession.ID(),
							screenName: user.ScreenName,
							result:     userSession,
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				Body: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
							oscar.NewTLV(oscar.TLVReconnectHere, "127.0.0.1:1234"),
							oscar.NewTLV(oscar.TLVAuthorizationCookie, sessUUID.String()),
						},
					},
				},
			},
		},
		{
			name: "user logs in with non-existent screen name--account is created and logged in successfully",
			cfg: server.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     1234,
				DisableAuth: true,
			},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, user.PassHash),
						oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.ScreenName,
							result:     nil,
						},
					},
					upsertUserParams: upsertUserParams{
						{
							user: user,
						},
					},
				},
				sessionManagerParams: sessionManagerParams{
					newSessionWithSNParams: newSessionWithSNParams{
						{
							sessID:     userSession.ID(),
							screenName: user.ScreenName,
							result:     userSession,
						},
					},
				},
			},
			newUserFn: func(screenName string) (state.User, error) {
				return user, nil
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				Body: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
							oscar.NewTLV(oscar.TLVReconnectHere, "127.0.0.1:1234"),
							oscar.NewTLV(oscar.TLVAuthorizationCookie, sessUUID.String()),
						},
					},
				},
			},
		},
		{
			name: "user logs in with invalid password--account is created and logged in successfully",
			cfg: server.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     1234,
				DisableAuth: true,
			},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, []byte("bad-password-hash")),
						oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.ScreenName,
							result:     &user,
						},
					},
					upsertUserParams: upsertUserParams{
						{
							user: user,
						},
					},
				},
				sessionManagerParams: sessionManagerParams{
					newSessionWithSNParams: newSessionWithSNParams{
						{
							sessID:     userSession.ID(),
							screenName: user.ScreenName,
							result:     userSession,
						},
					},
				},
			},
			newUserFn: func(screenName string) (state.User, error) {
				return user, nil
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				Body: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
							oscar.NewTLV(oscar.TLVReconnectHere, "127.0.0.1:1234"),
							oscar.NewTLV(oscar.TLVAuthorizationCookie, sessUUID.String()),
						},
					},
				},
			},
		},
		{
			name: "user provides invalid password--account creation fails due to user creation runtime error",
			cfg: server.Config{
				DisableAuth: true,
			},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, []byte("bad-password-hash")),
						oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.ScreenName,
							result:     &user,
						},
					},
				},
			},
			newUserFn: func(screenName string) (state.User, error) {
				return user, io.EOF
			},
			wantErr: io.EOF,
		},
		{
			name: "user provides invalid password--account creation fails due to user upsert runtime error",
			cfg: server.Config{
				DisableAuth: true,
			},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, []byte("bad-password-hash")),
						oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.ScreenName,
							result:     &user,
						},
					},
					upsertUserParams: upsertUserParams{
						{
							user: user,
							err:  io.EOF,
						},
					},
				},
			},
			newUserFn: func(screenName string) (state.User, error) {
				return user, nil
			},
			wantErr: io.EOF,
		},
		{
			name: "user provides invalid password and receives invalid login response",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, []byte("bad_password")),
						oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.ScreenName,
							result:     nil,
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				Body: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
							oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
						},
					},
				},
			},
		},
		{
			name: "login fails on user manager lookup",
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, user.PassHash),
						oscar.NewTLV(oscar.TLVScreenName, user.ScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.ScreenName,
							err:        io.EOF,
						},
					},
				},
			},
			wantErr: io.EOF,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.getUserParams {
				userManager.EXPECT().
					GetUser(params.screenName).
					Return(params.result, params.err)
			}
			for _, params := range tc.mockParams.upsertUserParams {
				userManager.EXPECT().
					UpsertUser(params.user).
					Return(params.err)
			}
			sessionManager := newMockSessionManager(t)
			for _, params := range tc.mockParams.newSessionWithSNParams {
				sessionManager.EXPECT().
					NewSessionWithSN(params.sessID, params.screenName).
					Return(params.result)
			}
			svc := AuthService{
				config:         tc.cfg,
				sessionManager: sessionManager,
				userManager:    userManager,
			}
			fnNewUUID := func() uuid.UUID {
				return sessUUID
			}
			outputSNAC, err := svc.BUCPLoginRequestHandler(tc.inputSNAC, fnNewUUID, tc.newUserFn)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAuthService_BUCPChallengeRequestHandler(t *testing.T) {
	sessUUID := uuid.UUID{1, 2, 3}
	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg server.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNAC_0x17_0x06_BUCPChallengeRequest
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput oscar.SNACMessage
		// wantErr is the error we expect from the method
		wantErr error
	}{
		{
			name: "login with valid username, expect OK login response",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_a"),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: "sn_user_a",
							result: &state.User{
								ScreenName: "sn_user_a",
								AuthKey:    "auth_key_user_a",
							},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPChallengeResponse,
				},
				Body: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
					AuthKey: "auth_key_user_a",
				},
			},
		},
		{
			name: "login with invalid username, expect OK login response (Cfg.DisableAuth=true)",
			cfg: server.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     1234,
				DisableAuth: true,
			},
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_b"),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: "sn_user_b",
							result:     nil,
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPChallengeResponse,
				},
				Body: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
					AuthKey: sessUUID.String(),
				},
			},
		},
		{
			name: "login with invalid username, expect failed login response (Cfg.DisableAuth=false)",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_b"),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: "sn_user_b",
							result:     nil,
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				Body: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
						},
					},
				},
			},
		},
		{
			name: "login fails on user manager lookup",
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_b"),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: "sn_user_b",
							err:        io.EOF,
						},
					},
				},
			},
			wantErr: io.EOF,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.getUserParams {
				userManager.EXPECT().
					GetUser(params.screenName).
					Return(params.result, params.err)
			}
			svc := AuthService{
				config:      tc.cfg,
				userManager: userManager,
			}
			fnNewUUID := func() uuid.UUID {
				return sessUUID
			}
			outputSNAC, err := svc.BUCPChallengeRequestHandler(tc.inputSNAC, fnNewUUID)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAuthService_RetrieveChatSession_HappyPath(t *testing.T) {
	chatID := "chat-1234"
	sess := newTestSession("screen-name", sessOptCannedID)

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		Retrieve(sess.ID()).
		Return(sess)

	chatRegistry := newMockChatRegistry(t)
	chatRegistry.EXPECT().
		Retrieve(chatID).
		Return(state.ChatRoom{}, sessionManager, nil)

	svc := NewAuthService(server.Config{}, nil, nil, nil, nil, chatRegistry)

	have, err := svc.RetrieveChatSession(chatID, sess.ID())
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RetrieveChatSession_ChatNotFound(t *testing.T) {
	chatID := "chat-1234"
	sess := newTestSession("screen-name", sessOptCannedID)

	chatRegistry := newMockChatRegistry(t)
	chatRegistry.EXPECT().
		Retrieve(chatID).
		Return(state.ChatRoom{}, nil, state.ErrChatRoomNotFound)

	svc := NewAuthService(server.Config{}, nil, nil, nil, nil, chatRegistry)

	_, err := svc.RetrieveChatSession(chatID, sess.ID())
	assert.ErrorIs(t, err, state.ErrChatRoomNotFound)
}

func TestAuthService_RetrieveChatSession_SessionNotFound(t *testing.T) {
	chatID := "chat-1234"
	sess := newTestSession("screen-name", sessOptCannedID)

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		Retrieve(sess.ID()).
		Return(nil)

	chatRegistry := newMockChatRegistry(t)
	chatRegistry.EXPECT().
		Retrieve(chatID).
		Return(state.ChatRoom{}, sessionManager, nil)

	svc := NewAuthService(server.Config{}, nil, nil, nil, nil, chatRegistry)

	have, err := svc.RetrieveChatSession(chatID, sess.ID())
	assert.NoError(t, err)
	assert.Nil(t, have)
}

func TestAuthService_RetrieveBOSSession_HappyPath(t *testing.T) {
	sess := newTestSession("screen-name", sessOptCannedID)

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		Retrieve(sess.ID()).
		Return(sess)

	svc := NewAuthService(server.Config{}, sessionManager, nil, nil, nil, nil)

	have, err := svc.RetrieveBOSSession(sess.ID())
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RetrieveBOSSession_SessionNotFound(t *testing.T) {
	sess := newTestSession("screen-name", sessOptCannedID)

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		Retrieve(sess.ID()).
		Return(nil)

	svc := NewAuthService(server.Config{}, sessionManager, nil, nil, nil, nil)

	have, err := svc.RetrieveBOSSession(sess.ID())
	assert.NoError(t, err)
	assert.Nil(t, have)
}

func TestAuthService_SignoutChat(t *testing.T) {
	sess := newTestSession("", sessOptCannedSignonTime)

	tests := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user signing out
		userSession *state.Session
		// chatRoom is the chat room user is exiting
		chatRoom state.ChatRoom
		// wantErr is the error we expect from the method
		wantErr error
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "user signs out of chat room, room is empty after user leaves",
			userSession: sess,
			chatRoom: state.ChatRoom{
				Cookie: "the-chat-cookie",
			},
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					broadcastExceptParams: broadcastExceptParams{
						{
							except: sess,
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Chat,
									SubGroup:  oscar.ChatUsersLeft,
								},
								Body: oscar.SNAC_0x0E_0x04_ChatUsersLeft{
									Users: []oscar.TLVUserInfo{
										sess.TLVUserInfo(),
									},
								},
							},
						},
					},
				},
				sessionManagerParams: sessionManagerParams{
					emptyParams: emptyParams{
						{
							result: true,
						},
					},
					removeParams: removeParams{
						{
							sess: sess,
						},
					},
				},
			},
		},
		{
			name:        "user signs out of chat room, room is not empty after user leaves",
			userSession: sess,
			chatRoom: state.ChatRoom{
				Cookie: "the-chat-cookie",
			},
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					broadcastExceptParams: broadcastExceptParams{
						{
							except: sess,
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Chat,
									SubGroup:  oscar.ChatUsersLeft,
								},
								Body: oscar.SNAC_0x0E_0x04_ChatUsersLeft{
									Users: []oscar.TLVUserInfo{
										sess.TLVUserInfo(),
									},
								},
							},
						},
					},
				},
				sessionManagerParams: sessionManagerParams{
					emptyParams: emptyParams{
						{
							result: false,
						},
					},
					removeParams: removeParams{
						{
							sess: sess,
						},
					},
				},
			},
		},
		{
			name:        "user can't sign out because chat room doesn't exist",
			userSession: sess,
			chatRoom: state.ChatRoom{
				Cookie: "the-chat-cookie",
			},
			wantErr: state.ErrChatRoomNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatMessageRelayer := newMockChatMessageRelayer(t)
			for _, params := range tt.mockParams.broadcastExceptParams {
				chatMessageRelayer.EXPECT().
					BroadcastExcept(nil, params.except, params.message)
			}

			sessionManager := newMockSessionManager(t)
			chatRegistry := newMockChatRegistry(t)
			for _, params := range tt.mockParams.removeParams {
				sessionManager.EXPECT().Remove(params.sess)
			}
			for _, params := range tt.mockParams.emptyParams {
				sessionManager.EXPECT().Empty().Return(params.result)
				if params.result {
					chatRegistry.EXPECT().Remove(tt.chatRoom.Cookie)
				}
			}
			chatSessionManager := struct {
				ChatMessageRelayer
				SessionManager
			}{
				chatMessageRelayer,
				sessionManager,
			}
			chatRegistry.EXPECT().
				Retrieve(tt.chatRoom.Cookie).
				Return(tt.chatRoom, chatSessionManager, tt.wantErr)

			svc := NewAuthService(server.Config{}, nil, nil, nil, nil, chatRegistry)

			err := svc.SignoutChat(nil, tt.userSession, tt.chatRoom.Cookie)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestAuthService_Signout(t *testing.T) {
	sess := newTestSession("user_screen_name", sessOptCannedSignonTime)

	tests := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user signing out
		userSession *state.Session
		// chatRoom is the chat room user is exiting
		chatRoom state.ChatRoom
		// wantErr is the error we expect from the method
		wantErr error
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "user signs out of chat room, room is empty after user leaves",
			userSession: sess,
			chatRoom: state.ChatRoom{
				Cookie: "the-chat-cookie",
			},
			mockParams: mockParams{
				sessionManagerParams: sessionManagerParams{
					removeParams: removeParams{
						{
							sess: sess,
						},
					},
				},
				feedbagManagerParams: feedbagManagerParams{
					interestedUsersParams: interestedUsersParams{
						{
							screenName: "user_screen_name",
							users:      []string{"friend1", "friend2"},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					broadcastToScreenNamesParams: broadcastToScreenNamesParams{
						{
							screenNames: []string{"friend1", "friend2"},
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyDeparted,
								},
								Body: oscar.SNAC_0x03_0x0C_BuddyDeparted{
									TLVUserInfo: oscar.TLVUserInfo{
										ScreenName:   sess.ScreenName(),
										WarningLevel: sess.Warning(),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "user signs out of chat room, room is empty after user leaves",
			userSession: sess,
			chatRoom: state.ChatRoom{
				Cookie: "the-chat-cookie",
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					interestedUsersParams: interestedUsersParams{
						{
							screenName: "user_screen_name",
							users:      []string{"friend1", "friend2"},
							err:        io.EOF,
						},
					},
				},
			},
			wantErr: io.EOF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.broadcastToScreenNamesParams {
				messageRelayer.EXPECT().
					BroadcastToScreenNames(mock.Anything, params.screenNames, params.message)
			}
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tt.mockParams.interestedUsersParams {
				feedbagManager.EXPECT().
					InterestedUsers(params.screenName).
					Return(params.users, params.err)
			}
			sessionManager := newMockSessionManager(t)
			for _, params := range tt.mockParams.removeParams {
				sessionManager.EXPECT().Remove(params.sess)
			}

			svc := NewAuthService(server.Config{}, sessionManager, messageRelayer, feedbagManager, nil, nil)

			err := svc.Signout(nil, tt.userSession)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
