package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
	fnNewSess := func(screenName string) *state.Session {
		sess := state.NewSession()
		sess.SetIdentScreenName(state.NewIdentScreenName(screenName))
		sess.SetDisplayScreenName(state.DisplayScreenName(screenName))
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
				fnNewSess("userA"),
				fnNewSess("userB"),
			},
			want:       `{"count":2,"sessions":[{"screen_name":"userA"},{"screen_name":"userB"}]}`,
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
			},
			want:       `[{"screen_name":"userA"},{"screen_name":"userB"}]`,
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
	tt := []struct {
		name           string
		body           string
		UUID           uuid.UUID
		user           state.User
		userHandlerErr error
		want           string
		statusCode     int
	}{
		{
			name: "with valid user",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			user: func() state.User {
				user := state.User{
					AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
					DisplayScreenName: "userA",
					IdentScreenName:   state.NewIdentScreenName("userA"),
				}
				assert.NoError(t, user.HashPassword("thepassword"))
				return user
			}(),
			want:       `User account created successfully.`,
			statusCode: http.StatusCreated,
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA", "password":"thepassword"`,
			user:       state.User{},
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "user handler error",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			user: func() state.User {
				user := state.User{
					AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
					DisplayScreenName: "userA",
					IdentScreenName:   state.NewIdentScreenName("userA"),
				}
				assert.NoError(t, user.HashPassword("thepassword"))
				return user
			}(),
			userHandlerErr: io.EOF,
			want:           `internal server error`,
			statusCode:     http.StatusInternalServerError,
		},
		{
			name: "duplicate user",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			user: func() state.User {
				user := state.User{
					AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
					DisplayScreenName: "userA",
					IdentScreenName:   state.NewIdentScreenName("userA"),
				}
				assert.NoError(t, user.HashPassword("thepassword"))
				return user
			}(),
			userHandlerErr: state.ErrDupUser,
			want:           `user already exists`,
			statusCode:     http.StatusConflict,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			if tc.user.IdentScreenName.String() != "" {
				userManager.EXPECT().
					InsertUser(tc.user).
					Return(tc.userHandlerErr)
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
	tt := []struct {
		name           string
		body           string
		user           state.User
		UUID           uuid.UUID
		userHandlerErr error
		want           string
		statusCode     int
	}{
		{
			name: "with valid password",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			user: func() state.User {
				user := state.User{
					AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
					DisplayScreenName: "userA",
					IdentScreenName:   state.NewIdentScreenName("userA"),
				}
				assert.NoError(t, user.HashPassword("thepassword"))
				return user
			}(),
			want:       ``,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "with malformed body",
			body:       `{"screen_name":"userA", "password":"thepassword"`,
			user:       state.User{},
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "user password handler error",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			user: func() state.User {
				user := state.User{
					AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
					DisplayScreenName: "userA",
					IdentScreenName:   state.NewIdentScreenName("userA"),
				}
				assert.NoError(t, user.HashPassword("thepassword"))
				return user
			}(),
			userHandlerErr: io.EOF,
			want:           `internal server error`,
			statusCode:     http.StatusInternalServerError,
		},
		{
			name: "user doesn't exist",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			UUID: uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b"),
			user: func() state.User {
				user := state.User{
					AuthKey:           uuid.MustParse("07c70701-ba68-49a9-9f9b-67a53816e37b").String(),
					DisplayScreenName: "userA",
					IdentScreenName:   state.NewIdentScreenName("userA"),
				}
				assert.NoError(t, user.HashPassword("thepassword"))
				return user
			}(),
			userHandlerErr: state.ErrNoUser,
			want:           `user does not exist`,
			statusCode:     http.StatusNotFound,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/user", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			if tc.user.IdentScreenName.String() != "" {
				userManager.EXPECT().
					SetUserPassword(tc.user).
					Return(tc.userHandlerErr)
			}

			newUUID := func() uuid.UUID { return tc.UUID }
			userPasswordHandler(responseRecorder, request, userManager, newUUID, slog.Default())

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
					{
						Cookie:     "chat-room-1-cookie",
						Creator:    state.NewIdentScreenName("chat-room-1-creator"),
						Name:       "chat-room-1-name",
						CreateTime: time.Date(2024, 06, 01, 1, 2, 3, 4, time.UTC),
					},
					{
						Cookie:     "chat-room-2-cookie",
						Creator:    state.NewIdentScreenName("chat-room-2-creator"),
						Name:       "chat-room-2-name",
						CreateTime: time.Date(2022, 01, 04, 6, 8, 1, 2, time.UTC),
					},
				},
			},
			allSessionsParams: []allSessionsParams{
				{
					cookie: "chat-room-1-cookie",
					result: []*state.Session{
						fnNewSess("userA"),
						fnNewSess("userB"),
					},
				},
				{
					cookie: "chat-room-2-cookie",
					result: []*state.Session{
						fnNewSess("userC"),
						fnNewSess("userD"),
					},
				},
			},
			want:       `[{"name":"chat-room-1-name","create_time":"2024-06-01T01:02:03.000000004Z","url":"aim:gochat?exchange=0\u0026roomname=chat-room-1-name","participants":[{"id":"usera","screen_name":"userA"},{"id":"userb","screen_name":"userB"}]},{"name":"chat-room-2-name","create_time":"2022-01-04T06:08:01.000000002Z","url":"aim:gochat?exchange=0\u0026roomname=chat-room-2-name","participants":[{"id":"userc","screen_name":"userC"},{"id":"userd","screen_name":"userD"}]}]`,
			statusCode: http.StatusOK,
		},
		{
			name: "chat room without participants",
			allChatRoomsParams: allChatRoomsParams{
				exchange: state.PublicExchange,
				result: []state.ChatRoom{
					{
						Cookie:     "chat-room-1-cookie",
						Creator:    state.NewIdentScreenName("chat-room-1-creator"),
						Name:       "chat-room-1-name",
						CreateTime: time.Date(2024, 06, 01, 1, 2, 3, 4, time.UTC),
					},
				},
			},
			allSessionsParams: []allSessionsParams{
				{
					cookie: "chat-room-1-cookie",
					result: []*state.Session{},
				},
			},
			want:       `[{"name":"chat-room-1-name","create_time":"2024-06-01T01:02:03.000000004Z","url":"aim:gochat?exchange=0\u0026roomname=chat-room-1-name","participants":[]}]`,
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
					{
						Cookie:     "chat-room-1-cookie",
						Creator:    state.NewIdentScreenName("chat-room-1-creator"),
						Name:       "chat-room-1-name",
						CreateTime: time.Date(2024, 06, 01, 1, 2, 3, 4, time.UTC),
					},
					{
						Cookie:     "chat-room-2-cookie",
						Creator:    state.NewIdentScreenName("chat-room-2-creator"),
						Name:       "chat-room-2-name",
						CreateTime: time.Date(2022, 01, 04, 6, 8, 1, 2, time.UTC),
					},
				},
			},
			allSessionsParams: []allSessionsParams{
				{
					cookie: "chat-room-1-cookie",
					result: []*state.Session{
						fnNewSess("userA"),
						fnNewSess("userB"),
					},
				},
				{
					cookie: "chat-room-2-cookie",
					result: []*state.Session{
						fnNewSess("userC"),
						fnNewSess("userD"),
					},
				},
			},
			want:       `[{"name":"chat-room-1-name","create_time":"2024-06-01T01:02:03.000000004Z","creator_id":"chat-room-1-creator","url":"aim:gochat?exchange=0\u0026roomname=chat-room-1-name","participants":[{"id":"usera","screen_name":"userA"},{"id":"userb","screen_name":"userB"}]},{"name":"chat-room-2-name","create_time":"2022-01-04T06:08:01.000000002Z","creator_id":"chat-room-2-creator","url":"aim:gochat?exchange=0\u0026roomname=chat-room-2-name","participants":[{"id":"userc","screen_name":"userC"},{"id":"userd","screen_name":"userD"}]}]`,
			statusCode: http.StatusOK,
		},
		{
			name: "chat room without participants",
			allChatRoomsParams: allChatRoomsParams{
				exchange: state.PrivateExchange,
				result: []state.ChatRoom{
					{
						Cookie:     "chat-room-1-cookie",
						Creator:    state.NewIdentScreenName("chat-room-1-creator"),
						Name:       "chat-room-1-name",
						CreateTime: time.Date(2024, 06, 01, 1, 2, 3, 4, time.UTC),
					},
				},
			},
			allSessionsParams: []allSessionsParams{
				{
					cookie: "chat-room-1-cookie",
					result: []*state.Session{},
				},
			},
			want:       `[{"name":"chat-room-1-name","create_time":"2024-06-01T01:02:03.000000004Z","creator_id":"chat-room-1-creator","url":"aim:gochat?exchange=0\u0026roomname=chat-room-1-name","participants":[]}]`,
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
