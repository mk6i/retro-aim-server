package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		name           string
		sessions       []*state.Session
		userHandlerErr error
		want           string
		statusCode     int
	}{
		{
			name:       "without sessions",
			sessions:   []*state.Session{},
			want:       `{"count":0,"sessions":[]}`,
			statusCode: http.StatusOK,
		},
		{
			name: "with sessions",
			sessions: []*state.Session{
				fnNewSess("userA", 0),
				fnNewSess("userB", 0),
				fnNewSess("100003", 100003),
			},
			want:       `{"count":3,"sessions":[{"id":"usera","screen_name":"userA","is_icq":false},{"id":"userb","screen_name":"userB","is_icq":false},{"id":"100003","screen_name":"100003","is_icq":true}]}`,
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/session", nil)
			responseRecorder := httptest.NewRecorder()

			sessionRetriever := newMockSessionRetriever(t)
			sessionRetriever.EXPECT().
				AllSessions().
				Return(tc.sessions)

			sessionHandler(responseRecorder, request, sessionRetriever)

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestSessionHandler_DisallowedMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/session", nil)
	responseRecorder := httptest.NewRecorder()

	sessionHandler(responseRecorder, request, nil)

	wantCode := http.StatusMethodNotAllowed
	if responseRecorder.Code != wantCode {
		t.Errorf("want status '%d', got '%d'", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	wantBody := `method not allowed`
	if strings.TrimSpace(responseRecorder.Body.String()) != wantBody {
		t.Errorf("want '%s', got '%s'", wantBody, responseRecorder.Body)
	}
}

func TestUserHandler_GET(t *testing.T) {
	tt := []struct {
		name           string
		users          []state.User
		userHandlerErr error
		want           string
		statusCode     int
	}{
		{
			name:       "empty user store",
			users:      []state.User{},
			want:       `[]`,
			statusCode: http.StatusOK,
		},
		{
			name: "user store containing 2 users",
			users: []state.User{
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
			want:       `[{"id":"usera","screen_name":"userA","is_icq":false},{"id":"userb","screen_name":"userB","is_icq":false},{"id":"100003","screen_name":"100003","is_icq":true}]`,
			statusCode: http.StatusOK,
		},
		{
			name:           "user handler error",
			users:          []state.User{},
			userHandlerErr: io.EOF,
			want:           `internal server error`,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/user", nil)
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			userManager.EXPECT().
				AllUsers().
				Return(tc.users, tc.userHandlerErr)

			userHandler(responseRecorder, request, userManager, nil, slog.Default())

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
	type insertUserParams struct {
		user state.User
		err  error
	}
	tt := []struct {
		name             string
		body             string
		UUID             uuid.UUID
		insertUserParams []insertUserParams
		want             string
		statusCode       int
	}{
		{
			name: "with valid AIM user",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			insertUserParams: []insertUserParams{
				{
					user: func() state.User {
						user := state.User{
							AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
							DisplayScreenName: "userA",
							IdentScreenName:   state.NewIdentScreenName("userA"),
						}
						assert.NoError(t, user.HashPassword("thepassword"))
						return user
					}(),
				},
			},
			want:       `User account created successfully.`,
			statusCode: http.StatusCreated,
		},
		{
			name: "with valid ICQ user",
			body: `{"screen_name":"100003", "password":"thepass"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			insertUserParams: []insertUserParams{
				{
					user: func() state.User {
						user := state.User{
							AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
							DisplayScreenName: "100003",
							IdentScreenName:   state.NewIdentScreenName("100003"),
							IsICQ:             true,
						}
						assert.NoError(t, user.HashPassword("thepass"))
						return user
					}(),
				},
			},
			want:       `User account created successfully.`,
			statusCode: http.StatusCreated,
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA", "password":"thepassword"`,
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "user handler error",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			insertUserParams: []insertUserParams{
				{
					user: func() state.User {
						user := state.User{
							AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
							DisplayScreenName: "userA",
							IdentScreenName:   state.NewIdentScreenName("userA"),
						}
						assert.NoError(t, user.HashPassword("thepassword"))
						return user
					}(),
					err: io.EOF,
				},
			},
			want:       `internal server error`,
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "duplicate user",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			insertUserParams: []insertUserParams{
				{
					user: func() state.User {
						user := state.User{
							AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
							DisplayScreenName: "userA",
							IdentScreenName:   state.NewIdentScreenName("userA"),
						}
						assert.NoError(t, user.HashPassword("thepassword"))
						return user
					}(),
					err: state.ErrDupUser,
				},
			},
			want:       `user already exists`,
			statusCode: http.StatusConflict,
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
			for _, params := range tc.insertUserParams {
				userManager.EXPECT().
					InsertUser(params.user).
					Return(params.err)
			}

			newUUID := func() uuid.UUID { return tc.UUID }
			userHandler(responseRecorder, request, userManager, newUUID, slog.Default())

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
		name           string
		body           string
		user           state.User
		userHandlerErr error
		want           string
		statusCode     int
	}{
		{
			name: "with valid user",
			body: `{"screen_name":"userA"}`,
			user: state.User{
				IdentScreenName: state.NewIdentScreenName("userA"),
			},
			want:       `User account successfully deleted.`,
			statusCode: http.StatusNoContent,
		},
		{
			name: "with non-existent user",
			body: `{"screen_name":"userA"}`,
			user: state.User{
				IdentScreenName: state.NewIdentScreenName("userA"),
			},
			userHandlerErr: state.ErrNoUser,
			want:           `user does not exist`,
			statusCode:     http.StatusNotFound,
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA"`,
			user:       state.User{},
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "user handler error",
			body: `{"screen_name":"userA"}`,
			user: state.User{
				IdentScreenName: state.NewIdentScreenName("userA"),
			},
			userHandlerErr: io.EOF,
			want:           `internal server error`,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/user", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			if tc.user.IdentScreenName.String() != "" {
				userManager.EXPECT().
					DeleteUser(tc.user.IdentScreenName).
					Return(tc.userHandlerErr)
			}

			userHandler(responseRecorder, request, userManager, nil, slog.Default())

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
	type setUserPasswordParams struct {
		screenName  state.IdentScreenName
		newPassword string
		wantErr     error
	}
	tt := []struct {
		name             string
		body             string
		updateUserParams []setUserPasswordParams
		want             string
		statusCode       int
	}{
		{
			name: "user with valid password",
			body: `{"screen_name":"userA", "password":"thenewpassword"}`,
			updateUserParams: []setUserPasswordParams{
				{
					screenName:  state.NewIdentScreenName("userA"),
					newPassword: "thenewpassword",
				},
			},
			want:       `Password successfully reset.`,
			statusCode: http.StatusNoContent,
		},
		{
			name: "user with invalid password",
			body: `{"screen_name":"userA", "password":"a"}`,
			updateUserParams: []setUserPasswordParams{
				{
					screenName:  state.NewIdentScreenName("userA"),
					newPassword: "a",
					wantErr:     state.ErrPasswordInvalid,
				},
			},
			want:       `invalid password length`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA", "password":"thepassword"`,
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "password updater returns runtime error",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			updateUserParams: []setUserPasswordParams{
				{
					screenName:  state.NewIdentScreenName("userA"),
					newPassword: "thepassword",
					wantErr:     io.EOF,
				},
			},
			want:       `internal server error`,
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "user doesn't exist",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			updateUserParams: []setUserPasswordParams{
				{
					screenName:  state.NewIdentScreenName("userA"),
					newPassword: "thepassword",
					wantErr:     state.ErrNoUser,
				},
			},
			want:       `user does not exist`,
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/user", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			for _, params := range tc.updateUserParams {
				userManager.EXPECT().
					SetUserPassword(params.screenName, params.newPassword).
					Return(params.wantErr)
			}

			userPasswordHandler(responseRecorder, request, userManager, slog.Default())

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

func TestUserHandler_DisallowedMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/user", nil)
	responseRecorder := httptest.NewRecorder()

	userHandler(responseRecorder, request, nil, nil, nil)

	wantCode := http.StatusMethodNotAllowed
	if responseRecorder.Code != wantCode {
		t.Errorf("want status '%d', got '%d'", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	wantBody := `method not allowed`
	if strings.TrimSpace(responseRecorder.Body.String()) != wantBody {
		t.Errorf("want '%s', got '%s'", wantBody, responseRecorder.Body)
	}
}

func TestPublicChatHandler_GET(t *testing.T) {
	fnNewSess := func(screenName string) *state.Session {
		sess := state.NewSession()
		sess.SetIdentScreenName(state.NewIdentScreenName(screenName))
		sess.SetDisplayScreenName(state.DisplayScreenName(screenName))
		return sess
	}
	type allChatRoomsParams struct {
		exchange uint16
		result   []state.ChatRoom
		err      error
	}
	type allSessionsParams struct {
		cookie string
		result []*state.Session
	}

	chatRoom1 := state.NewChatRoom("chat-room-1-name", state.NewIdentScreenName("chat-room-1-creator"), state.PublicExchange)
	chatRoom2 := state.NewChatRoom("chat-room-2-name", state.NewIdentScreenName("chat-room-1-creator"), state.PublicExchange)

	tt := []struct {
		name               string
		allChatRoomsParams allChatRoomsParams
		allSessionsParams  []allSessionsParams
		userHandlerErr     error
		want               string
		statusCode         int
	}{
		{
			name: "multiple chat rooms with participants",
			allChatRoomsParams: allChatRoomsParams{
				exchange: state.PublicExchange,
				result: []state.ChatRoom{
					chatRoom1,
					chatRoom2,
				},
			},
			allSessionsParams: []allSessionsParams{
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
			want:       `[{"name":"chat-room-1-name","create_time":"0001-01-01T00:00:00Z","url":"aim:gochat?exchange=5\u0026roomname=chat-room-1-name","participants":[{"id":"usera","screen_name":"userA"},{"id":"userb","screen_name":"userB"}]},{"name":"chat-room-2-name","create_time":"0001-01-01T00:00:00Z","url":"aim:gochat?exchange=5\u0026roomname=chat-room-2-name","participants":[{"id":"userc","screen_name":"userC"},{"id":"userd","screen_name":"userD"}]}]`,
			statusCode: http.StatusOK,
		},
		{
			name: "chat room without participants",
			allChatRoomsParams: allChatRoomsParams{
				exchange: state.PublicExchange,
				result: []state.ChatRoom{
					chatRoom1,
				},
			},
			allSessionsParams: []allSessionsParams{
				{
					cookie: chatRoom1.Cookie(),
					result: []*state.Session{},
				},
			},
			want:       `[{"name":"chat-room-1-name","create_time":"0001-01-01T00:00:00Z","url":"aim:gochat?exchange=5\u0026roomname=chat-room-1-name","participants":[]}]`,
			statusCode: http.StatusOK,
		},
		{
			name: "no chat rooms",
			allChatRoomsParams: allChatRoomsParams{
				exchange: state.PublicExchange,
				result:   []state.ChatRoom{},
			},
			allSessionsParams: []allSessionsParams{},
			want:              `[]`,
			statusCode:        http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/chat/room/public", nil)
			responseRecorder := httptest.NewRecorder()

			chatRoomRetriever := newMockChatRoomRetriever(t)
			chatRoomRetriever.EXPECT().
				AllChatRooms(tc.allChatRoomsParams.exchange).
				Return(tc.allChatRoomsParams.result, tc.allChatRoomsParams.err)

			chatSessionRetriever := newMockChatSessionRetriever(t)
			for _, params := range tc.allSessionsParams {
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
	type allChatRoomsParams struct {
		exchange uint16
		result   []state.ChatRoom
		err      error
	}
	type allSessionsParams struct {
		cookie string
		result []*state.Session
	}

	chatRoom1 := state.NewChatRoom("chat-room-1-name", state.NewIdentScreenName("chat-room-1-creator"), state.PrivateExchange)
	chatRoom2 := state.NewChatRoom("chat-room-2-name", state.NewIdentScreenName("chat-room-2-creator"), state.PrivateExchange)

	tt := []struct {
		name               string
		allChatRoomsParams allChatRoomsParams
		allSessionsParams  []allSessionsParams
		userHandlerErr     error
		want               string
		statusCode         int
	}{
		{
			name: "multiple chat rooms with participants",
			allChatRoomsParams: allChatRoomsParams{
				exchange: state.PrivateExchange,
				result: []state.ChatRoom{
					chatRoom1,
					chatRoom2,
				},
			},
			allSessionsParams: []allSessionsParams{
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
			want:       `[{"name":"chat-room-1-name","create_time":"0001-01-01T00:00:00Z","creator_id":"chat-room-1-creator","url":"aim:gochat?exchange=4\u0026roomname=chat-room-1-name","participants":[{"id":"usera","screen_name":"userA"},{"id":"userb","screen_name":"userB"}]},{"name":"chat-room-2-name","create_time":"0001-01-01T00:00:00Z","creator_id":"chat-room-2-creator","url":"aim:gochat?exchange=4\u0026roomname=chat-room-2-name","participants":[{"id":"userc","screen_name":"userC"},{"id":"userd","screen_name":"userD"}]}]`,
			statusCode: http.StatusOK,
		},
		{
			name: "chat room without participants",
			allChatRoomsParams: allChatRoomsParams{
				exchange: state.PrivateExchange,
				result: []state.ChatRoom{
					chatRoom1,
				},
			},
			allSessionsParams: []allSessionsParams{
				{
					cookie: chatRoom1.Cookie(),
					result: []*state.Session{},
				},
			},
			want:       `[{"name":"chat-room-1-name","create_time":"0001-01-01T00:00:00Z","creator_id":"chat-room-1-creator","url":"aim:gochat?exchange=4\u0026roomname=chat-room-1-name","participants":[]}]`,
			statusCode: http.StatusOK,
		},
		{
			name: "no chat rooms",
			allChatRoomsParams: allChatRoomsParams{
				exchange: state.PrivateExchange,
				result:   []state.ChatRoom{},
			},
			allSessionsParams: []allSessionsParams{},
			want:              `[]`,
			statusCode:        http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/chat/room/private", nil)
			responseRecorder := httptest.NewRecorder()

			chatRoomRetriever := newMockChatRoomRetriever(t)
			chatRoomRetriever.EXPECT().
				AllChatRooms(tc.allChatRoomsParams.exchange).
				Return(tc.allChatRoomsParams.result, tc.allChatRoomsParams.err)

			chatSessionRetriever := newMockChatSessionRetriever(t)
			for _, params := range tc.allSessionsParams {
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
	type relayToScreenNameParams struct {
		sender    state.IdentScreenName
		recipient state.IdentScreenName
		msg       string
	}

	tt := []struct {
		name                    string
		relayToScreenNameParams []relayToScreenNameParams
		body                    string
		want                    string
		statusCode              int
	}{
		{
			name: "send an instant message",
			relayToScreenNameParams: []relayToScreenNameParams{
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

			for _, params := range tc.relayToScreenNameParams {
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
