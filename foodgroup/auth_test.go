package foodgroup

import (
	"bytes"
	"io"
	"testing"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_BUCPLoginRequest(t *testing.T) {
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
		cfg config.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNAC_0x17_0x02_BUCPLoginRequest
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// newUserFn is the function that registers a new user account
		newUserFn func(screenName string) (state.User, error)
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// wantErr is the error we expect from the method
		wantErr error
	}{
		{
			name: "user provides valid credentials and logs in successfully",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVPasswordHash, user.StrongMD5Pass),
						wire.NewTLV(wire.TLVScreenName, user.ScreenName),
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
					addSessionParams: addSessionParams{
						{
							sessID:     userSession.ID(),
							screenName: user.ScreenName,
							result:     userSession,
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginResponse,
				},
				Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.TLVScreenName, user.ScreenName),
							wire.NewTLV(wire.TLVReconnectHere, "127.0.0.1:1234"),
							wire.NewTLV(wire.TLVAuthorizationCookie, sessUUID.String()),
						},
					},
				},
			},
		},
		{
			name: "user logs in with non-existent screen name--account is created and logged in successfully",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     1234,
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVPasswordHash, user.StrongMD5Pass),
						wire.NewTLV(wire.TLVScreenName, user.ScreenName),
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
					addSessionParams: addSessionParams{
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
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginResponse,
				},
				Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.TLVScreenName, user.ScreenName),
							wire.NewTLV(wire.TLVReconnectHere, "127.0.0.1:1234"),
							wire.NewTLV(wire.TLVAuthorizationCookie, sessUUID.String()),
						},
					},
				},
			},
		},
		{
			name: "user logs in with invalid password--account is created and logged in successfully",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     1234,
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVPasswordHash, []byte("bad-password-hash")),
						wire.NewTLV(wire.TLVScreenName, user.ScreenName),
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
					addSessionParams: addSessionParams{
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
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginResponse,
				},
				Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.TLVScreenName, user.ScreenName),
							wire.NewTLV(wire.TLVReconnectHere, "127.0.0.1:1234"),
							wire.NewTLV(wire.TLVAuthorizationCookie, sessUUID.String()),
						},
					},
				},
			},
		},
		{
			name: "user provides invalid password--account creation fails due to user creation runtime error",
			cfg: config.Config{
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVPasswordHash, []byte("bad-password-hash")),
						wire.NewTLV(wire.TLVScreenName, user.ScreenName),
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
			cfg: config.Config{
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVPasswordHash, []byte("bad-password-hash")),
						wire.NewTLV(wire.TLVScreenName, user.ScreenName),
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
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVPasswordHash, []byte("bad_password")),
						wire.NewTLV(wire.TLVScreenName, user.ScreenName),
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
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginResponse,
				},
				Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.TLVScreenName, user.ScreenName),
							wire.NewTLV(wire.TLVErrorSubcode, uint16(0x01)),
						},
					},
				},
			},
		},
		{
			name: "login fails on user manager lookup",
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVPasswordHash, user.StrongMD5Pass),
						wire.NewTLV(wire.TLVScreenName, user.ScreenName),
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
					User(params.screenName).
					Return(params.result, params.err)
			}
			for _, params := range tc.mockParams.upsertUserParams {
				userManager.EXPECT().
					InsertUser(params.user).
					Return(params.err)
			}
			sessionManager := newMockSessionManager(t)
			for _, params := range tc.mockParams.addSessionParams {
				sessionManager.EXPECT().
					AddSession(params.sessID, params.screenName).
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
			outputSNAC, err := svc.BUCPLoginRequest(tc.inputSNAC, fnNewUUID, tc.newUserFn)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAuthService_BUCPChallengeRequest(t *testing.T) {
	sessUUID := uuid.UUID{1, 2, 3}
	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg config.Config
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNAC_0x17_0x06_BUCPChallengeRequest
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// wantErr is the error we expect from the method
		wantErr error
	}{
		{
			name: "login with valid username, expect OK login response",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			inputSNAC: wire.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVScreenName, "sn_user_a"),
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
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPChallengeResponse,
				},
				Body: wire.SNAC_0x17_0x07_BUCPChallengeResponse{
					AuthKey: "auth_key_user_a",
				},
			},
		},
		{
			name: "login with invalid username, expect OK login response (Cfg.DisableAuth=true)",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     1234,
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVScreenName, "sn_user_b"),
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
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPChallengeResponse,
				},
				Body: wire.SNAC_0x17_0x07_BUCPChallengeResponse{
					AuthKey: sessUUID.String(),
				},
			},
		},
		{
			name: "login with invalid username, expect failed login response (Cfg.DisableAuth=false)",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			inputSNAC: wire.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVScreenName, "sn_user_b"),
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
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginResponse,
				},
				Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.TLVErrorSubcode, uint16(0x01)),
						},
					},
				},
			},
		},
		{
			name: "login fails on user manager lookup",
			inputSNAC: wire.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.TLVScreenName, "sn_user_b"),
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
					User(params.screenName).
					Return(params.result, params.err)
			}
			svc := AuthService{
				config:      tc.cfg,
				userManager: userManager,
			}
			fnNewUUID := func() uuid.UUID {
				return sessUUID
			}
			outputSNAC, err := svc.BUCPChallengeRequest(tc.inputSNAC, fnNewUUID)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAuthService_RetrieveChatSession_HappyPath(t *testing.T) {
	cookie := "chat-1234"
	sess := newTestSession("screen-name", sessOptCannedID)

	c := chatLoginCookie{
		Cookie: cookie,
		SessID: sess.ID(),
	}
	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(c, buf))

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		RetrieveSession(sess.ID()).
		Return(sess)

	chatRegistry := newMockChatRegistry(t)
	chatRegistry.EXPECT().
		Retrieve(cookie).
		Return(state.ChatRoom{}, sessionManager, nil)

	svc := NewAuthService(config.Config{}, nil, nil, nil, nil, chatRegistry, nil)

	have, err := svc.RetrieveChatSession(buf.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RetrieveChatSession_ChatNotFound(t *testing.T) {
	cookie := "chat-1234"
	sess := newTestSession("screen-name", sessOptCannedID)

	c := chatLoginCookie{
		Cookie: cookie,
		SessID: sess.ID(),
	}
	loginCookie := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(c, loginCookie))

	chatRegistry := newMockChatRegistry(t)
	chatRegistry.EXPECT().
		Retrieve(cookie).
		Return(state.ChatRoom{}, nil, state.ErrChatRoomNotFound)

	svc := NewAuthService(config.Config{}, nil, nil, nil, nil, chatRegistry, nil)

	_, err := svc.RetrieveChatSession(loginCookie.Bytes())
	assert.ErrorIs(t, err, state.ErrChatRoomNotFound)
}

