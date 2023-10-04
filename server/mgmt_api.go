package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func StartManagementAPI(fs *FeedbagStore) {
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
	port := 8080
	fmt.Printf("Server is running on :%d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

// getUsers handles the GET /user endpoint.
func getUsers(fs *FeedbagStore, w http.ResponseWriter, r *http.Request) {
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
	User
	Password string `json:"password,omitempty"`
}

// createUser handles the POST /user endpoint.
func createUser(fs *FeedbagStore, w http.ResponseWriter, r *http.Request) {
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
