package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/google/uuid"

	"github.com/mk6i/retro-aim-server/state"
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
}

type SessionRetriever interface {
	AllSessions() []*state.Session
}

func StartManagementAPI(userManager UserManager, sessionRetriever SessionRetriever, logger *slog.Logger) {
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
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		sessionHandler(w, r, sessionRetriever)
	})

	//todo make port configurable
	addr := net.JoinHostPort("", "8080")
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
