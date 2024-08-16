package http

import (
	"context"
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
	messageRelayer MessageRelayer,
	logger *slog.Logger,
) {

	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r, userManager, uuid.New, logger)
	})
	mux.HandleFunc("/user/password", func(w http.ResponseWriter, r *http.Request) {
		userPasswordHandler(w, r, userManager, logger)
	})
	mux.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, userManager, logger)
	})
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		sessionHandler(w, r, sessionRetriever)
	})
	mux.HandleFunc("/chat/room/public", func(w http.ResponseWriter, r *http.Request) {
		publicChatHandler(w, r, chatRoomRetriever, chatRoomCreator, chatSessionRetriever, logger)
	})
	mux.HandleFunc("/chat/room/private", func(w http.ResponseWriter, r *http.Request) {
		privateChatHandler(w, r, chatRoomRetriever, chatSessionRetriever, logger)
	})
	mux.HandleFunc("/instant-message", func(w http.ResponseWriter, r *http.Request) {
		instantMessageHandler(w, r, messageRelayer, logger)
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

	err = manager.DeleteUser(state.NewIdentScreenName(user.ScreenName))
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
	_, _ = fmt.Fprintln(w, "User account successfully deleted.")
}

func userPasswordHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, logger *slog.Logger) {
	switch r.Method {
	case http.MethodPut:
		putUserPasswordHandler(w, r, userManager, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// putUserPasswordHandler handles the PUT /user/password endpoint.
func putUserPasswordHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, logger *slog.Logger) {
	input, err := userFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sn := state.NewIdentScreenName(input.ScreenName)

	if err := userManager.SetUserPassword(sn, input.Password); err != nil {
		switch {
		case errors.Is(err, state.ErrNoUser):
			http.Error(w, "user does not exist", http.StatusNotFound)
			return
		case errors.Is(err, state.ErrPasswordInvalid):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case err != nil:
			logger.Error("error updating user password PUT /user/password", "err", err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
	_, _ = fmt.Fprintln(w, "Password successfully reset.")
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
		Sessions: make([]userHandle, len(allUsers)),
	}

	for i, s := range allUsers {
		ou.Sessions[i] = userHandle{
			ID:         s.IdentScreenName().String(),
			ScreenName: s.DisplayScreenName().String(),
			IsICQ:      s.UIN() > 0,
		}
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

	out := make([]userHandle, len(users))
	for i, u := range users {
		out[i] = userHandle{
			ID:         u.IdentScreenName.String(),
			ScreenName: u.DisplayScreenName.String(),
			IsICQ:      u.IsICQ,
		}
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// postUserHandler handles the POST /user endpoint.
func postUserHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, newUUID func() uuid.UUID, logger *slog.Logger) {
	input, err := userFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sn := state.DisplayScreenName(input.ScreenName)

	if sn.IsUIN() {
		if err := sn.ValidateUIN(); err != nil {
			http.Error(w, fmt.Sprintf("invalid uin: %s", err), http.StatusBadRequest)
			return
		}
	} else {
		if err := sn.ValidateAIMHandle(); err != nil {
			http.Error(w, fmt.Sprintf("invalid screen name: %s", err), http.StatusBadRequest)
			return
		}
	}

	user := state.User{
		AuthKey:           newUUID().String(),
		DisplayScreenName: sn,
		IdentScreenName:   sn.IdentScreenName(),
		IsICQ:             sn.IsUIN(),
	}

	if err := user.HashPassword(input.Password); err != nil {
		http.Error(w, fmt.Sprintf("invalid password: %s", err), http.StatusBadRequest)
		return
	}

	err = userManager.InsertUser(user)
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
	_, _ = fmt.Fprintln(w, "User account created successfully.")
}

func userFromBody(r *http.Request) (userWithPassword, error) {
	user := userWithPassword{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return userWithPassword{}, errors.New("malformed input")
	}
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
		_, _ = w.Write([]byte("401 Unauthorized\n"))
		return
	}

	auth := strings.SplitN(authHeader, " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("401 Unauthorized: Missing Basic prefix\n"))
		return
	}

	payload, err := base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("401 Unauthorized: Invalid Base64 Encoding\n"))
		return
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("401 Unauthorized: Invalid Authentication Token\n"))
		return
	}

	username, password := state.NewIdentScreenName(pair[0]), pair[1]

	user, err := userManager.User(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 InternalServerError\n"))
		logger.Error("error getting user", "err", err.Error())
		return
	}
	if user == nil || !user.ValidateHash(wire.StrongMD5PasswordHash(password, user.AuthKey)) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("401 Unauthorized: Invalid Credentials\n"))
		return
	}

	// Successfully authenticated
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("200 OK: Successfully Authenticated\n"))
}

