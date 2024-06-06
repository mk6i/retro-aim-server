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
	user := state.User{
		ScreenName: "screen_name",
		AuthKey:    "auth_key",
	}
	assert.NoError(t, user.HashPassword("the_password"))

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
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, user.StrongMD5Pass),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.ScreenName),
						cookie: []byte("the-cookie"),
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
							wire.NewTLV(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
						},
					},
				},
			},
		},
		{
			name: "user logs in with non-existent screen name--account is created and logged in successfully",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, user.StrongMD5Pass),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
					insertUserParams: insertUserParams{
						{
							user: user,
						},
					},
				},
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.ScreenName),
						cookie: []byte("the-cookie"),
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
							wire.NewTLV(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
						},
					},
				},
			},
		},
		{
			name: "user logs in with invalid password--account is created and logged in successfully",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-password-hash")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
					insertUserParams: insertUserParams{
						{
							user: user,
						},
					},
				},
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.ScreenName),
						cookie: []byte("the-cookie"),
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
							wire.NewTLV(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
						},
					},
				},
			},
		},
		{
			name: "user logs in with invalid password--account already exists and logged in successfully",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-password-hash")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
					insertUserParams: insertUserParams{
						{
							user: user,
							err:  state.ErrDupUser,
						},
					},
				},
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.ScreenName),
						cookie: []byte("the-cookie"),
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
							wire.NewTLV(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
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
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-password-hash")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-password-hash")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
					insertUserParams: insertUserParams{
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
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad_password")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
							wire.NewTLV(wire.LoginTLVTagsErrorSubcode, uint16(0x01)),
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
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, user.StrongMD5Pass),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
			for _, params := range tc.mockParams.insertUserParams {
				userManager.EXPECT().
					InsertUser(params.user).
					Return(params.err)
			}
			sessionManager := newMockSessionManager(t)
			cookieIssuer := newMockCookieIssuer(t)
			for _, params := range tc.mockParams.cookieIssuerParams {
				cookieIssuer.EXPECT().
					Issue(params.data).
					Return(params.cookie, params.err)
			}

			svc := AuthService{
				config:         tc.cfg,
				cookieIssuer:   cookieIssuer,
				sessionManager: sessionManager,
				userManager:    userManager,
			}
			outputSNAC, err := svc.BUCPLogin(tc.inputSNAC, tc.newUserFn)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAuthService_FLAPLoginResponse(t *testing.T) {
	user := state.User{
		ScreenName: "screen_name",
		AuthKey:    "auth_key",
	}
	assert.NoError(t, user.HashPassword("the_password"))

	// obfuscated password value: "the_password"
	roastedPassword := []byte{0x87, 0x4E, 0xE4, 0x9B, 0x49, 0xE7, 0xA8, 0xE1, 0x06, 0xCC, 0xCB, 0x82}

	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the app configuration
		cfg config.Config
		// inputSNAC is the authentication FLAP frame sent from the client to the server
		inputSNAC wire.FLAPSignonFrame
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// newUserFn is the function that registers a new user account
		newUserFn func(screenName string) (state.User, error)
		// expectOutput is the response sent from the server to client
		expectOutput wire.TLVRestBlock
		// wantErr is the error we expect from the method
		wantErr error
	}{
		{
			name: "user provides valid credentials and logs in successfully",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsRoastedPassword, roastedPassword),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.ScreenName),
						cookie: []byte("the-cookie"),
					},
				},
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
					wire.NewTLV(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
					wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
				},
			},
		},
		{
			name: "user logs in with non-existent screen name--account is created and logged in successfully",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsRoastedPassword, roastedPassword),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
					insertUserParams: insertUserParams{
						{
							user: user,
						},
					},
				},
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.ScreenName),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName string) (state.User, error) {
				return user, nil
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
					wire.NewTLV(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
					wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
				},
			},
		},
		{
			name: "user logs in with invalid password--account is created and logged in successfully",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-roasted-password")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
					insertUserParams: insertUserParams{
						{
							user: user,
						},
					},
				},
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.ScreenName),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName string) (state.User, error) {
				return user, nil
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
					wire.NewTLV(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
					wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
				},
			},
		},
		{
			name: "user logs in with invalid password--account already exists and logged in successfully",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-roasted-password")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
					insertUserParams: insertUserParams{
						{
							user: user,
							err:  state.ErrDupUser,
						},
					},
				},
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.ScreenName),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName string) (state.User, error) {
				return user, nil
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
					wire.NewTLV(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
					wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
				},
			},
		},
		{
			name: "user provides invalid password--account creation fails due to user creation runtime error",
			cfg: config.Config{
				DisableAuth: true,
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-roasted-password")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-roasted-password")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
					insertUserParams: insertUserParams{
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
				BOSPort:   "1234",
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, []byte("bad-roasted-password")),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
					wire.NewTLV(wire.LoginTLVTagsErrorSubcode, uint16(0x01)),
				},
			},
		},
		{
			name: "login fails on user manager lookup",
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsPasswordHash, user.StrongMD5Pass),
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.ScreenName),
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
			for _, params := range tc.mockParams.insertUserParams {
				userManager.EXPECT().
					InsertUser(params.user).
					Return(params.err)
			}
			sessionManager := newMockSessionManager(t)
			cookieIssuer := newMockCookieIssuer(t)
			for _, params := range tc.mockParams.cookieIssuerParams {
				cookieIssuer.EXPECT().
					Issue(params.data).
					Return(params.cookie, params.err)
			}
			svc := AuthService{
				config:         tc.cfg,
				cookieIssuer:   cookieIssuer,
				sessionManager: sessionManager,
				userManager:    userManager,
			}
			outputSNAC, err := svc.FLAPLogin(tc.inputSNAC, tc.newUserFn)
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
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsScreenName, "sn_user_a"),
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
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsScreenName, "sn_user_b"),
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
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(wire.LoginTLVTagsScreenName, "sn_user_b"),
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
							wire.NewTLV(wire.LoginTLVTagsErrorSubcode, uint16(0x01)),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, "sn_user_b"),
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
			outputSNAC, err := svc.BUCPChallenge(tc.inputSNAC, fnNewUUID)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAuthService_RegisterChatSession_HappyPath(t *testing.T) {
	cookie := "chat-1234"
	sess := newTestSession("screen-name")

	c := chatLoginCookie{
		ChatCookie: cookie,
		ScreenName: sess.ScreenName(),
	}
	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(c, buf))

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		AddSession(sess.ScreenName()).
		Return(sess)

	chatRegistry := newMockChatRegistry(t)
	chatRegistry.EXPECT().
		Retrieve(cookie).
		Return(state.ChatRoom{}, sessionManager, nil)

	cookieIssuer := newMockCookieIssuer(t)

	svc := NewAuthService(config.Config{}, nil, nil, chatRegistry, nil, cookieIssuer, nil)

	have, err := svc.RegisterChatSession(buf.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RegisterBOSSession_ChatNotFound(t *testing.T) {
	cookie := "chat-1234"
	sess := newTestSession("screen-name")

	c := chatLoginCookie{
		ChatCookie: cookie,
		ScreenName: sess.ScreenName(),
	}
	loginCookie := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(c, loginCookie))

	chatRegistry := newMockChatRegistry(t)
	chatRegistry.EXPECT().
		Retrieve(cookie).
		Return(state.ChatRoom{}, nil, state.ErrChatRoomNotFound)

	cookieIssuer := newMockCookieIssuer(t)
	svc := NewAuthService(config.Config{}, nil, nil, chatRegistry, nil, cookieIssuer, nil)

	_, err := svc.RegisterChatSession(loginCookie.Bytes())
	assert.ErrorIs(t, err, state.ErrChatRoomNotFound)
}

