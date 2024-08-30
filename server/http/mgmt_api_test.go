package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func TestSessionHandler_GET(t *testing.T) {
	fnNewSess := func(screenName string, uin uint32) *state.Session {
		sess := state.NewSession()
		sess.SetIdentScreenName(state.NewIdentScreenName(screenName))
		sess.SetDisplayScreenName(state.DisplayScreenName(screenName))
		sess.SetUIN(uin)
		return sess
	}
	tt := []struct {
		name          string
		want          string
		statusCode    int
		timeSinceFunc func(t time.Time) time.Duration
		mockParams    mockParams
	}{
		{
			name:       "without sessions",
			want:       `{"count":0,"sessions":[]}`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				sessionRetrieverParams: sessionRetrieverParams{
					sessionRetrieverAllSessionsParams: sessionRetrieverAllSessionsParams{
						{
							result: []*state.Session{},
						},
					},
				},
			},
		},
		{
			name:          "with sessions",
			want:          `{"count":3,"sessions":[{"id":"usera","screen_name":"userA","online_seconds":0,"away_message":"","idle_seconds":0,"is_icq":false},{"id":"userb","screen_name":"userB","online_seconds":0,"away_message":"","idle_seconds":0,"is_icq":false},{"id":"100003","screen_name":"100003","online_seconds":0,"away_message":"","idle_seconds":0,"is_icq":true}]}`,
			statusCode:    http.StatusOK,
			timeSinceFunc: func(t time.Time) time.Duration { t0 := time.Now(); return t0.Sub(t0) },
			mockParams: mockParams{
				sessionRetrieverParams: sessionRetrieverParams{
					sessionRetrieverAllSessionsParams: sessionRetrieverAllSessionsParams{
						{
							result: []*state.Session{
								fnNewSess("userA", 0),
								fnNewSess("userB", 0),
								fnNewSess("100003", 100003),
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/session", nil)
			responseRecorder := httptest.NewRecorder()

			sessionRetriever := newMockSessionRetriever(t)
			for _, params := range tc.mockParams.sessionRetrieverParams.sessionRetrieverAllSessionsParams {
				sessionRetriever.EXPECT().
					AllSessions().
					Return(params.result)
			}

			getSessionHandler(responseRecorder, request, sessionRetriever, tc.timeSinceFunc)

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestSessionHandlerScreenname_GET(t *testing.T) {
	fnNewSess := func(screenName string, uin uint32) *state.Session {
		sess := state.NewSession()
		sess.SetIdentScreenName(state.NewIdentScreenName(screenName))
		sess.SetDisplayScreenName(state.DisplayScreenName(screenName))
		sess.SetUIN(uin)
		return sess
	}
	tt := []struct {
		name              string
		sessions          []*state.Session
		requestScreenName state.IdentScreenName
		want              string
		statusCode        int
		timeSinceFunc     func(t time.Time) time.Duration
		mockParams        mockParams
	}{
		{
			name:              "no session for screenname",
			sessions:          []*state.Session{},
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              `session not found`,
			statusCode:        http.StatusNotFound,
			mockParams: mockParams{
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionByNameParams: retrieveSessionByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result:     nil,
						},
					},
				},
			},
		},
		{
			name:              "active session found for screenname",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              `{"count":1,"sessions":[{"id":"usera","screen_name":"userA","online_seconds":0,"away_message":"","idle_seconds":0,"is_icq":false}]}`,
			statusCode:        http.StatusOK,
			timeSinceFunc:     func(t time.Time) time.Duration { t0 := time.Now(); return t0.Sub(t0) },
			mockParams: mockParams{
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionByNameParams: retrieveSessionByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result:     fnNewSess("userA", 0),
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/session/"+tc.requestScreenName.String(), nil)
			request.SetPathValue("screenname", tc.requestScreenName.String())
			responseRecorder := httptest.NewRecorder()

			sessionRetriever := newMockSessionRetriever(t)
			for _, params := range tc.mockParams.sessionRetrieverParams.retrieveSessionByNameParams {
				sessionRetriever.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.result)
			}

			getSessionHandler(responseRecorder, request, sessionRetriever, tc.timeSinceFunc)

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestUserAccountHandler_GET(t *testing.T) {
	tt := []struct {
		name              string
		requestScreenName state.IdentScreenName
		want              string
		statusCode        int
		mockParams        mockParams
	}{
		{
			name:              "invalid account",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              `user not found`,
			statusCode:        http.StatusNotFound,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result:     nil,
						},
					},
				},
			},
		},
		{
			name:              "valid aim account",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              `{"id":"usera","screen_name":"userA","profile":"My Profile Text","email_address":"\u003cuserA@aol.com\u003e","reg_status":2,"confirmed":true,"is_icq":false}`,
			statusCode:        http.StatusOK,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &state.User{
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
						},
					},
				},
				accountRetrieverParams: accountRetrieverParams{
					emailAddressByNameParams: emailAddressByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &mail.Address{
								Address: "userA@aol.com",
							},
						},
					},
					regStatusByNameParams: regStatusByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result:     uint16(0x02),
						},
					},
					confirmStatusByNameParams: confirmStatusByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result:     true,
						},
					},
				},
				profileRetrieverParams: profileRetrieverParams{
					retrieveProfileParams: retrieveProfileParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result:     "My Profile Text",
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/user/"+tc.requestScreenName.String()+"/account", nil)
			request.SetPathValue("screenname", tc.requestScreenName.String())
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.userManagerParams.getUserParams {
				userManager.EXPECT().
					User(params.screenName).
					Return(params.result, params.err)
			}

			accountRetriever := newMockAccountRetriever(t)
			for _, params := range tc.mockParams.accountRetrieverParams.emailAddressByNameParams {
				accountRetriever.EXPECT().
					EmailAddressByName(params.screenName).
					Return(params.result, params.err)
			}
			for _, params := range tc.mockParams.accountRetrieverParams.regStatusByNameParams {
				accountRetriever.EXPECT().
					RegStatusByName(params.screenName).
					Return(params.result, params.err)
			}
			for _, params := range tc.mockParams.accountRetrieverParams.confirmStatusByNameParams {
				accountRetriever.EXPECT().
					ConfirmStatusByName(params.screenName).
					Return(params.result, params.err)
			}

			profileRetriever := newMockProfileRetriever(t)
			for _, params := range tc.mockParams.profileRetrieverParams.retrieveProfileParams {
				profileRetriever.EXPECT().
					Profile(params.screenName).
					Return(params.result, params.err)
			}

			getUserAccountHandler(responseRecorder, request, userManager, accountRetriever, profileRetriever, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestUserBuddyIconHandler_GET(t *testing.T) {
	sampleGIF := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x32, 0x00, 0x32, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
		0x32, 0x00, 0x32, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b}

	sampleJPG := []byte{0xFF, 0xD8, 0xFF, 0x43, 0x13, 0x37}
	tt := []struct {
		name              string
		requestScreenName state.IdentScreenName
		want              string
		statusCode        int
		contentType       string
		mockParams        mockParams
	}{
		{
			name:              "invalid account",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              `user not found`,
			statusCode:        http.StatusNotFound,
			contentType:       "text/plain; charset=utf-8",
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result:     nil,
						},
					},
				},
			},
		},
		{
			name:              "account with gif buddy icon",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              string(sampleGIF),
			statusCode:        http.StatusOK,
			contentType:       "image/gif",
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &state.User{
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
						},
					},
				},
				feedBagRetrieverParams: feedBagRetrieverParams{
					buddyIconRefByNameParams: buddyIconRefByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &wire.BARTID{
								Type: wire.BARTTypesBuddyIcon,
								BARTInfo: wire.BARTInfo{
									Flags: 0x00,
									Hash:  []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
								},
							},
						},
					},
				},
				bartRetrieverParams: bartRetrieverParams{
					bartRetrieveParams: bartRetrieveParams{
						{
							itemHash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
							result:   sampleGIF,
						},
					},
				},
			},
		},
		{
			name:              "account with jpg buddy icon",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              string(sampleJPG),
			statusCode:        http.StatusOK,
			contentType:       "image/jpeg",
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &state.User{
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
						},
					},
				},
				feedBagRetrieverParams: feedBagRetrieverParams{
					buddyIconRefByNameParams: buddyIconRefByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &wire.BARTID{
								Type: wire.BARTTypesBuddyIcon,
								BARTInfo: wire.BARTInfo{
									Flags: 0x00,
									Hash:  []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
								},
							},
						},
					},
				},
				bartRetrieverParams: bartRetrieverParams{
					bartRetrieveParams: bartRetrieveParams{
						{
							itemHash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
							result:   sampleJPG,
						},
					},
				},
			},
		},
		{
			name:              "account with unknown format buddy icon",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              string([]byte{0x13, 0x37, 0x13, 0x37, 0x13, 0x37}),
			statusCode:        http.StatusOK,
			contentType:       "application/octet-stream",
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &state.User{
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
						},
					},
				},
				feedBagRetrieverParams: feedBagRetrieverParams{
					buddyIconRefByNameParams: buddyIconRefByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &wire.BARTID{
								Type: wire.BARTTypesBuddyIcon,
								BARTInfo: wire.BARTInfo{
									Flags: 0x00,
									Hash:  []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
								},
							},
						},
					},
				},
				bartRetrieverParams: bartRetrieverParams{
					bartRetrieveParams: bartRetrieveParams{
						{
							itemHash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
							result:   []byte{0x13, 0x37, 0x13, 0x37, 0x13, 0x37},
						},
					},
				},
			},
		},
		{
			name:              "account with cleared buddy icon",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              "icon not found",
			statusCode:        http.StatusNotFound,
			contentType:       "text/plain; charset=utf-8",
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &state.User{
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
						},
					},
				},
				feedBagRetrieverParams: feedBagRetrieverParams{
					buddyIconRefByNameParams: buddyIconRefByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &wire.BARTID{
								Type: wire.BARTTypesBuddyIcon,
								BARTInfo: wire.BARTInfo{
									Flags: 0x00,
									Hash:  wire.GetClearIconHash(),
								},
							},
						},
					},
				},
			},
		},
		{
			name:              "account with no buddy icon",
			requestScreenName: state.NewIdentScreenName("userA"),
			want:              "icon not found",
			statusCode:        http.StatusNotFound,
			contentType:       "text/plain; charset=utf-8",
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					getUserParams: getUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result: &state.User{
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
						},
					},
				},
				feedBagRetrieverParams: feedBagRetrieverParams{
					buddyIconRefByNameParams: buddyIconRefByNameParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							result:     nil,
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/user/"+tc.requestScreenName.String()+"/icon", nil)
			request.SetPathValue("screenname", tc.requestScreenName.String())
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.userManagerParams.getUserParams {
				userManager.EXPECT().
					User(params.screenName).
					Return(params.result, params.err)
			}

			feedbagRetriever := newMockFeedBagRetriever(t)
			for _, params := range tc.mockParams.feedBagRetrieverParams.buddyIconRefByNameParams {
				feedbagRetriever.EXPECT().
					BuddyIconRefByName(params.screenName).
					Return(params.result, params.err)
			}

			bartRetriever := newMockBARTRetriever(t)
			for _, params := range tc.mockParams.bartRetrieverParams.bartRetrieveParams {
				bartRetriever.EXPECT().
					BARTRetrieve(params.itemHash).
					Return(params.result, params.err)
			}

			getUserBuddyIconHandler(responseRecorder, request, userManager, feedbagRetriever, bartRetriever, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			contentType := responseRecorder.Header().Get("Content-Type")
			if contentType != tc.contentType {
				t.Errorf("Want content type '%s', got '%s'", tc.contentType, contentType)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestUserHandler_GET(t *testing.T) {
	tt := []struct {
		name       string
		want       string
		statusCode int
		mockParams mockParams
	}{
		{
			name:       "empty user store",
			want:       `[]`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					allUsersParams: allUsersParams{
						{
							result: []state.User{},
						},
					},
				},
			},
		},
		{
			name:       "user store containing 3 users",
			want:       `[{"id":"usera","screen_name":"userA","is_icq":false},{"id":"userb","screen_name":"userB","is_icq":false},{"id":"100003","screen_name":"100003","is_icq":true}]`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					allUsersParams: allUsersParams{
						{
							result: []state.User{
								{
									DisplayScreenName: "userA",
									IdentScreenName:   state.NewIdentScreenName("userA"),
								},
								{
									DisplayScreenName: "userB",
									IdentScreenName:   state.NewIdentScreenName("userB"),
								},
								{
									DisplayScreenName: "100003",
									IdentScreenName:   state.NewIdentScreenName("100003"),
									IsICQ:             true,
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "user handler error",
			want:       `internal server error`,
			statusCode: http.StatusInternalServerError,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					allUsersParams: allUsersParams{
						{
							result: []state.User{},
							err:    io.EOF,
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.userManagerParams.allUsersParams {
				userManager.EXPECT().
					AllUsers().
					Return(params.result, params.err)
			}

			getUserHandler(responseRecorder, userManager, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestUserHandler_POST(t *testing.T) {
	tt := []struct {
		name       string
		body       string
		UUID       uuid.UUID
		want       string
		password   string
		statusCode int
		mockParams mockParams
	}{
		{
			name: "with valid AIM user",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),

			want:       `User account created successfully.`,
			password:   "thepassword",
			statusCode: http.StatusCreated,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					insertUserParams: insertUserParams{
						{
							u: state.User{
								AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
							err: nil,
						},
					},
				},
			},
		},
		{
			name:       "with valid ICQ user",
			body:       `{"screen_name":"100003", "password":"thepass"}`,
			UUID:       uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			want:       `User account created successfully.`,
			password:   "thepass",
			statusCode: http.StatusCreated,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					insertUserParams: insertUserParams{
						{
							u: state.User{
								AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
								DisplayScreenName: "100003",
								IdentScreenName:   state.NewIdentScreenName("100003"),
								IsICQ:             true,
							},
							err: nil,
						},
					},
				},
			},
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA", "password":"thepassword"`, // missing closing }
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "user handler error",
			body:       `{"screen_name":"userA", "password":"thepassword"}`,
			UUID:       uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			want:       `internal server error`,
			password:   "thepassword",
			statusCode: http.StatusInternalServerError,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					insertUserParams: insertUserParams{
						{
							u: state.User{
								AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
							err: io.EOF,
						},
					},
				},
			},
		},
		{
			name:       "duplicate user",
			body:       `{"screen_name":"userA", "password":"thepassword"}`,
			UUID:       uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			want:       `user already exists`,
			password:   "thepassword",
			statusCode: http.StatusConflict,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					insertUserParams: insertUserParams{
						{
							u: state.User{
								AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
								DisplayScreenName: "userA",
								IdentScreenName:   state.NewIdentScreenName("userA"),
							},
							err: state.ErrDupUser,
						},
					},
				},
			},
		},
		{
			name:       "invalid AIM screen name",
			body:       `{"screen_name":"a", "password":"thepassword"}`,
			UUID:       uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			want:       `invalid screen name: screen name must be between 3 and 16 characters`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid AIM password",
			body:       `{"screen_name":"userA", "password":"1"}`,
			UUID:       uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			want:       `invalid password: invalid password length: password length must be between 4-16 characters`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid ICQ UIN",
			body:       `{"screen_name":"1000", "password":"thepass"}`,
			UUID:       uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			want:       `invalid uin: uin must be a number in the range 10000-2147483646`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid ICQ password",
			body:       `{"screen_name":"100003", "password":"thelongpassword"}`,
			UUID:       uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			want:       `invalid password: invalid password length: password must be between 6-8 characters`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.userManagerParams.insertUserParams {
				assert.NoError(t, params.u.HashPassword(tc.password))
				userManager.EXPECT().
					InsertUser(params.u).
					Return(params.err)
			}

			newUUID := func() uuid.UUID { return tc.UUID }
			postUserHandler(responseRecorder, request, userManager, newUUID, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestUserHandler_DELETE(t *testing.T) {
	tt := []struct {
		name       string
		body       string
		want       string
		statusCode int
		mockParams mockParams
	}{
		{
			name:       "with valid user",
			body:       `{"screen_name":"userA"}`,
			want:       `User account successfully deleted.`,
			statusCode: http.StatusNoContent,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					deleteUserParams: deleteUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
						},
					},
				},
			},
		},
		{
			name:       "with non-existent user",
			body:       `{"screen_name":"userA"}`,
			want:       `user does not exist`,
			statusCode: http.StatusNotFound,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					deleteUserParams: deleteUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							err:        state.ErrNoUser,
						},
					},
				},
			},
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA"`, // missing closing }
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "user handler error",
			body:       `{"screen_name":"userA"}`,
			want:       `internal server error`,
			statusCode: http.StatusInternalServerError,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					deleteUserParams: deleteUserParams{
						{
							screenName: state.NewIdentScreenName("userA"),
							err:        io.EOF,
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/user", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.userManagerParams.deleteUserParams {
				userManager.EXPECT().
					DeleteUser(params.screenName).
					Return(params.err)
			}

			deleteUserHandler(responseRecorder, request, userManager, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestUserPasswordHandler_PUT(t *testing.T) {
	tt := []struct {
		name       string
		body       string
		want       string
		statusCode int
		mockParams mockParams
	}{
		{
			name:       "user with valid password",
			body:       `{"screen_name":"userA", "password":"thenewpassword"}`,
			want:       `Password successfully reset.`,
			statusCode: http.StatusNoContent,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					setUserPasswordParams: setUserPasswordParams{
						{
							screenName:  state.NewIdentScreenName("userA"),
							newPassword: "thenewpassword",
						},
					},
				},
			},
		},
		{
			name:       "user with invalid password",
			body:       `{"screen_name":"userA", "password":"a"}`,
			want:       `invalid password length`,
			statusCode: http.StatusBadRequest,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					setUserPasswordParams: setUserPasswordParams{
						{
							screenName:  state.NewIdentScreenName("userA"),
							newPassword: "a",
							err:         state.ErrPasswordInvalid,
						},
					},
				},
			},
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA", "password":"thepassword"`, // missing closing }
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "password updater returns runtime error",
			body:       `{"screen_name":"userA", "password":"thepassword"}`,
			want:       `internal server error`,
			statusCode: http.StatusInternalServerError,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					setUserPasswordParams: setUserPasswordParams{
						{
							screenName:  state.NewIdentScreenName("userA"),
							newPassword: "thepassword",
							err:         io.EOF,
						},
					},
				},
			},
		},
		{
			name:       "user doesn't exist",
			body:       `{"screen_name":"userA", "password":"thepassword"}`,
			want:       `user does not exist`,
			statusCode: http.StatusNotFound,
			mockParams: mockParams{
				userManagerParams: userManagerParams{
					setUserPasswordParams: setUserPasswordParams{
						{
							screenName:  state.NewIdentScreenName("userA"),
							newPassword: "thepassword",
							err:         state.ErrNoUser,
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/user", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			for _, params := range tc.mockParams.userManagerParams.setUserPasswordParams {
				userManager.EXPECT().
					SetUserPassword(params.screenName, params.newPassword).
					Return(params.err)
			}

			putUserPasswordHandler(responseRecorder, request, userManager, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestPublicChatHandler_GET(t *testing.T) {
	fnNewSess := func(screenName string) *state.Session {
		sess := state.NewSession()
		sess.SetIdentScreenName(state.NewIdentScreenName(screenName))
		sess.SetDisplayScreenName(state.DisplayScreenName(screenName))
		return sess
	}

	chatRoom1 := state.NewChatRoom("chat-room-1-name", state.NewIdentScreenName("chat-room-1-creator"), state.PublicExchange)
	chatRoom2 := state.NewChatRoom("chat-room-2-name", state.NewIdentScreenName("chat-room-1-creator"), state.PublicExchange)

	tt := []struct {
		name       string
		want       string
		statusCode int
		mockParams mockParams
	}{
		{
			name:       "multiple chat rooms with participants",
			want:       `[{"name":"chat-room-1-name","create_time":"0001-01-01T00:00:00Z","url":"aim:gochat?roomname=chat-room-1-name\u0026exchange=5","participants":[{"id":"usera","screen_name":"userA"},{"id":"userb","screen_name":"userB"}]},{"name":"chat-room-2-name","create_time":"0001-01-01T00:00:00Z","url":"aim:gochat?roomname=chat-room-2-name\u0026exchange=5","participants":[{"id":"userc","screen_name":"userC"},{"id":"userd","screen_name":"userD"}]}]`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				chatRoomRetrieverParams: chatRoomRetrieverParams{
					allChatRoomsParams: allChatRoomsParams{
						{
							exchange: state.PublicExchange,
							result: []state.ChatRoom{
								chatRoom1,
								chatRoom2,
							},
						},
					},
				},
				chatSessionRetrieverParams: chatSessionRetrieverParams{
					chatSessionRetrieverAllSessionsParams: chatSessionRetrieverAllSessionsParams{
						{
							cookie: chatRoom1.Cookie(),
							result: []*state.Session{
								fnNewSess("userA"),
								fnNewSess("userB"),
							},
						},
						{
							cookie: chatRoom2.Cookie(),
							result: []*state.Session{
								fnNewSess("userC"),
								fnNewSess("userD"),
							},
						},
					},
				},
			},
		},
		{
			name:       "chat room without participants",
			want:       `[{"name":"chat-room-1-name","create_time":"0001-01-01T00:00:00Z","url":"aim:gochat?roomname=chat-room-1-name\u0026exchange=5","participants":[]}]`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				chatRoomRetrieverParams: chatRoomRetrieverParams{
					allChatRoomsParams: allChatRoomsParams{
						{
							exchange: state.PublicExchange,
							result: []state.ChatRoom{
								chatRoom1,
							},
						},
					},
				},
				chatSessionRetrieverParams: chatSessionRetrieverParams{
					chatSessionRetrieverAllSessionsParams: chatSessionRetrieverAllSessionsParams{
						{
							cookie: chatRoom1.Cookie(),
							result: []*state.Session{},
						},
					},
				},
			},
		},
		{
			name:       "no chat rooms",
			want:       `[]`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				chatRoomRetrieverParams: chatRoomRetrieverParams{
					allChatRoomsParams: allChatRoomsParams{
						{
							exchange: state.PublicExchange,
							result:   []state.ChatRoom{},
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/chat/room/public", nil)
			responseRecorder := httptest.NewRecorder()

			chatRoomRetriever := newMockChatRoomRetriever(t)
			for _, params := range tc.mockParams.chatRoomRetrieverParams.allChatRoomsParams {
				chatRoomRetriever.EXPECT().
					AllChatRooms(params.exchange).
					Return(params.result, params.err)
			}

			chatSessionRetriever := newMockChatSessionRetriever(t)
			for _, params := range tc.mockParams.chatSessionRetrieverParams.chatSessionRetrieverAllSessionsParams {
				chatSessionRetriever.EXPECT().
					AllSessions(params.cookie).
					Return(params.result)
			}

			getPublicChatHandler(responseRecorder, request, chatRoomRetriever, chatSessionRetriever, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestPrivateChatHandler_GET(t *testing.T) {
	fnNewSess := func(screenName string) *state.Session {
		sess := state.NewSession()
		sess.SetIdentScreenName(state.NewIdentScreenName(screenName))
		sess.SetDisplayScreenName(state.DisplayScreenName(screenName))
		return sess
	}

	chatRoom1 := state.NewChatRoom("chat-room-1-name", state.NewIdentScreenName("chat-room-1-creator"), state.PrivateExchange)
	chatRoom2 := state.NewChatRoom("chat-room-2-name", state.NewIdentScreenName("chat-room-2-creator"), state.PrivateExchange)

	tt := []struct {
		name       string
		want       string
		statusCode int
		mockParams mockParams
	}{
		{
			name:       "multiple chat rooms with participants",
			want:       `[{"name":"chat-room-1-name","create_time":"0001-01-01T00:00:00Z","creator_id":"chat-room-1-creator","url":"aim:gochat?roomname=chat-room-1-name\u0026exchange=4","participants":[{"id":"usera","screen_name":"userA"},{"id":"userb","screen_name":"userB"}]},{"name":"chat-room-2-name","create_time":"0001-01-01T00:00:00Z","creator_id":"chat-room-2-creator","url":"aim:gochat?roomname=chat-room-2-name\u0026exchange=4","participants":[{"id":"userc","screen_name":"userC"},{"id":"userd","screen_name":"userD"}]}]`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				chatRoomRetrieverParams: chatRoomRetrieverParams{
					allChatRoomsParams: allChatRoomsParams{
						{
							exchange: state.PrivateExchange,
							result: []state.ChatRoom{
								chatRoom1,
								chatRoom2,
							},
						},
					},
				},
				chatSessionRetrieverParams: chatSessionRetrieverParams{
					chatSessionRetrieverAllSessionsParams: chatSessionRetrieverAllSessionsParams{
						{
							cookie: chatRoom1.Cookie(),
							result: []*state.Session{
								fnNewSess("userA"),
								fnNewSess("userB"),
							},
						},
						{
							cookie: chatRoom2.Cookie(),
							result: []*state.Session{
								fnNewSess("userC"),
								fnNewSess("userD"),
							},
						},
					},
				},
			},
		},
		{
			name:       "chat room without participants",
			want:       `[{"name":"chat-room-1-name","create_time":"0001-01-01T00:00:00Z","creator_id":"chat-room-1-creator","url":"aim:gochat?roomname=chat-room-1-name\u0026exchange=4","participants":[]}]`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				chatRoomRetrieverParams: chatRoomRetrieverParams{
					allChatRoomsParams: allChatRoomsParams{
						{
							exchange: state.PrivateExchange,
							result: []state.ChatRoom{
								chatRoom1,
							},
						},
					},
				},
				chatSessionRetrieverParams: chatSessionRetrieverParams{
					chatSessionRetrieverAllSessionsParams: chatSessionRetrieverAllSessionsParams{
						{
							cookie: chatRoom1.Cookie(),
							result: []*state.Session{},
						},
					},
				},
			},
		},
		{
			name:       "no chat rooms",
			want:       `[]`,
			statusCode: http.StatusOK,
			mockParams: mockParams{
				chatRoomRetrieverParams: chatRoomRetrieverParams{
					allChatRoomsParams: allChatRoomsParams{
						{
							exchange: state.PrivateExchange,
							result:   []state.ChatRoom{},
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/chat/room/private", nil)
			responseRecorder := httptest.NewRecorder()

			chatRoomRetriever := newMockChatRoomRetriever(t)
			for _, params := range tc.mockParams.chatRoomRetrieverParams.allChatRoomsParams {
				chatRoomRetriever.EXPECT().
					AllChatRooms(params.exchange).
					Return(params.result, params.err)
			}

			chatSessionRetriever := newMockChatSessionRetriever(t)
			for _, params := range tc.mockParams.chatSessionRetrieverParams.chatSessionRetrieverAllSessionsParams {
				chatSessionRetriever.EXPECT().
					AllSessions(params.cookie).
					Return(params.result)
			}

			getPrivateChatHandler(responseRecorder, request, chatRoomRetriever, chatSessionRetriever, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestInstantMessageHandler_POST(t *testing.T) {
	type relayToScreenNameInputs struct {
		sender    state.IdentScreenName
		recipient state.IdentScreenName
		msg       string
	}

	tt := []struct {
		name                    string
		relayToScreenNameInputs []relayToScreenNameInputs
		body                    string
		want                    string
		statusCode              int
	}{
		{
			name: "send an instant message",
			relayToScreenNameInputs: []relayToScreenNameInputs{
				{
					sender:    state.NewIdentScreenName("sender_sn"),
					recipient: state.NewIdentScreenName("recip_sn"),
					msg:       "hello world!",
				},
			},
			body:       `{"from":"sender_sn","to":"recip_sn","text":"hello world!"}`,
			want:       `Message sent successfully.`,
			statusCode: http.StatusOK,
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA", "password":"thepassword"`,
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			messageRelayer := newMockMessageRelayer(t)

			for _, params := range tc.relayToScreenNameInputs {
				validateSNAC := func(msg wire.SNACMessage) bool {
					body := msg.Body.(wire.SNAC_0x04_0x07_ICBMChannelMsgToClient)
					assert.Equal(t, params.sender.String(), body.TLVUserInfo.ScreenName)

					b, ok := body.Slice(wire.ICBMTLVAOLIMData)
					assert.True(t, ok)

					txt, err := wire.UnmarshalICBMMessageText(b)
					assert.NoError(t, err)
					assert.Equal(t, params.msg, txt)
					return true
				}
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.recipient, mock.MatchedBy(validateSNAC))
			}

			postInstantMessageHandler(responseRecorder, request, messageRelayer, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}