func publicChatHandler(w http.ResponseWriter, r *http.Request, chatRoomRetriever ChatRoomRetriever, chatRoomCreator ChatRoomCreator, chatSessionRetriever ChatSessionRetriever, logger *slog.Logger) {
	switch r.Method {
	case http.MethodGet:
		getPublicChatHandler(w, r, chatRoomRetriever, chatSessionRetriever, logger)
	case http.MethodPost:
		postPublicChatHandler(w, r, chatRoomCreator, logger)
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
		sessions := chatSessionRetriever.AllSessions(room.Cookie())
		cr := chatRoom{
			CreateTime:   room.CreateTime(),
			Name:         room.Name(),
			Participants: make([]aimChatUserHandle, len(sessions)),
			URL:          room.URL().String(),
		}
		for j, sess := range sessions {
			cr.Participants[j] = aimChatUserHandle{
				ID:         sess.IdentScreenName().String(),
				ScreenName: sess.DisplayScreenName().String(),
			}
		}

		out[i] = cr
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// postPublicChatHandler handles the POST /chat/room/public endpoint.
func postPublicChatHandler(w http.ResponseWriter, r *http.Request, chatRoomCreator ChatRoomCreator, logger *slog.Logger) {
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

	cr := state.NewChatRoom(input.Name, state.NewIdentScreenName("system"), state.PublicExchange)

	err := chatRoomCreator.CreateChatRoom(&cr)
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
	_, _ = fmt.Fprintln(w, "Chat room created successfully.")
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
		sessions := chatSessionRetriever.AllSessions(room.Cookie())
		cr := chatRoom{
			CreateTime:   room.CreateTime(),
			CreatorID:    room.Creator().String(),
			Name:         room.Name(),
			Participants: make([]aimChatUserHandle, len(sessions)),
			URL:          room.URL().String(),
		}
		for j, sess := range sessions {
			cr.Participants[j] = aimChatUserHandle{
				ID:         sess.IdentScreenName().String(),
				ScreenName: sess.DisplayScreenName().String(),
			}
		}

		out[i] = cr
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func instantMessageHandler(w http.ResponseWriter, r *http.Request, messageRelayer MessageRelayer, logger *slog.Logger) {
	switch r.Method {
	case http.MethodPost:
		postInstantMessageHandler(w, r, messageRelayer, logger)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// postIMHandler handles the POST /instant-message endpoint.
func postInstantMessageHandler(w http.ResponseWriter, r *http.Request, messageRelayer MessageRelayer, logger *slog.Logger) {
	input := instantMessage{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "malformed input", http.StatusBadRequest)
		return
	}

	tlv, err := wire.ICBMFragmentList(input.Text)
	if err != nil {
		logger.Error("error sending message POST /instant-message", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	msg := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMChannelMsgToClient,
		},
		Body: wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
			ChannelID: 1,
			TLVUserInfo: wire.TLVUserInfo{
				ScreenName: input.From,
			},
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.ICBMTLVAOLIMData, tlv),
				},
			},
		},
	}
	messageRelayer.RelayToScreenName(context.Background(), state.NewIdentScreenName(input.To), msg)

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "Message sent successfully.")
}
