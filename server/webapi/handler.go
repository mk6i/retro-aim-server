package webapi

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mk6i/retro-aim-server/wire"
)

type Handler struct {
	AdminService      AdminService
	AuthService       AuthService
	BuddyListRegistry BuddyListRegistry
	BuddyService      BuddyService
	ChatNavService    ChatNavService
	ChatService       ChatService
	CookieBaker       CookieBaker
	DirSearchService  DirSearchService
	ICBMService       ICBMService
	LocateService     LocateService
	Logger            *slog.Logger
	OServiceService   OServiceService
	PermitDenyService PermitDenyService
	TOCConfigStore    TOCConfigStore
	SNACRateLimits    wire.SNACRateLimits
	// New fields for WebAPI handlers
	SessionRetriever SessionRetriever
	FeedbagRetriever FeedbagRetriever
	FeedbagManager   FeedbagManager
	// Phase 2 additions
	MessageRelayer        MessageRelayer
	OfflineMessageManager OfflineMessageManager
	BuddyBroadcaster      BuddyBroadcaster
	ProfileManager        ProfileManager
	// Authentication support
	UserManager UserManager
	TokenStore  TokenStore
}

func (h Handler) GetHelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("got a Hello World request!")
	_, _ = fmt.Fprintln(w, "Hello ukozi!")
}
