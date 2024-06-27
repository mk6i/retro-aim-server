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

func StartManagementAPI(
	cfg config.Config,
	userManager UserManager,
	sessionRetriever SessionRetriever,
	chatRoomRetriever ChatRoomRetriever,
	chatRoomCreator ChatRoomCreator,
	chatSessionRetriever ChatSessionRetriever,
	logger *slog.Logger,
) {

	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r, userManager, uuid.New, logger)
	})
	mux.HandleFunc("/user/password", func(w http.ResponseWriter, r *http.Request) {
		userPasswordHandler(w, r, userManager, uuid.New, logger)
	})
	mux.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, userManager, logger)
	})
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		sessionHandler(w, r, sessionRetriever)
	})
	mux.HandleFunc("/chat/room/public", func(w http.ResponseWriter, r *http.Request) {
		publicChatHandler(w, r, chatRoomRetriever, chatRoomCreator, chatSessionRetriever, state.NewChatRoom, logger)
	})
	mux.HandleFunc("/chat/room/private", func(w http.ResponseWriter, r *http.Request) {
		privateChatHandler(w, r, chatRoomRetriever, chatSessionRetriever, logger)
	})

	addr := net.JoinHostPort(cfg.ApiHost, cfg.ApiPort)
	logger.Info("starting management API server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("unable to bind management API address address", "err", err.Error())
		os.Exit(1)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, newUUID func() uuid.UUID, logger *slog.Logger) {
	switch r.Method {
	case http.MethodDelete:
		deleteUserHandler(w, r, userManager, logger)
	case http.MethodGet:
		getUserHandler(w, r, userManager, logger)
	case http.MethodPost:
		postUserHandler(w, r, userManager, newUUID, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request, manager UserManager, logger *slog.Logger) {
	user, err := userFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = manager.DeleteUser(user.DisplayScreenName.IdentScreenName())
	switch {
	case errors.Is(err, state.ErrNoUser):
		http.Error(w, "user does not exist", http.StatusNotFound)
		return
	case err != nil:
		logger.Error("error deleting user DELETE /user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	fmt.Fprintln(w, "User account successfully deleted.")
}

func userPasswordHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, newUUID func() uuid.UUID, logger *slog.Logger) {
	switch r.Method {
	case http.MethodPut:
		putUserPasswordHandler(w, r, userManager, newUUID, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// putUserPasswordHandler handles the PUT /user/password endpoint.
func putUserPasswordHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, newUUID func() uuid.UUID, logger *slog.Logger) {
	user, err := userFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user.AuthKey = newUUID().String()

	user.IdentScreenName = user.DisplayScreenName.IdentScreenName()
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
			ScreenName: s.DisplayScreenName().String(),
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
func postUserHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, newUUID func() uuid.UUID, logger *slog.Logger) {
	user, err := userFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user.AuthKey = newUUID().String()

	if err := user.HashPassword(user.Password); err != nil {
		logger.Error("error hashing user password in POST /user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = userManager.InsertUser(user.User)
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

func userFromBody(r *http.Request) (userWithPassword, error) {
	user := userWithPassword{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return userWithPassword{}, errors.New("malformed input")
	}
	user.IdentScreenName = user.DisplayScreenName.IdentScreenName()
	return user, nil
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

	username, password := state.NewIdentScreenName(pair[0]), pair[1]

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

func publicChatHandler(w http.ResponseWriter, r *http.Request, chatRoomRetriever ChatRoomRetriever, chatRoomCreator ChatRoomCreator, chatSessionRetriever ChatSessionRetriever, newChatRoom func() state.ChatRoom, logger *slog.Logger) {
	switch r.Method {
	case http.MethodGet:
		getPublicChatHandler(w, r, chatRoomRetriever, chatSessionRetriever, logger)
	case http.MethodPost:
		postPublicChatHandler(w, r, chatRoomCreator, newChatRoom, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func privateChatHandler(w http.ResponseWriter, r *http.Request, chatRoomRetriever ChatRoomRetriever, chatSessionRetriever ChatSessionRetriever, logger *slog.Logger) {
	switch r.Method {
	case http.MethodGet:
		getPrivateChatHandler(w, r, chatRoomRetriever, chatSessionRetriever, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// getPublicChatHandler handles the GET /chat/room/public endpoint.
func getPublicChatHandler(w http.ResponseWriter, _ *http.Request, chatRoomRetriever ChatRoomRetriever, chatSessionRetriever ChatSessionRetriever, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	rooms, err := chatRoomRetriever.AllChatRooms(state.PublicExchange)
	if err != nil {
		logger.Error("error in GET /chat/rooms/public", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]chatRoom, len(rooms))
	for i, room := range rooms {
		sessions := chatSessionRetriever.AllSessions(room.Cookie)
		cr := chatRoom{
			CreateTime:   room.CreateTime,
			Name:         room.Name,
			Participants: make([]userHandle, 0, len(sessions)),
			URL:          room.URL().String(),
		}
		for _, sess := range sessions {
			cr.Participants = append(cr.Participants, userHandle{
				ID:         sess.IdentScreenName().String(),
				ScreenName: sess.DisplayScreenName().String(),
			})
		}

		out[i] = cr
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// postPublicChatHandler handles the POST /chat/room/public endpoint.
func postPublicChatHandler(w http.ResponseWriter, r *http.Request, chatRoomCreator ChatRoomCreator, newChatRoom func() state.ChatRoom, logger *slog.Logger) {
	input := chatRoomCreate{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 50 {
		http.Error(w, "chat room name must be between 1 and 50 characters", http.StatusBadRequest)
		return
	}

	cr := newChatRoom()
	cr.Name = input.Name
	cr.Exchange = state.PublicExchange

	err := chatRoomCreator.CreateChatRoom(cr)
	switch {
	case errors.Is(err, state.ErrDupChatRoom):
		http.Error(w, "Chat room already exists.", http.StatusConflict)
		return
	case err != nil:
		logger.Error("error inserting chat room POST /chat/room/public", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Chat room created successfully.")
}

// getPrivateChatHandler handles the GET /chat/room/private endpoint.
func getPrivateChatHandler(w http.ResponseWriter, _ *http.Request, chatRoomRetriever ChatRoomRetriever, chatSessionRetriever ChatSessionRetriever, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	rooms, err := chatRoomRetriever.AllChatRooms(state.PrivateExchange)
	if err != nil {
		logger.Error("error in GET /chat/rooms/private", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]chatRoom, len(rooms))
	for i, room := range rooms {
		sessions := chatSessionRetriever.AllSessions(room.Cookie)
		cr := chatRoom{
			CreateTime:   room.CreateTime,
			CreatorID:    room.Creator.String(),
			Name:         room.Name,
			Participants: make([]userHandle, 0, len(sessions)),
			URL:          room.URL().String(),
		}
		for _, sess := range sessions {
			cr.Participants = append(cr.Participants, userHandle{
				ID:         sess.IdentScreenName().String(),
				ScreenName: sess.DisplayScreenName().String(),
			})
		}

		out[i] = cr
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