func TestAuthService_RetrieveChatSession_SessionNotFound(t *testing.T) {
	cookie := "chat-1234"
	sess := newTestSession("screen-name", sessOptCannedID)

	c := chatLoginCookie{
		Cookie: cookie,
		SessID: sess.ID(),
	}
	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(c, buf))

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		RetrieveSession(sess.ID()).
		Return(nil)

	chatRegistry := newMockChatRegistry(t)
	chatRegistry.EXPECT().
		Retrieve(cookie).
		Return(state.ChatRoom{}, sessionManager, nil)

	svc := NewAuthService(config.Config{}, nil, nil, nil, nil, chatRegistry, nil)

	have, err := svc.RetrieveChatSession(buf.Bytes())
	assert.NoError(t, err)
	assert.Nil(t, have)
}

func TestAuthService_RetrieveBOSSession_HappyPath(t *testing.T) {
	sess := newTestSession("screen-name", sessOptCannedID)

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		RetrieveSession(sess.ID()).
		Return(sess)

	svc := NewAuthService(config.Config{}, sessionManager, nil, nil, nil, nil, nil)

	have, err := svc.RetrieveBOSSession(sess.ID())
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RetrieveBOSSession_SessionNotFound(t *testing.T) {
	sess := newTestSession("screen-name", sessOptCannedID)

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		RetrieveSession(sess.ID()).
		Return(nil)

	svc := NewAuthService(config.Config{}, sessionManager, nil, nil, nil, nil, nil)

	have, err := svc.RetrieveBOSSession(sess.ID())
	assert.NoError(t, err)
	assert.Nil(t, have)
}

func TestAuthService_SignoutChat(t *testing.T) {
	sess := newTestSession("", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie"))

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
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatUsersLeft,
								},
								Body: wire.SNAC_0x0E_0x04_ChatUsersLeft{
									Users: []wire.TLVUserInfo{
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
					removeSessionParams: removeSessionParams{
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
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatUsersLeft,
								},
								Body: wire.SNAC_0x0E_0x04_ChatUsersLeft{
									Users: []wire.TLVUserInfo{
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
					removeSessionParams: removeSessionParams{
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
					RelayToAllExcept(nil, params.except, params.message)
			}

			sessionManager := newMockSessionManager(t)
			chatRegistry := newMockChatRegistry(t)
			for _, params := range tt.mockParams.removeSessionParams {
				sessionManager.EXPECT().RemoveSession(params.sess)
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

			svc := NewAuthService(config.Config{}, nil, nil, nil, nil, chatRegistry, nil)

			err := svc.SignoutChat(nil, tt.userSession)
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
					removeSessionParams: removeSessionParams{
						{
							sess: sess,
						},
					},
				},
				feedbagManagerParams: feedbagManagerParams{
					adjacentUsersParams: adjacentUsersParams{
						{
							screenName: "user_screen_name",
							users:      []string{"friend1", "friend2"},
						},
					},
				},
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					deleteUserParams: deleteUserParams{
						{
							userScreenName: "user_screen_name",
						},
					},
					whoAddedUserParams: whoAddedUserParams{
						{
							userScreenName: "user_screen_name",
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNamesParams: relayToScreenNamesParams{
						{
							screenNames: []string{"friend1", "friend2"},
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyDeparted,
								},
								Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
									TLVUserInfo: wire.TLVUserInfo{
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
			name:        "user signs out of chat room, feedbag lookup returns error",
			userSession: sess,
			chatRoom: state.ChatRoom{
				Cookie: "the-chat-cookie",
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					adjacentUsersParams: adjacentUsersParams{
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
			for _, params := range tt.mockParams.relayToScreenNamesParams {
				messageRelayer.EXPECT().
					RelayToScreenNames(mock.Anything, params.screenNames, params.message)
			}
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tt.mockParams.adjacentUsersParams {
				feedbagManager.EXPECT().
					AdjacentUsers(params.screenName).
					Return(params.users, params.err)
			}
			sessionManager := newMockSessionManager(t)
			for _, params := range tt.mockParams.removeSessionParams {
				sessionManager.EXPECT().RemoveSession(params.sess)
			}
			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tt.mockParams.deleteUserParams {
				legacyBuddyListManager.EXPECT().DeleteUser(params.userScreenName)
			}
			for _, params := range tt.mockParams.whoAddedUserParams {
				legacyBuddyListManager.EXPECT().
					WhoAddedUser(params.userScreenName).
					Return(params.result)
			}

			svc := NewAuthService(config.Config{}, sessionManager, messageRelayer, feedbagManager, nil, nil, legacyBuddyListManager)

			err := svc.Signout(nil, tt.userSession)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
