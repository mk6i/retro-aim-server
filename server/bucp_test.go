package server

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

func TestReceiveAndSendBUCPLoginRequest(t *testing.T) {
	userGoodPwd := User{
		ScreenName: "sn_user_a",
		AuthKey:    "auth_key_user",
	}
	assert.NoError(t, userGoodPwd.HashPassword("good_pwd"))
	userBadPwd := userGoodPwd
	assert.NoError(t, userBadPwd.HashPassword("bad_pwd"))

	cases := []struct {
		name            string
		cfg             Config
		userInDB        User
		sessionUUID     uuid.UUID
		inputSNAC       oscar.SNAC_0x17_0x02_BUCPLoginRequest
		expectSnacFrame oscar.SnacFrame
		expectSNACBody  oscar.SNAC_0x17_0x03_BUCPLoginResponse
	}{
		{
			name: "login with valid password, expect OK login response",
			cfg: Config{
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
			expectSnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.BUCP,
				SubGroup:  BUCPLoginResponse,
			},
			expectSNACBody: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, userGoodPwd.ScreenName),
						oscar.NewTLV(oscar.TLVReconnectHere, "127.0.0.1:1234"),
						oscar.NewTLV(oscar.TLVAuthorizationCookie, uuid.UUID{1, 2, 3}.String()),
					},
				},
			},
		},
		{
			name: "login with bad password, expect OK login response (cfg.DisableAuth=true)",
			cfg: Config{
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
			expectSnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.BUCP,
				SubGroup:  BUCPLoginResponse,
			},
			expectSNACBody: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, userBadPwd.ScreenName),
						oscar.NewTLV(oscar.TLVReconnectHere, "127.0.0.1:1234"),
						oscar.NewTLV(oscar.TLVAuthorizationCookie, uuid.UUID{1, 2, 3}.String()),
					},
				},
			},
		},
		{
			name: "login with bad password, expect failed login response (cfg.DisableAuth=false)",
			cfg: Config{
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
			expectSnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.BUCP,
				SubGroup:  BUCPLoginResponse,
			},
			expectSNACBody: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, userBadPwd.ScreenName),
						oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			const testFile string = "aim_test.db"
			defer func() {
				assert.NoError(t, os.Remove(testFile))
			}()
			fs, err := NewSQLiteFeedbagStore(testFile)
			if err != nil {
				assert.NoError(t, err)
			}
			assert.NoError(t, fs.InsertUser(tc.userInDB))
			sm := NewSessionManager(NewLogger(Config{}))
			//
			// send input SNAC
			//
			input := &bytes.Buffer{}
			var seq uint32
			assert.NoError(t, writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, tc.inputSNAC, &seq, input))
			//
			// receive response
			//
			output := &bytes.Buffer{}
			fnNewUUID := func() uuid.UUID {
				return tc.sessionUUID
			}
			assert.NoError(t, ReceiveAndSendBUCPLoginRequest(tc.cfg, sm, fs, input, output, &seq, fnNewUUID))
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, output))
			//
			// verify output SNAC frame
			//
			SnacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&SnacFrame, output))
			assert.Equal(t, tc.expectSnacFrame, SnacFrame)
			//
			// verify output SNAC body
			//
			actual := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
			assert.NoError(t, oscar.Unmarshal(&actual, output))
			assert.Equal(t, tc.expectSNACBody, actual)
			assert.Equalf(t, 0, output.Len(), "the rest of the buffer is unread")
		})
	}
}

func TestReceiveAndSendAuthChallenge(t *testing.T) {
	cases := []struct {
		name            string
		cfg             Config
		userInDB        User
		fnNewUUID       uuid.UUID
		inputSNAC       oscar.SNAC_0x17_0x06_BUCPChallengeRequest
		expectSnacFrame oscar.SnacFrame
		expectSNACBody  any
	}{
		{
			name: "login with valid username, expect OK login response",
			cfg: Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			userInDB: User{
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
			expectSnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.BUCP,
				SubGroup:  BUCPChallengeResponse,
			},
			expectSNACBody: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
				AuthKey: "auth_key_user_a",
			},
		},
		{
			name: "login with invalid username, expect OK login response (cfg.DisableAuth=true)",
			cfg: Config{
				OSCARHost:   "127.0.0.1",
				BOSPort:     1234,
				DisableAuth: true,
			},
			userInDB: User{
				ScreenName: "sn_user_a",
				AuthKey:    "auth_key_user_a",
			},
			fnNewUUID: uuid.UUID{1, 2, 3},
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_b"),
					},
				},
			},
			expectSnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.BUCP,
				SubGroup:  BUCPChallengeResponse,
			},
			expectSNACBody: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
				AuthKey: uuid.UUID{1, 2, 3}.String(),
			},
		},
		{
			name: "login with invalid username, expect failed login response (cfg.DisableAuth=false)",
			cfg: Config{
				OSCARHost: "127.0.0.1",
				BOSPort:   1234,
			},
			userInDB: User{
				ScreenName: "sn_user_a",
				AuthKey:    "auth_key_user_a",
			},
			fnNewUUID: uuid.UUID{1, 2, 3},
			inputSNAC: oscar.SNAC_0x17_0x06_BUCPChallengeRequest{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVScreenName, "sn_user_b"),
					},
				},
			},
			expectSnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.BUCP,
				SubGroup:  BUCPLoginResponse,
			},
			expectSNACBody: oscar.SNAC_0x17_0x03_BUCPLoginResponse{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			const testFile string = "aim_test.db"
			defer func() {
				assert.NoError(t, os.Remove(testFile))
			}()
			fs, err := NewSQLiteFeedbagStore(testFile)
			if err != nil {
				assert.NoError(t, err)
			}
			assert.NoError(t, fs.InsertUser(tc.userInDB))
			//
			// send input SNAC
			//
			input := &bytes.Buffer{}
			var seq uint32
			assert.NoError(t, writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, tc.inputSNAC, &seq, input))
			//
			// receive response
			//
			output := &bytes.Buffer{}
			fnNewUUID := func() uuid.UUID {
				return tc.fnNewUUID
			}
			assert.NoError(t, ReceiveAndSendAuthChallenge(tc.cfg, fs, input, output, &seq, fnNewUUID))
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, output))
			//
			// verify output SNAC frame
			//
			SnacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&SnacFrame, output))
			assert.Equal(t, tc.expectSnacFrame, SnacFrame)
			//
			// verify output SNAC body
			//
			switch v := tc.expectSNACBody.(type) {
			case oscar.SNAC_0x17_0x07_BUCPChallengeResponse:
				actual := oscar.SNAC_0x17_0x07_BUCPChallengeResponse{}
				assert.NoError(t, oscar.Unmarshal(&actual, output))
				assert.Equal(t, v, actual)
			case oscar.SNAC_0x17_0x03_BUCPLoginResponse:
				actual := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
				assert.NoError(t, oscar.Unmarshal(&actual, output))
				assert.Equal(t, v, actual)
			default:
				t.Fatalf("unexpected output SNAC type")
			}
			assert.Equalf(t, 0, output.Len(), "the rest of the buffer is unread")
		})
	}
}
