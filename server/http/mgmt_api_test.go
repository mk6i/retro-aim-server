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

	"github.com/mk6i/retro-aim-server/state"
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
