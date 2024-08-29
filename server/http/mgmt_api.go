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
	"time"

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
	bartRetriever BARTRetriever,
	feedbagRetriever FeedBagRetriever,
	accountRetriever AccountRetriever,
	profileRetriever ProfileRetriever,
	logger *slog.Logger,
) {

	mux := http.NewServeMux()

	// Handlers for '/user' route
	mux.HandleFunc("DELETE /user", func(w http.ResponseWriter, r *http.Request) {
		deleteUserHandler(w, r, userManager, logger)
	})
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		getUserHandler(w, userManager, logger)
	})
	mux.HandleFunc("POST /user", func(w http.ResponseWriter, r *http.Request) {
		postUserHandler(w, r, userManager, uuid.New, logger)
	})

	// Handlers for '/user/password' route
	mux.HandleFunc("PUT /user/password", func(w http.ResponseWriter, r *http.Request) {
		putUserPasswordHandler(w, r, userManager, logger)
	})

	// Handlers for '/user/login' route
	mux.HandleFunc("GET /user/login", func(w http.ResponseWriter, r *http.Request) {
		getUserLoginHandler(w, r, userManager, logger)
	})

	// Handlers for '/user/{screenname}/account' route
	mux.HandleFunc("GET /user/{screenname}/account", func(w http.ResponseWriter, r *http.Request) {
		getUserAccountHandler(w, r, userManager, accountRetriever, profileRetriever, logger)
	})

	// Handlers for '/user/{screenname}/icon' route
	mux.HandleFunc("GET /user/{screenname}/icon", func(w http.ResponseWriter, r *http.Request) {
		getUserBuddyIconHandler(w, r, userManager, feedbagRetriever, bartRetriever, logger)
	})

	// Handlers for '/session' route
	mux.HandleFunc("GET /session", func(w http.ResponseWriter, r *http.Request) {
		getSessionHandler(w, r, sessionRetriever, time.Since)
	})

	// Handlers for '/session/{screenname}' route
	mux.HandleFunc("GET /session/{screenname}", func(w http.ResponseWriter, r *http.Request) {
		getSessionHandler(w, r, sessionRetriever, time.Since)
	})

	// Handlers for '/chat/room/public' route
	mux.HandleFunc("GET /chat/room/public", func(w http.ResponseWriter, r *http.Request) {
		getPublicChatHandler(w, r, chatRoomRetriever, chatSessionRetriever, logger)
	})
	mux.HandleFunc("POST /chat/room/public", func(w http.ResponseWriter, r *http.Request) {
		postPublicChatHandler(w, r, chatRoomCreator, logger)
	})

	// Handlers for '/chat/room/private' route
	mux.HandleFunc("GET /chat/room/private", func(w http.ResponseWriter, r *http.Request) {
		getPrivateChatHandler(w, r, chatRoomRetriever, chatSessionRetriever, logger)
	})

	// Handlers for '/instant-message' route
	mux.HandleFunc("POST /instant-message", func(w http.ResponseWriter, r *http.Request) {
		postInstantMessageHandler(w, r, messageRelayer, logger)
	})

	addr := net.JoinHostPort(cfg.ApiHost, cfg.ApiPort)
	logger.Info("starting management API server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("unable to bind management API address address", "err", err.Error())
		os.Exit(1)
	}
}

// deleteUserHandler handles the DELETE /user endpoint.
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
		default:
			logger.Error("error updating user password PUT /user/password", "err", err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
	_, _ = fmt.Fprintln(w, "Password successfully reset.")
}

// getSessionHandler handles GET /session
func getSessionHandler(w http.ResponseWriter, r *http.Request, sessionRetriever SessionRetriever, funcTimeSince func(t time.Time) time.Duration) {
	w.Header().Set("Content-Type", "application/json")

	var allUsers []*state.Session

	if screenName := r.PathValue("screenname"); screenName != "" {
		session := sessionRetriever.RetrieveByScreenName(state.NewIdentScreenName(screenName))
		if session == nil {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		allUsers = append(allUsers, session)
	} else {
		allUsers = sessionRetriever.AllSessions()
	}

	ou := onlineUsers{
		Count:    len(allUsers),
		Sessions: make([]sessionHandle, len(allUsers)),
	}

	for i, s := range allUsers {
		// report 0 if the user is not idle
		idleSeconds := funcTimeSince(s.IdleTime()).Seconds()
		if !s.Idle() {
			idleSeconds = 0
		}
		onlineSeconds := funcTimeSince(s.SignonTime()).Seconds()

		ou.Sessions[i] = sessionHandle{
			ID:            s.IdentScreenName().String(),
			ScreenName:    s.DisplayScreenName().String(),
			OnlineSeconds: onlineSeconds,
			AwayMessage:   s.AwayMessage(),
			IdleSeconds:   idleSeconds,
			IsICQ:         s.UIN() > 0,
		}
	}

	if err := json.NewEncoder(w).Encode(ou); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getUserHandler handles the GET /user endpoint.
func getUserHandler(w http.ResponseWriter, userManager UserManager, logger *slog.Logger) {
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

// getUserLoginHandler is a temporary endpoint for validating user credentials for
// chivanet. do not rely on this endpoint, as it will be eventually removed.
func getUserLoginHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, logger *slog.Logger) {
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

// getUserBuddyIconHandler handles the GET /user/{screenname}/icon endpoint.
func getUserBuddyIconHandler(w http.ResponseWriter, r *http.Request, u UserManager, f FeedBagRetriever, b BARTRetriever, logger *slog.Logger) {
	screenName := state.NewIdentScreenName(r.PathValue("screenname"))
	user, err := u.User(screenName)
	if err != nil {
		logger.Error("error retrieving user", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	iconRef, err := f.BuddyIconRefByName(screenName)
	if err != nil {
		logger.Error("error retrieving buddy icon ref", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if iconRef == nil || iconRef.HasClearIconHash() {
		http.Error(w, "icon not found", http.StatusNotFound)
		return
	}
	icon, err := b.BARTRetrieve(iconRef.Hash)
	if err != nil {
		logger.Error("error retrieving buddy icon bart item", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", http.DetectContentType(icon))
	w.Write(icon)
}

// getUserAccountHandler handles the GET /user/{screenname}/account endpoint.
func getUserAccountHandler(w http.ResponseWriter, r *http.Request, userManager UserManager, a AccountRetriever, p ProfileRetriever, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")

	screenName := r.PathValue("screenname")
	user, err := userManager.User(state.NewIdentScreenName(screenName))
	if err != nil {
		logger.Error("error in GET /user/{screenname}/account", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	emailAddress := ""
	email, err := a.EmailAddressByName(user.IdentScreenName)
	if err != nil {
		emailAddress = ""
	} else {
		emailAddress = email.String()
	}
	regStatus, err := a.RegStatusByName(user.IdentScreenName)
	if err != nil {
		logger.Error("error in GET /user/*/account RegStatus", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	confirmStatus, err := a.ConfirmStatusByName(user.IdentScreenName)
	if err != nil {
		logger.Error("error in GET /user/*/account ConfirmStatus", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	profile, err := p.Profile(user.IdentScreenName)
	if err != nil {
		logger.Error("error in GET /user/*/account Profile", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := userAccountHandle{
		ID:           user.IdentScreenName.String(),
		ScreenName:   user.DisplayScreenName.String(),
		EmailAddress: emailAddress,
		RegStatus:    regStatus,
		Confirmed:    confirmStatus,
		Profile:      profile,
		IsICQ:        user.IsICQ,
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
