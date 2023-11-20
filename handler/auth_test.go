package handler

import (
	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestReceiveAndSendBUCPLoginRequest(t *testing.T) {
	userGoodPwd := state.User{
		ScreenName: "sn_user_a",
		AuthKey:    "auth_key_user",
	}
	assert.NoError(t, userGoodPwd.HashPassword("good_pwd"))
	userBadPwd := userGoodPwd
	assert.NoError(t, userBadPwd.HashPassword("bad_pwd"))

	cases := []struct {
		name        string
		cfg         server.Config
		userInDB    state.User
		sessionUUID uuid.UUID
		inputSNAC   oscar.SNAC_0x17_0x02_BUCPLoginRequest
		// expectOutput is the SNAC payload sent from the server to the
		// recipient client
		expectOutput oscar.XMessage
	}{
		{
			name: "login with valid password, expect OK login response",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			userInDB:    userGoodPwd,
			sessionUUID: uuid.UUID{1, 2, 3},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, userGoodPwd.PassHash),
						oscar.NewTLV(oscar.TLVScreenName, userGoodPwd.ScreenName),
					},
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				SnacOut: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVScreenName, userGoodPwd.ScreenName),
							oscar.NewTLV(oscar.TLVReconnectHere, "127.0.0.1:1234"),
							oscar.NewTLV(oscar.TLVAuthorizationCookie, uuid.UUID{1, 2, 3}.String()),
						},
					},
				},
			},
		},
		{
			name: "login with bad password, expect OK login response (Cfg.DisableAuth=true)",
			cfg: server.Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     1234,
				DisableAuth: true,
			},
			userInDB:    userGoodPwd,
			sessionUUID: uuid.UUID{1, 2, 3},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, userBadPwd.PassHash),
						oscar.NewTLV(oscar.TLVScreenName, userBadPwd.ScreenName),
					},
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				SnacOut: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVScreenName, userBadPwd.ScreenName),
							oscar.NewTLV(oscar.TLVReconnectHere, "127.0.0.1:1234"),
							oscar.NewTLV(oscar.TLVAuthorizationCookie, uuid.UUID{1, 2, 3}.String()),
						},
					},
				},
			},
		},
		{
			name: "login with bad password, expect failed login response (Cfg.DisableAuth=false)",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			userInDB:    userGoodPwd,
			sessionUUID: uuid.UUID{1, 2, 3},
			inputSNAC: oscar.SNAC_0x17_0x02_BUCPLoginRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVPasswordHash, userBadPwd.PassHash),
						oscar.NewTLV(oscar.TLVScreenName, userBadPwd.ScreenName),
					},
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				SnacOut: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVScreenName, userBadPwd.ScreenName),
							oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess := newTestSession(tc.userInDB.ScreenName, sessOptID(tc.sessionUUID.String()))
			um := NewMockUserManager(t)
			um.EXPECT().
				GetUser(tc.userInDB.ScreenName).
				Return(&userGoodPwd, nil).
				Maybe()
			um.EXPECT().
				UpsertUser(mock.Anything).
				Return(nil).
				Maybe()
			sm := NewMockSessionManager(t)
			sm.EXPECT().
				NewSessionWithSN(tc.sessionUUID.String(), tc.userInDB.ScreenName).
				Return(sess).
				Maybe()
			svc := AuthService{
				cfg: tc.cfg,
				sm:  sm,
				um:  um,
			}
			fnNewUUID := func() uuid.UUID {
				return tc.sessionUUID
			}
			outputSNAC, err := svc.ReceiveAndSendBUCPLoginRequest(tc.inputSNAC, fnNewUUID)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestReceiveAndSendAuthChallenge(t *testing.T) {
	cases := []struct {
		name         string
		cfg          server.Config
		userInDB     *state.User
		fnNewUUID    uuid.UUID
		inputSNAC    oscar.SNAC_0x17_0x06_BUCPChallengeRequest
		expectOutput oscar.XMessage
	}{
		{
			name: "login with valid username, expect OK login response",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			userInDB: &state.User{
				ScreenName: "sn_user_a",
				AuthKey:    "auth_key_user_a",
			},
			fnNewUUID: uuid.UUID{1, 2, 3},
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_a"),
					},
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPChallengeResponse,
				},
				SnacOut: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
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
			userInDB:  nil,
			fnNewUUID: uuid.UUID{1, 2, 3},
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_b"),
					},
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPChallengeResponse,
				},
				SnacOut: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
					AuthKey: uuid.UUID{1, 2, 3}.String(),
				},
			},
		},
		{
			name: "login with invalid username, expect failed login response (Cfg.DisableAuth=false)",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			userInDB:  nil,
			fnNewUUID: uuid.UUID{1, 2, 3},
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_b"),
					},
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.BUCP,
					SubGroup:  oscar.BUCPLoginResponse,
				},
				SnacOut: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			um := NewMockUserManager(t)
			um.EXPECT().
				GetUser(string(tc.inputSNAC.TLVList[0].Val)).
				Return(tc.userInDB, nil).
				Maybe()
			svc := AuthService{
				cfg: tc.cfg,
				um:  um,
			}
			fnNewUUID := func() uuid.UUID {
				return tc.fnNewUUID
			}
			outputSNAC, err := svc.ReceiveAndSendAuthChallenge(tc.inputSNAC, fnNewUUID)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
