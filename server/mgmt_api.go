package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/state"
)

func StartManagementAPI(fs *state.SQLiteFeedbagStore, logger *slog.Logger) {
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getUsers(fs, w, r)
		case http.MethodPost:
			createUser(fs, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	//todo make port configurable
	addr := Address("", 8080)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("unable to bind management API address address", "err", err.Error())
		os.Exit(1)
	}
	logger.Info("starting management API server", "addr", addr)
	if err := http.Serve(listener, nil); err != nil {
		logger.Info("unable to start management API server", "err", err.Error())
		os.Exit(1)
	}
}

// getUsers handles the GET /user endpoint.
func getUsers(fs *state.SQLiteFeedbagStore, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := fs.Users()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func createUser(fs *state.SQLiteFeedbagStore, w http.ResponseWriter, r *http.Request) {
	var newUser CreateUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newUser.AuthKey = uuid.New().String()
	// todo does the request contain authkey?
	newUser.HashPassword(newUser.Password)
	if err := fs.InsertUser(newUser.User); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User account created successfully.")
}
