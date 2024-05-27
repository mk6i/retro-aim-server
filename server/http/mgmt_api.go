package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type userWithPassword struct {
	state.User
	Password string `json:"password,omitempty"`
}

type userSession struct {
	ScreenName string `json:"screen_name"`
}

type onlineUsers struct {
	Count    int           `json:"count"`
	Sessions []userSession `json:"sessions"`
}

type UserManager interface {
	AllUsers() ([]state.User, error)
	InsertUser(u state.User) error
	SetUserPassword(u state.User) error
	User(screenName string) (*state.User, error)
}

type SessionRetriever interface {
	AllSessions() []*state.Session
}

func StartManagementAPI(cfg config.Config, userManager UserManager, sessionRetriever SessionRetriever, logger *slog.Logger) {
	mux := http.NewServeMux()
	newUser := func() state.User {
		return state.User{AuthKey: uuid.New().String()}
	}
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r, userManager, newUser, logger)
	})
	mux.HandleFunc("/user/password", func(w http.ResponseWriter, r *http.Request) {
		userPasswordHandler(w, r, userManager, newUser, logger)
	})
	mux.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, userManager, logger)
	})
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		sessionHandler(w, r, sessionRetriever)
	})

	addr := net.JoinHostPort(cfg.ApiHost, cfg.ApiPort)
	logger.Info("starting management API server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("unable to bind management API address address", "err", err.Error())
		os.Exit(1)
	}
}

func userHandler(
	w http.ResponseWriter,
	r *http.Request,
	userManager UserManager,
	newUser func() state.User,
	logger *slog.Logger,
) {
	switch r.Method {
	case http.MethodGet:
		getUserHandler(w, r, userManager, logger)
	case http.MethodPost:
		postUserHandler(w, r, userManager, newUser, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func userPasswordHandler(
	w http.ResponseWriter,
	r *http.Request,
	userManager UserManager,
	userFactory func() state.User,
	logger *slog.Logger,
) {
	switch r.Method {
	case http.MethodPut:
		putUserPasswordHandler(w, r, userManager, userFactory, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// putUserPasswordHandler handles the PUT /user/password endpoint.
func putUserPasswordHandler(
	w http.ResponseWriter,
	r *http.Request,
	userManager UserManager,
	newUser func() state.User,
	logger *slog.Logger,
) {
	user := userWithPassword{
		User: newUser(),
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "malformed input", http.StatusBadRequest)
		return
	}
	if err := user.HashPassword(user.Password); err != nil {
		logger.Error("error hashing user password in PUT /user/password", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := userManager.SetUserPassword(user.User); err != nil {
		switch {
		case errors.Is(err, state.ErrNoUser):
			http.Error(w, "user does not exist", http.StatusNotFound)
			return
		case err != nil:
			logger.Error("error updating user password PUT /user/password", "err", err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// sessionHandler handles GET /session
func sessionHandler(w http.ResponseWriter, r *http.Request, sessionRetriever SessionRetriever) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	allUsers := sessionRetriever.AllSessions()

	ou := onlineUsers{
		Count:    len(allUsers),
		Sessions: make([]userSession, 0),
	}

	for _, s := range allUsers {
		ou.Sessions = append(ou.Sessions, userSession{
			ScreenName: s.ScreenName(),
		})
	}

	if err := json.NewEncoder(w).Encode(ou); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getUserHandler handles the GET /user endpoint.
func getUserHandler(w http.ResponseWriter, _ *http.Request, userManager UserManager, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	users, err := userManager.AllUsers()
	if err != nil {
		logger.Error("error in GET /user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// postUserHandler handles the POST /user endpoint.
func postUserHandler(
	w http.ResponseWriter,
	r *http.Request,
	userManager UserManager,
	newUser func() state.User,
	logger *slog.Logger,
) {
	user := userWithPassword{
		User: newUser(),
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "malformed input", http.StatusBadRequest)
		return
	}
	if err := user.HashPassword(user.Password); err != nil {
		logger.Error("error hashing user password in POST /user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err := userManager.InsertUser(user.User)
	switch {
	case errors.Is(err, state.ErrDupUser):
		http.Error(w, "user already exists", http.StatusConflict)
		return
	case err != nil:
		logger.Error("error inserting user POST /user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User account created successfully.")
}

// loginHandler is a temporary endpoint for validating user credentials for
// chivanet. do not rely on this endpoint, as it will be eventually removed.
func loginHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, logger *slog.Logger) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// No authentication header found
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", `Basic realm="User Login"`)
		w.Write([]byte("401 Unauthorized\n"))
		return
	}

	auth := strings.SplitN(authHeader, " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized: Missing Basic prefix\n"))
		return
	}

	payload, err := base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized: Invalid Base64 Encoding\n"))
		return
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized: Invalid Authentication Token\n"))
		return
	}

	username, password := pair[0], pair[1]

	user, err := userManager.User(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 InternalServerError\n"))
		logger.Error("error getting user", "err", err.Error())
		return
	}
	if user == nil || !user.ValidateHash(wire.StrongMD5PasswordHash(password, user.AuthKey)) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized: Invalid Credentials\n"))
		return
	}

	// Successfully authenticated
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK: Successfully Authenticated\n"))
}
