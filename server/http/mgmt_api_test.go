package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mk6i/retro-aim-server/state"

	"github.com/stretchr/testify/mock"
)

func TestUserHandler_GET(t *testing.T) {
	tt := []struct {
		name           string
		users          []state.User
		userHandlerErr error
		want           string
		statusCode     int
	}{
		{
			name:       "without users",
			users:      []state.User{},
			want:       `[]`,
			statusCode: http.StatusOK,
		},
		{
			name: "with users",
			users: []state.User{
				{ScreenName: "userA"},
				{ScreenName: "userB"},
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
			request := httptest.NewRequest(http.MethodGet, "/users", nil)
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			userManager.EXPECT().
				AllUsers().
				Return(tc.users, tc.userHandlerErr)

			userHandler := userHandler{
				UserManager: userManager,
				logger:      slog.Default(),
			}
			userHandler.ServeHTTP(responseRecorder, request)

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
		user           state.User
		userHandlerErr error
		want           string
		statusCode     int
	}{
		{
			name: "with valid user",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			user: state.User{
				ScreenName: "userA",
			},
			want:       `User account created successfully.`,
			statusCode: http.StatusCreated,
		},
		{
			name: "with malformed body",
			body: `{"screen_name":"userA", "password":"thepassword"`,
			user: state.User{
				ScreenName: "userA",
			},
			want:       `malformed input`,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "user handler error",
			body: `{"screen_name":"userA", "password":"thepassword"}`,
			user: state.User{
				ScreenName: "userA",
			},
			userHandlerErr: io.EOF,
			want:           `internal server error`,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			userManager := newMockUserManager(t)
			userManager.EXPECT().
				InsertUser(mock.Anything). // todo make this more concrete
				Return(tc.userHandlerErr).
				Maybe()

			userHandler := userHandler{
				UserManager: userManager,
				logger:      slog.Default(),
			}
			userHandler.ServeHTTP(responseRecorder, request)

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
	request := httptest.NewRequest(http.MethodPut, "/users", nil)
	responseRecorder := httptest.NewRecorder()

	userHandler := userHandler{}
	userHandler.ServeHTTP(responseRecorder, request)

	wantCode := http.StatusMethodNotAllowed
	if responseRecorder.Code != wantCode {
		t.Errorf("want status '%d', got '%d'", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	wantBody := `method not allowed`
	if strings.TrimSpace(responseRecorder.Body.String()) != wantBody {
		t.Errorf("want '%s', got '%s'", wantBody, responseRecorder.Body)
	}
}
