package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/state"
)

type UserManager interface {
	AllUsers() ([]state.User, error)
	InsertUser(u state.User) error
}

func StartManagementAPI(userManager UserManager, logger *slog.Logger) {
	uh := userHandler{
		UserManager: userManager,
		logger:      logger,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/user", uh.ServeHTTP)

	//todo make port configurable
	addr := Address("", 8080)
	logger.Info("starting management API server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("unable to bind management API address address", "err", err.Error())
		os.Exit(1)
	}
}

type userHandler struct {
	UserManager
	logger *slog.Logger
}

func (uh userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		uh.getUsers(w, r)
	case http.MethodPost:
		uh.createUser(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// getUsers handles the GET /user endpoint.
func (uh userHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := uh.AllUsers()
	if err != nil {
		uh.logger.Error("error in GET /user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type CreateUser struct {
	state.User
	Password string `json:"password,omitempty"`
}

// createUser handles the POST /user endpoint.
func (uh userHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var newUser CreateUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "malformed input", http.StatusBadRequest)
		return
	}
	newUser.AuthKey = uuid.New().String()
	// todo does the request contain authkey?
	newUser.HashPassword(newUser.Password)
	if err := uh.InsertUser(newUser.User); err != nil {
		uh.logger.Error("error in GET /user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User account created successfully.")
}
