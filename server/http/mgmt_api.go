package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
)

type createUser struct {
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
}

type SessionRetriever interface {
	AllSessions() []*state.Session
}

func StartManagementAPI(userManager UserManager, sessionRetriever SessionRetriever, logger *slog.Logger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r, userManager, logger)
	})
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		sessionHandler(w, r, sessionRetriever)
	})

	//todo make port configurable
	addr := config.Address("", 8080)
	logger.Info("starting management API server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("unable to bind management API address address", "err", err.Error())
		os.Exit(1)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, logger *slog.Logger) {
	switch r.Method {
	case http.MethodGet:
		getUserHandler(w, r, userManager, logger)
	case http.MethodPost:
		postUserHandler(w, r, userManager, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
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
func postUserHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, logger *slog.Logger) {
	var newUser createUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "malformed input", http.StatusBadRequest)
		return
	}
	newUser.AuthKey = uuid.New().String()
	// todo does the request contain authkey?
	newUser.HashPassword(newUser.Password)
	if err := userManager.InsertUser(newUser.User); err != nil {
		logger.Error("error in GET /user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User account created successfully.")
}
