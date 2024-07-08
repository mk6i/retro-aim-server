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
		IdentScreenName: state.NewIdentScreenName("screen_name"),
		AuthKey:         "auth_key",
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
		newUserFn func(screenName state.DisplayScreenName) (state.User, error)
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
							result:     &user,
						},
					},
				},
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.IdentScreenName.String()),
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
						data:   []byte(user.IdentScreenName.String()),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
						data:   []byte(user.IdentScreenName.String()),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
						data:   []byte(user.IdentScreenName.String()),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
							result:     &user,
						},
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
							wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
			cookieBaker := newMockCookieBaker(t)
			for _, params := range tc.mockParams.cookieIssuerParams {
				cookieBaker.EXPECT().
					Issue(params.data).
					Return(params.cookie, params.err)
			}

			svc := AuthService{
				config:         tc.cfg,
				cookieBaker:    cookieBaker,
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
		IdentScreenName: state.NewIdentScreenName("screen_name"),
		AuthKey:         "auth_key",
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
		newUserFn func(screenName state.DisplayScreenName) (state.User, error)
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
							result:     &user,
						},
					},
				},
				cookieIssuerParams: cookieIssuerParams{
					{
						data:   []byte(user.IdentScreenName.String()),
						cookie: []byte("the-cookie"),
					},
				},
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
						data:   []byte(user.IdentScreenName.String()),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
				return user, nil
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
						data:   []byte(user.IdentScreenName.String()),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
				return user, nil
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
						data:   []byte(user.IdentScreenName.String()),
						cookie: []byte("the-cookie"),
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
				return user, nil
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
							result:     &user,
						},
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
							result:     nil,
						},
					},
				},
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
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
						wire.NewTLV(wire.LoginTLVTagsScreenName, user.IdentScreenName),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: user.IdentScreenName,
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
			cookieBaker := newMockCookieBaker(t)
			for _, params := range tc.mockParams.cookieIssuerParams {
				cookieBaker.EXPECT().
					Issue(params.data).
					Return(params.cookie, params.err)
			}
			svc := AuthService{
				config:         tc.cfg,
				cookieBaker:    cookieBaker,
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
							screenName: state.NewIdentScreenName("sn_user_a"),
							result: &state.User{
								IdentScreenName: state.NewIdentScreenName("sn_user_a"),
								AuthKey:         "auth_key_user_a",
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
							screenName: state.NewIdentScreenName("sn_user_b"),
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
							screenName: state.NewIdentScreenName("sn_user_b"),
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
							screenName: state.NewIdentScreenName("sn_user_b"),
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
	sess := newTestSession("ScreenName")

	chatCookie := "the-chat-cookie"
	chatSessionRegistry := newMockChatSessionRegistry(t)
	chatSessionRegistry.EXPECT().
		AddSession(chatCookie, sess.DisplayScreenName()).
		Return(sess)

	c := chatLoginCookie{
		ChatCookie: chatCookie,
		ScreenName: sess.DisplayScreenName(),
	}
	chatCookieBuf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(c, chatCookieBuf))

	authCookie := []byte("the-auth-cookie")
	cookieBaker := newMockCookieBaker(t)
	cookieBaker.EXPECT().
		Crack(authCookie).
		Return(chatCookieBuf.Bytes(), nil)

	svc := NewAuthService(config.Config{}, nil, chatSessionRegistry, nil, nil, cookieBaker, nil, nil, nil)

	have, err := svc.RegisterChatSession(authCookie)
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RegisterBOSSession_HappyPath(t *testing.T) {
	sess := newTestSession("screen-name")

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		AddSession(sess.DisplayScreenName()).
		Return(sess)

	authCookie := []byte(`the-auth-cookie`)

	cookieBaker := newMockCookieBaker(t)
	cookieBaker.EXPECT().
		Crack(authCookie).
		Return([]byte("screen-name"), nil)

	userManager := newMockUserManager(t)
	userManager.EXPECT().
		User(sess.IdentScreenName()).
		Return(&state.User{DisplayScreenName: sess.DisplayScreenName()}, nil)

	svc := NewAuthService(config.Config{}, sessionManager, nil, userManager, nil, cookieBaker, nil, nil, nil)

	have, err := svc.RegisterBOSSession(authCookie)
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RetrieveBOSSession_HappyPath(t *testing.T) {
	sess := newTestSession("screen-name")

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		RetrieveSession(sess.IdentScreenName()).
		Return(sess)

	authCookie := []byte(`the-auth-cookie`)

	cookieBaker := newMockCookieBaker(t)
	cookieBaker.EXPECT().
		Crack(authCookie).
		Return([]byte("screen-name"), nil)

	userManager := newMockUserManager(t)
	userManager.EXPECT().
		User(sess.IdentScreenName()).
		Return(&state.User{IdentScreenName: sess.IdentScreenName()}, nil)

	svc := NewAuthService(config.Config{}, sessionManager, nil, userManager, nil, cookieBaker, nil, nil, nil)

	have, err := svc.RetrieveBOSSession(authCookie)
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RetrieveBOSSession_SessionNotFound(t *testing.T) {
	sess := newTestSession("screen-name")

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().
		RetrieveSession(sess.IdentScreenName()).
		Return(nil)

	authCookie := []byte(`the-auth-cookie`)
	cookieBaker := newMockCookieBaker(t)

	cookieBaker.EXPECT().
		Crack(authCookie).
		Return([]byte("screen-name"), nil)

	userManager := newMockUserManager(t)
	userManager.EXPECT().
		User(sess.IdentScreenName()).
		Return(&state.User{IdentScreenName: sess.IdentScreenName()}, nil)

	svc := NewAuthService(config.Config{}, sessionManager, nil, userManager, nil, cookieBaker, nil, nil, nil)

	have, err := svc.RetrieveBOSSession(authCookie)
	assert.NoError(t, err)
	assert.Nil(t, have)
}

func TestAuthService_SignoutChat(t *testing.T) {
	tests := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user signing out
		userSession *state.Session
		// chatRoom is the chat room user is exiting
		chatRoom state.ChatRoom
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "user signs out of chat room, room is empty after user leaves",
			userSession: newTestSession("the-screen-name", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie")),
			chatRoom: state.ChatRoom{
				Cookie: "the-chat-cookie",
			},
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					chatRelayToAllExceptParams: chatRelayToAllExceptParams{
						{
							screenName: state.NewIdentScreenName("the-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatUsersLeft,
								},
								Body: wire.SNAC_0x0E_0x04_ChatUsersLeft{
									Users: []wire.TLVUserInfo{
										newTestSession("the-screen-name", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie")).TLVUserInfo(),
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
							screenName: state.NewIdentScreenName("the-screen-name"),
						},
					},
				},
			},
		},
		{
			name:        "user signs out of chat room, room is not empty after user leaves",
			userSession: newTestSession("the-screen-name", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie")),
			chatRoom: state.ChatRoom{
				Cookie: "the-chat-cookie",
			},
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					chatRelayToAllExceptParams: chatRelayToAllExceptParams{
						{
							screenName: state.NewIdentScreenName("the-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatUsersLeft,
								},
								Body: wire.SNAC_0x0E_0x04_ChatUsersLeft{
									Users: []wire.TLVUserInfo{
										newTestSession("the-screen-name", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie")).TLVUserInfo(),
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
							screenName: state.NewIdentScreenName("the-screen-name"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatMessageRelayer := newMockChatMessageRelayer(t)
			for _, params := range tt.mockParams.chatRelayToAllExceptParams {
				chatMessageRelayer.EXPECT().
					RelayToAllExcept(nil, tt.chatRoom.Cookie, params.screenName, params.message)
			}
			sessionManager := newMockChatSessionRegistry(t)
			for _, params := range tt.mockParams.removeSessionParams {
				sessionManager.EXPECT().
					RemoveSession(matchSession(params.screenName))
			}

			svc := NewAuthService(config.Config{}, nil, sessionManager, nil, nil, nil, nil, nil, chatMessageRelayer)
			svc.SignoutChat(nil, tt.userSession)
		})
	}
}

func TestAuthService_Signout(t *testing.T) {
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
			userSession: newTestSession("user_screen_name", sessOptCannedSignonTime),
			mockParams: mockParams{
				sessionManagerParams: sessionManagerParams{
					removeSessionParams: removeSessionParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
						},
					},
				},
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					deleteUserParams: deleteUserParams{
						{
							userScreenName: state.NewIdentScreenName("user_screen_name"),
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyDepartedParams: broadcastBuddyDepartedParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
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
				sessionManager.EXPECT().RemoveSession(matchSession(params.screenName))
			}
			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tt.mockParams.deleteUserParams {
				legacyBuddyListManager.EXPECT().DeleteUser(params.userScreenName)
			}
			buddyUpdateBroadcaster := newMockbuddyBroadcaster(t)
			for _, params := range tt.mockParams.broadcastBuddyDepartedParams {
				p := params
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyDeparted(mock.Anything, mock.MatchedBy(func(s *state.Session) bool {
						return s.IdentScreenName() == p.screenName
					})).
					Return(nil)
			}
			svc := NewAuthService(config.Config{}, sessionManager, nil, nil, legacyBuddyListManager, nil, nil, nil, nil)
			svc.buddyUpdateBroadcaster = buddyUpdateBroadcaster

			err := svc.Signout(nil, tt.userSession)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