func TestAuthService_RegisterBOSSession_HappyPath(t *testing.T) {
	sess := newTestSession("screen-name")

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		AddSession(sess.ScreenName()).
		Return(sess)

	cookieIssuer := newMockCookieIssuer(t)

	svc := NewAuthService(config.Config{}, sessionManager, nil, nil, nil, cookieIssuer, nil)

	have, err := svc.RegisterBOSSession(sess.ScreenName())
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RegisterBOSSession_SessionNotFound(t *testing.T) {
	sess := newTestSession("screen-name")

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		AddSession(sess.ScreenName()).
		Return(nil)

	cookieIssuer := newMockCookieIssuer(t)

	svc := NewAuthService(config.Config{}, sessionManager, nil, nil, nil, cookieIssuer, nil)

	have, err := svc.RegisterBOSSession(sess.ScreenName())
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

			cookieIssuer := newMockCookieIssuer(t)

			svc := NewAuthService(config.Config{}, nil, nil, chatRegistry, nil, cookieIssuer, nil)

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
		// wantErr is the error we expect from the method
		wantErr error
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "user signs out of chat room, room is empty after user leaves",
			userSession: sess,
			mockParams: mockParams{
				sessionManagerParams: sessionManagerParams{
					removeSessionParams: removeSessionParams{
						{
							sess: sess,
						},
					},
				},
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					deleteUserParams: deleteUserParams{
						{
							userScreenName: "user_screen_name",
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyDepartedParams: broadcastBuddyDepartedParams{
						{
							screenName: "user_screen_name",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionManager := newMockSessionManager(t)
			for _, params := range tt.mockParams.removeSessionParams {
				sessionManager.EXPECT().RemoveSession(params.sess)
			}
			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tt.mockParams.deleteUserParams {
				legacyBuddyListManager.EXPECT().DeleteUser(params.userScreenName)
			}
			buddyUpdateBroadcaster := newMockBuddyBroadcaster(t)
			for _, params := range tt.mockParams.broadcastBuddyDepartedParams {
				p := params
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyDeparted(mock.Anything, mock.MatchedBy(func(s *state.Session) bool {
						return s.ScreenName() == p.screenName
					})).
					Return(nil)
			}
			svc := NewAuthService(config.Config{}, sessionManager, nil, nil, legacyBuddyListManager, nil, buddyUpdateBroadcaster)

			err := svc.Signout(nil, tt.userSession)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
