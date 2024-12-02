package foodgroup

import (
	"bytes"
	"fmt"
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
		IdentScreenName:   state.NewIdentScreenName("screenName"),
		DisplayScreenName: "screenName",
		AuthKey:           "auth_key",
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
			name: "AIM account exists, correct password, login OK",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
						wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, user.StrongMD5Pass),
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
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: func() []byte {
								loginCookie := bosCookie{
									ScreenName: user.DisplayScreenName,
								}
								buf := &bytes.Buffer{}
								assert.NoError(t, wire.MarshalBE(loginCookie, buf))
								return buf.Bytes()
							}(),
							cookieOut: []byte("the-cookie"),
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
							wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
							wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
						},
					},
				},
			},
		},
		{
			name: "ICQ account exists, correct password, login OK",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
						wire.NewTLVBE(wire.LoginTLVTagsClientIdentity, "ICQ 2000b"),
						wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, user.StrongMD5Pass),
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
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: func() []byte {
								loginCookie := bosCookie{
									ScreenName: user.DisplayScreenName,
									ClientID:   "ICQ 2000b",
								}
								buf := &bytes.Buffer{}
								assert.NoError(t, wire.MarshalBE(loginCookie, buf))
								return buf.Bytes()
							}(),
							cookieOut: []byte("the-cookie"),
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
							wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
							wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
						},
					},
				},
			},
		},
		{
			name: "AIM account exists, incorrect password, login fails",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
						wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, []byte("bad_password")),
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
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginResponse,
				},
				Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: []wire.TLV{
							wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
							wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, wire.LoginErrInvalidPassword),
						},
					},
				},
			},
		},
		{
			name: "AIM account doesn't exist, login fails",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, []byte("password")),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, []byte("non_existent_screen_name")),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("non_existent_screen_name"),
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
						TLVList: []wire.TLV{
							wire.NewTLVBE(wire.LoginTLVTagsScreenName, state.NewIdentScreenName("non_existent_screen_name")),
							wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, wire.LoginErrInvalidUsernameOrPassword),
						},
					},
				},
			},
		},
		{
			name: "ICQ account doesn't exist, login fails",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, []byte("password")),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, []byte("100003")),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("100003"),
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
						TLVList: []wire.TLV{
							wire.NewTLVBE(wire.LoginTLVTagsScreenName, state.NewIdentScreenName("100003")),
							wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, wire.LoginErrICQUserErr),
						},
					},
				},
			},
		},
		{
			name: "account doesn't exist, authentication is disabled, account is created, login succeeds",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
						wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, user.StrongMD5Pass),
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
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: func() []byte {
								loginCookie := bosCookie{
									ScreenName: user.DisplayScreenName,
								}
								buf := &bytes.Buffer{}
								assert.NoError(t, wire.MarshalBE(loginCookie, buf))
								return buf.Bytes()
							}(),
							cookieOut: []byte("the-cookie"),
						},
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
							wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
							wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
						},
					},
				},
			},
		},
		{
			name: "account exists, password is invalid, authentication is disabled, login succeeds",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
						wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, []byte("bad-password-hash")),
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
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: func() []byte {
								loginCookie := bosCookie{
									ScreenName: user.DisplayScreenName,
								}
								buf := &bytes.Buffer{}
								assert.NoError(t, wire.MarshalBE(loginCookie, buf))
								return buf.Bytes()
							}(),
							cookieOut: []byte("the-cookie"),
						},
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
							wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
							wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
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
						wire.NewTLVBE(wire.LoginTLVTagsPasswordHash, user.StrongMD5Pass),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
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
			for _, params := range tc.mockParams.userManagerParams.getUserParams {
				userManager.EXPECT().
					User(params.screenName).
					Return(params.result, params.err)
			}
			for _, params := range tc.mockParams.insertUserParams {
				userManager.EXPECT().
					InsertUser(params.user).
					Return(params.err)
			}
			cookieBaker := newMockCookieBaker(t)
			for _, params := range tc.mockParams.cookieIssueParams {
				cookieBaker.EXPECT().
					Issue(params.dataIn).
					Return(params.cookieOut, params.err)
			}

			svc := AuthService{
				config:      tc.cfg,
				cookieBaker: cookieBaker,
				userManager: userManager,
			}
			outputSNAC, err := svc.BUCPLogin(tc.inputSNAC, tc.newUserFn)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestAuthService_FLAPLoginResponse(t *testing.T) {
	user := state.User{
		AuthKey:           "auth_key",
		DisplayScreenName: "screenName",
		IdentScreenName:   state.NewIdentScreenName("screenName"),
	}
	assert.NoError(t, user.HashPassword("the_password"))

	// roastedPassword the roasted form of "the_password"
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
			name: "AIM account exists, correct password, login OK",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, roastedPassword),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
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
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: func() []byte {
								loginCookie := bosCookie{
									ScreenName: user.DisplayScreenName,
								}
								buf := &bytes.Buffer{}
								assert.NoError(t, wire.MarshalBE(loginCookie, buf))
								return buf.Bytes()
							}(),
							cookieOut: []byte("the-cookie"),
						},
					},
				},
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
					wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
					wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
				},
			},
		},
		{
			name: "ICQ account exists, correct password, login OK",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsClientIdentity, "ICQ 2000b"),
						wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, roastedPassword),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
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
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: func() []byte {
								loginCookie := bosCookie{
									ScreenName: user.DisplayScreenName,
									ClientID:   "ICQ 2000b",
								}
								buf := &bytes.Buffer{}
								assert.NoError(t, wire.MarshalBE(loginCookie, buf))
								return buf.Bytes()
							}(),
							cookieOut: []byte("the-cookie"),
						},
					},
				},
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
					wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
					wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
				},
			},
		},
		{
			name: "AIM account exists, incorrect password, login fails",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, []byte("bad_roasted_password")),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
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
			expectOutput: wire.TLVRestBlock{
				TLVList: []wire.TLV{
					wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
					wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, wire.LoginErrInvalidPassword),
				},
			},
		},
		{
			name: "AIM account doesn't exist, login fails",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, roastedPassword),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, []byte("non_existent_screen_name")),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("non_existent_screen_name"),
							result:     nil,
						},
					},
				},
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: []wire.TLV{
					wire.NewTLVBE(wire.LoginTLVTagsScreenName, state.NewIdentScreenName("non_existent_screen_name")),
					wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, wire.LoginErrInvalidUsernameOrPassword),
				},
			},
		},
		{
			name: "ICQ account doesn't exist, login fails",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   "1234",
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsClientIdentity, "ICQ 2000b"),
						wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, roastedPassword),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, []byte("100003")),
					},
				},
			},
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("100003"),
							result:     nil,
						},
					},
				},
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: []wire.TLV{
					wire.NewTLVBE(wire.LoginTLVTagsScreenName, state.NewIdentScreenName("100003")),
					wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, wire.LoginErrICQUserErr),
				},
			},
		},
		{
			name: "account doesn't exist, authentication is disabled, account is created, login succeeds",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, roastedPassword),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
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
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: func() []byte {
								loginCookie := bosCookie{
									ScreenName: user.DisplayScreenName,
								}
								buf := &bytes.Buffer{}
								assert.NoError(t, wire.MarshalBE(loginCookie, buf))
								return buf.Bytes()
							}(),
							cookieOut: []byte("the-cookie"),
						},
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
				return user, nil
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
					wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
					wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
				},
			},
		},
		{
			name: "account exists, password is invalid, authentication is disabled, login succeeds",
			cfg: config.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     "1234",
				DisableAuth: true,
			},
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, "bad-roasted-password"),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
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
				cookieBakerParams: cookieBakerParams{
					cookieIssueParams: cookieIssueParams{
						{
							dataIn: func() []byte {
								loginCookie := bosCookie{
									ScreenName: user.DisplayScreenName,
								}
								buf := &bytes.Buffer{}
								assert.NoError(t, wire.MarshalBE(loginCookie, buf))
								return buf.Bytes()
							}(),
							cookieOut: []byte("the-cookie"),
						},
					},
				},
			},
			newUserFn: func(screenName state.DisplayScreenName) (state.User, error) {
				return user, nil
			},
			expectOutput: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
					wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, "127.0.0.1:1234"),
					wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, []byte("the-cookie")),
				},
			},
		},
		{
			name: "login fails on user manager lookup",
			inputSNAC: wire.FLAPSignonFrame{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, roastedPassword),
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, user.DisplayScreenName),
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
			for _, params := range tc.mockParams.userManagerParams.getUserParams {
				userManager.EXPECT().
					User(params.screenName).
					Return(params.result, params.err)
			}
			for _, params := range tc.mockParams.insertUserParams {
				userManager.EXPECT().
					InsertUser(params.user).
					Return(params.err)
			}
			cookieBaker := newMockCookieBaker(t)
			for _, params := range tc.mockParams.cookieIssueParams {
				cookieBaker.EXPECT().
					Issue(params.dataIn).
					Return(params.cookieOut, params.err)
			}
			svc := AuthService{
				config:      tc.cfg,
				cookieBaker: cookieBaker,
				userManager: userManager,
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
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, "sn_user_a"),
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
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, "sn_user_b"),
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
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, "sn_user_b"),
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
							wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, uint16(0x01)),
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
						wire.NewTLVBE(wire.LoginTLVTagsScreenName, "sn_user_b"),
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
			for _, params := range tc.mockParams.userManagerParams.getUserParams {
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
	assert.NoError(t, wire.MarshalBE(c, chatCookieBuf))

	authCookie := []byte("the-auth-cookie")
	cookieBaker := newMockCookieBaker(t)
	cookieBaker.EXPECT().
		Crack(authCookie).
		Return(chatCookieBuf.Bytes(), nil)

	svc := NewAuthService(config.Config{}, nil, chatSessionRegistry, nil, cookieBaker, nil, nil, nil, nil, nil)

	have, err := svc.RegisterChatSession(authCookie)
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RegisterBOSSession(t *testing.T) {
	screenName := state.DisplayScreenName("UserScreenName")
	aimAuthCookie := bosCookie{
		ScreenName: screenName,
	}
	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(aimAuthCookie, buf))
	aimCookie := buf.Bytes()

	uin := state.DisplayScreenName("100003")
	icqAuthCookie := bosCookie{
		ScreenName: uin,
	}
	buf = &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(icqAuthCookie, buf))
	icqCookie := buf.Bytes()

	cases := []struct {
		// name is the unit test name
		name string
		// cookieOut is the auth cookieOut that contains session information
		cookie []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// wantSess asserts the values of one or more session properties
		wantSess func(*state.Session) bool
		// wantErr is the error we expect from the method
		wantErr error
	}{
		{
			name:   "successfully register an AIM session",
			cookie: aimCookie,
			mockParams: mockParams{
				cookieBakerParams: cookieBakerParams{
					cookieCrackParams: cookieCrackParams{
						{
							dataOut:  aimCookie,
							cookieIn: aimCookie,
						},
					},
				},
				sessionRegistryParams: sessionRegistryParams{
					addSessionParams: addSessionParams{
						{
							screenName: screenName,
							result:     newTestSession(screenName),
						},
					},
				},
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: screenName.IdentScreenName(),
							result: &state.User{
								IdentScreenName:   screenName.IdentScreenName(),
								DisplayScreenName: screenName,
							},
						},
					},
				},
				accountManagerParams: accountManagerParams{
					accountManagerConfirmStatusByNameParams: accountManagerConfirmStatusByNameParams{
						{
							screenName:    screenName.IdentScreenName(),
							confirmStatus: true,
						},
					},
				},
			},
			wantSess: func(session *state.Session) bool {
				return true
			},
		},
		{
			name:   "successfully register an ICQ session",
			cookie: icqCookie,
			mockParams: mockParams{
				cookieBakerParams: cookieBakerParams{
					cookieCrackParams: cookieCrackParams{
						{
							dataOut:  icqCookie,
							cookieIn: icqCookie,
						},
					},
				},
				sessionRegistryParams: sessionRegistryParams{
					addSessionParams: addSessionParams{
						{
							screenName: uin,
							result:     newTestSession(uin),
						},
					},
				},
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: uin.IdentScreenName(),
							result: &state.User{
								IdentScreenName:   uin.IdentScreenName(),
								DisplayScreenName: uin,
							},
						},
					},
				},
				accountManagerParams: accountManagerParams{
					accountManagerConfirmStatusByNameParams: accountManagerConfirmStatusByNameParams{
						{
							screenName:    uin.IdentScreenName(),
							confirmStatus: true,
						},
					},
				},
			},
			wantSess: func(sess *state.Session) bool {
				uinMatches := fmt.Sprintf("%d", sess.UIN()) == uin.String()
				flagsMatch := sess.UserInfoBitmask()&wire.OServiceUserFlagICQ == wire.OServiceUserFlagICQ
				return uinMatches && flagsMatch
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionRegistry := newMockSessionRegistry(t)
			for _, params := range tc.mockParams.addSessionParams {
				sessionRegistry.EXPECT().
					AddSession(params.screenName).
					Return(params.result)
			}
			cookieBaker := newMockCookieBaker(t)
			for _, params := range tc.mockParams.cookieCrackParams {
				cookieBaker.EXPECT().
					Crack(params.cookieIn).
					Return(params.dataOut, nil)
			}
			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.userManagerParams.getUserParams {
				userManager.EXPECT().
					User(params.screenName).
					Return(params.result, nil)
			}
			accountManager := newMockAccountManager(t)
			for _, params := range tc.mockParams.accountManagerConfirmStatusByNameParams {
				accountManager.EXPECT().
					ConfirmStatusByName(params.screenName).
					Return(params.confirmStatus, nil)
			}

			svc := NewAuthService(config.Config{}, sessionRegistry, nil, userManager, cookieBaker, nil, nil, accountManager, nil, nil)

			have, err := svc.RegisterBOSSession(tc.cookie)
			assert.NoError(t, err)

			if tc.wantSess != nil {
				assert.True(t, tc.wantSess(have))
			}
		})
	}

}

func TestAuthService_RetrieveBOSSession_HappyPath(t *testing.T) {
	sess := newTestSession("screenName")

	aimAuthCookie := bosCookie{
		ScreenName: sess.DisplayScreenName(),
	}
	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(aimAuthCookie, buf))
	authCookie := buf.Bytes()

	sessionRetriever := newMockSessionRetriever(t)
	sessionRetriever.EXPECT().
		RetrieveSession(sess.IdentScreenName()).
		Return(sess)

	cookieBaker := newMockCookieBaker(t)
	cookieBaker.EXPECT().
		Crack(authCookie).
		Return(authCookie, nil)

	userManager := newMockUserManager(t)
	userManager.EXPECT().
		User(sess.IdentScreenName()).
		Return(&state.User{IdentScreenName: sess.IdentScreenName()}, nil)

	svc := NewAuthService(config.Config{}, nil, nil, userManager, cookieBaker, nil, nil, nil, nil, sessionRetriever)

	have, err := svc.RetrieveBOSSession(authCookie)
	assert.NoError(t, err)
	assert.Equal(t, sess, have)
}

func TestAuthService_RetrieveBOSSession_SessionNotFound(t *testing.T) {
	sess := newTestSession("screenName")

	aimAuthCookie := bosCookie{
		ScreenName: sess.DisplayScreenName(),
	}
	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(aimAuthCookie, buf))
	authCookie := buf.Bytes()

	sessionRetriever := newMockSessionRetriever(t)
	sessionRetriever.EXPECT().
		RetrieveSession(sess.IdentScreenName()).
		Return(nil)

	cookieBaker := newMockCookieBaker(t)

	cookieBaker.EXPECT().
		Crack(authCookie).
		Return(authCookie, nil)

	userManager := newMockUserManager(t)
	userManager.EXPECT().
		User(sess.IdentScreenName()).
		Return(&state.User{IdentScreenName: sess.IdentScreenName()}, nil)

	svc := NewAuthService(config.Config{}, nil, nil, userManager, cookieBaker, nil, nil, nil, nil, sessionRetriever)

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
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "user signs out of chat room, room is empty after user leaves",
			userSession: newTestSession("me", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie")),
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					chatRelayToAllExceptParams: chatRelayToAllExceptParams{
						{
							screenName: state.NewIdentScreenName("me"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatUsersLeft,
								},
								Body: wire.SNAC_0x0E_0x04_ChatUsersLeft{
									Users: []wire.TLVUserInfo{
										newTestSession("me", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie")).TLVUserInfo(),
									},
								},
							},
						},
					},
				},
				sessionRegistryParams: sessionRegistryParams{
					removeSessionParams: removeSessionParams{
						{
							screenName: state.NewIdentScreenName("me"),
						},
					},
				},
			},
		},
		{
			name:        "user signs out of chat room, room is not empty after user leaves",
			userSession: newTestSession("me", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie")),
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					chatRelayToAllExceptParams: chatRelayToAllExceptParams{
						{
							screenName: state.NewIdentScreenName("me"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatUsersLeft,
								},
								Body: wire.SNAC_0x0E_0x04_ChatUsersLeft{
									Users: []wire.TLVUserInfo{
										newTestSession("me", sessOptCannedSignonTime, sessOptChatRoomCookie("the-chat-cookie")).TLVUserInfo(),
									},
								},
							},
						},
					},
				},
				sessionRegistryParams: sessionRegistryParams{
					removeSessionParams: removeSessionParams{
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
			chatMessageRelayer := newMockChatMessageRelayer(t)
			for _, params := range tt.mockParams.chatRelayToAllExceptParams {
				chatMessageRelayer.EXPECT().
					RelayToAllExcept(nil, tt.userSession.ChatRoomCookie(), params.screenName, params.message)
			}
			sessionManager := newMockChatSessionRegistry(t)
			for _, params := range tt.mockParams.removeSessionParams {
				sessionManager.EXPECT().
					RemoveSession(matchSession(params.screenName))
			}

			svc := NewAuthService(config.Config{}, nil, sessionManager, nil, nil, nil, chatMessageRelayer, nil, nil, nil)
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
			userSession: newTestSession("me", sessOptCannedSignonTime),
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyDepartedParams: broadcastBuddyDepartedParams{
						{
							screenName: state.NewIdentScreenName("me"),
						},
					},
				},
				sessionRegistryParams: sessionRegistryParams{
					removeSessionParams: removeSessionParams{
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
			sessionManager := newMockSessionRegistry(t)
			for _, params := range tt.mockParams.removeSessionParams {
				sessionManager.EXPECT().RemoveSession(matchSession(params.screenName))
			}
			buddyUpdateBroadcaster := newMockbuddyBroadcaster(t)
			for _, params := range tt.mockParams.broadcastBuddyDepartedParams {
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyDeparted(mock.Anything, matchSession(params.screenName)).
					Return(params.err)
			}
			svc := NewAuthService(config.Config{}, sessionManager, nil, nil, nil, nil, nil, nil, nil, nil)
			svc.buddyBroadcaster = buddyUpdateBroadcaster

			err := svc.Signout(nil, tt.userSession)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
