package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/time/rate"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/foodgroup"
	"github.com/mk6i/retro-aim-server/server/http"
	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/server/oscar/handler"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/server/toc"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// Container groups together common dependencies.
type Container struct {
	cfg                    config.Config
	chatSessionManager     *state.InMemoryChatSessionManager
	hmacCookieBaker        state.HMACCookieBaker
	inMemorySessionManager *state.InMemorySessionManager
	logger                 *slog.Logger
	rateLimitClasses       wire.RateLimitClasses
	snacRateLimits         wire.SNACRateLimits
	sqLiteUserStore        *state.SQLiteUserStore
}

// MakeCommonDeps creates common dependencies used by the food group services.
func MakeCommonDeps() (Container, error) {
	c := Container{}

	err := envconfig.Process("", &c.cfg)
	if err != nil {
		return c, fmt.Errorf("unable to process app config: %s\n", err.Error())
	}

	if c.cfg.OSCARHost == "0.0.0.0" {
		return c, errors.New("invalid config: OSCAR_HOST cannot be set to " +
			"the 'all interfaces' IP (0.0.0.0). it must be a specific IP " +
			"address or hostname reachable by AIM/ICQ clients")
	}

	c.sqLiteUserStore, err = state.NewSQLiteUserStore(c.cfg.DBPath)
	if err != nil {
		return c, fmt.Errorf("unable to create feedbag store: %s\n", err.Error())
	}

	c.hmacCookieBaker, err = state.NewHMACCookieBaker()
	if err != nil {
		return c, fmt.Errorf("unable to create HMAC cookie baker: %s\n", err.Error())
	}

	c.logger = middleware.NewLogger(c.cfg)
	c.inMemorySessionManager = state.NewInMemorySessionManager(c.logger)
	c.chatSessionManager = state.NewInMemoryChatSessionManager(c.logger)
	c.rateLimitClasses = wire.DefaultRateLimitClasses()
	c.snacRateLimits = wire.DefaultSNACRateLimits()
	return c, nil
}

// Admin creates an OSCAR server for the Admin food group.
func Admin(deps Container) oscar.AdminServer {
	logger := deps.logger.With("svc", "ADMIN")

	adminService := foodgroup.NewAdminService(
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.inMemorySessionManager,
		deps.logger,
	)
	authService := foodgroup.NewAuthService(
		deps.cfg,
		deps.inMemorySessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.rateLimitClasses,
	)
	oServiceService := foodgroup.NewOServiceServiceForAdmin(
		deps.cfg,
		logger,
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
		deps.rateLimitClasses,
		deps.snacRateLimits,
	)

	return oscar.AdminServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewAdminRouter(handler.Handlers{
			AdminHandler:    handler.NewAdminHandler(logger, adminService),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		Logger:           logger,
		OnlineNotifier:   oServiceService,
		ListenAddr:       net.JoinHostPort("", deps.cfg.AdminPort),
		RateLimitUpdater: oServiceService,
		SNACRateLimits:   deps.snacRateLimits,
	}
}

// Alert creates an OSCAR server for the Alert food group.
func Alert(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "ALERT")

	sessionManager := state.NewInMemorySessionManager(logger)
	authService := foodgroup.NewAuthService(
		deps.cfg,
		sessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		nil,
		deps.rateLimitClasses,
	)
	oServiceService := foodgroup.NewOServiceServiceForAlert(
		deps.cfg,
		logger,
		sessionManager,
		deps.sqLiteUserStore,
		sessionManager,
		deps.sqLiteUserStore,
		deps.rateLimitClasses,
		deps.snacRateLimits,
	)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewAlertRouter(handler.Handlers{
			AlertHandler:    handler.NewAlertHandler(logger),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		Logger:           logger,
		OnlineNotifier:   oServiceService,
		ListenAddr:       net.JoinHostPort("", deps.cfg.AlertPort),
		RateLimitUpdater: oServiceService,
		SNACRateLimits:   deps.snacRateLimits,
	}
}

// Auth creates an OSCAR server for the Auth food group.
func Auth(deps Container) oscar.AuthServer {
	logger := deps.logger.With("svc", "AUTH")

	authHandler := foodgroup.NewAuthService(
		deps.cfg,
		deps.inMemorySessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		nil,
		deps.rateLimitClasses,
	)

	return oscar.AuthServer{
		AuthService:   authHandler,
		Config:        deps.cfg,
		Logger:        logger,
		IPRateLimiter: oscar.NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
	}
}

// BART creates an OSCAR server for the BART food group.
func BART(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "BART")

	sessionManager := state.NewInMemorySessionManager(logger)
	bartService := foodgroup.NewBARTService(
		logger,
		deps.sqLiteUserStore,
		sessionManager,
		deps.sqLiteUserStore,
		sessionManager,
	)
	authService := foodgroup.NewAuthService(
		deps.cfg,
		sessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		nil,
		deps.rateLimitClasses,
	)
	oServiceService := foodgroup.NewOServiceServiceForBART(
		deps.cfg,
		logger,
		sessionManager,
		deps.sqLiteUserStore,
		sessionManager,
		deps.sqLiteUserStore,
		deps.rateLimitClasses,
		deps.snacRateLimits,
	)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewBARTRouter(handler.Handlers{
			BARTHandler:     handler.NewBARTHandler(logger, bartService),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		ListenAddr:       net.JoinHostPort("", deps.cfg.BARTPort),
		Logger:           logger,
		OnlineNotifier:   oServiceService,
		RateLimitUpdater: oServiceService,
		SNACRateLimits:   deps.snacRateLimits,
	}
}

// BOS creates an OSCAR server for the BOS food group.
func BOS(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "BOS")

	authService := foodgroup.NewAuthService(
		deps.cfg,
		deps.inMemorySessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		nil,
		deps.rateLimitClasses,
	)
	bartService := foodgroup.NewBARTService(
		logger,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
	)
	buddyService := foodgroup.NewBuddyService(
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
	)
	chatNavService := foodgroup.NewChatNavService(logger, deps.sqLiteUserStore)
	feedbagService := foodgroup.NewFeedbagService(
		logger,
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
	)
	permitDenyService := foodgroup.NewPermitDenyService(
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.inMemorySessionManager,
	)
	icbmService := foodgroup.NewICBMService(
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.snacRateLimits,
	)
	icqService := foodgroup.NewICQService(deps.inMemorySessionManager, deps.sqLiteUserStore, deps.sqLiteUserStore,
		logger, deps.inMemorySessionManager, deps.sqLiteUserStore)
	locateService := foodgroup.NewLocateService(
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
	)
	oServiceService := foodgroup.NewOServiceServiceForBOS(
		deps.cfg,
		deps.inMemorySessionManager,
		logger,
		deps.hmacCookieBaker,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.sqLiteUserStore,
		deps.rateLimitClasses,
		deps.snacRateLimits,
	)
	userLookupService := foodgroup.NewUserLookupService(deps.sqLiteUserStore)
	statsService := foodgroup.NewStatsService()

	return oscar.BOSServer{
		AuthService:        authService,
		BuddyListRegistry:  deps.sqLiteUserStore,
		Config:             deps.cfg,
		DepartureNotifier:  buddyService,
		ChatSessionManager: deps.chatSessionManager,
		Handler: handler.NewBOSRouter(handler.Handlers{
			AlertHandler:      handler.NewAlertHandler(logger),
			BARTHandler:       handler.NewBARTHandler(logger, bartService),
			BuddyHandler:      handler.NewBuddyHandler(logger, buddyService),
			ChatNavHandler:    handler.NewChatNavHandler(chatNavService, logger),
			FeedbagHandler:    handler.NewFeedbagHandler(logger, feedbagService),
			ICQHandler:        handler.NewICQHandler(logger, icqService),
			ICBMHandler:       handler.NewICBMHandler(logger, icbmService),
			LocateHandler:     handler.NewLocateHandler(locateService, logger),
			OServiceHandler:   handler.NewOServiceHandler(logger, oServiceService),
			PermitDenyHandler: handler.NewPermitDenyHandler(logger, permitDenyService),
			StatsHandler:      handler.NewStatsHandler(logger, statsService),
			UserLookupHandler: handler.NewUserLookupHandler(logger, userLookupService),
		}),
		Logger:           logger,
		OnlineNotifier:   oServiceService,
		ListenAddr:       net.JoinHostPort("", deps.cfg.BOSPort),
		RateLimitUpdater: oServiceService,
		SNACRateLimits:   deps.snacRateLimits,
	}
}

// Chat creates an OSCAR server for the Chat food group.
func Chat(deps Container) oscar.ChatServer {
	logger := deps.logger.With("svc", "CHAT")

	sessionManager := state.NewInMemorySessionManager(logger)
	authService := foodgroup.NewAuthService(
		deps.cfg,
		sessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		nil,
		deps.rateLimitClasses,
	)
	chatService := foodgroup.NewChatService(deps.chatSessionManager)
	oServiceService := foodgroup.NewOServiceServiceForChat(
		deps.cfg,
		logger,
		sessionManager,
		deps.sqLiteUserStore,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		sessionManager,
		deps.sqLiteUserStore,
		deps.rateLimitClasses,
		deps.snacRateLimits,
	)

	return oscar.ChatServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewChatRouter(handler.Handlers{
			ChatHandler:     handler.NewChatHandler(logger, chatService),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		Logger:           logger,
		OnlineNotifier:   oServiceService,
		SNACRateLimits:   deps.snacRateLimits,
		RateLimitUpdater: oServiceService,
	}
}

// ChatNav creates an OSCAR server for the ChatNav food group.
func ChatNav(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "CHAT_NAV")

	sessionManager := state.NewInMemorySessionManager(logger)
	authService := foodgroup.NewAuthService(
		deps.cfg,
		sessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		nil,
		deps.rateLimitClasses,
	)
	chatNavService := foodgroup.NewChatNavService(logger, deps.sqLiteUserStore)
	oServiceService := foodgroup.NewOServiceServiceForChatNav(
		deps.cfg,
		logger,
		sessionManager,
		deps.sqLiteUserStore,
		sessionManager,
		deps.sqLiteUserStore,
		deps.rateLimitClasses,
		deps.snacRateLimits,
	)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewChatNavRouter(handler.Handlers{
			ChatNavHandler:  handler.NewChatNavHandler(chatNavService, logger),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		Logger:           logger,
		OnlineNotifier:   oServiceService,
		ListenAddr:       net.JoinHostPort("", deps.cfg.ChatNavPort),
		RateLimitUpdater: oServiceService,
		SNACRateLimits:   deps.snacRateLimits,
	}
}

// MgmtAPI creates an HTTP server for the management API.
func MgmtAPI(deps Container) *http.Server {
	bld := config.Build{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	return http.NewManagementAPI(bld, deps.cfg, deps.sqLiteUserStore, deps.inMemorySessionManager, deps.sqLiteUserStore,
		deps.sqLiteUserStore, deps.chatSessionManager, deps.sqLiteUserStore, deps.inMemorySessionManager,
		deps.sqLiteUserStore, deps.sqLiteUserStore, deps.sqLiteUserStore, deps.sqLiteUserStore, deps.logger)
}

// ODir creates an OSCAR server for the ODir food group.
func ODir(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "ODIR")

	sessionManager := state.NewInMemorySessionManager(logger)
	authService := foodgroup.NewAuthService(
		deps.cfg,
		sessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		nil,
		deps.rateLimitClasses,
	)
	oServiceService := foodgroup.NewOServiceServiceForODir(
		deps.cfg,
		logger,
		deps.rateLimitClasses,
		deps.snacRateLimits,
	)
	oDirService := foodgroup.NewODirService(logger, deps.sqLiteUserStore)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewODirRouter(handler.Handlers{
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			ODirHandler:     handler.NewODirHandler(logger, oDirService),
		}),
		Logger:           logger,
		OnlineNotifier:   oServiceService,
		ListenAddr:       net.JoinHostPort("", deps.cfg.ODirPort),
		RateLimitUpdater: oServiceService,
		SNACRateLimits:   deps.snacRateLimits,
	}
}

// TOC creates a TOC server.
func TOC(deps Container) toc.Server {
	logger := deps.logger.With("svc", "TOC")
	sessionManager := state.NewInMemorySessionManager(logger)
	return toc.Server{
		Logger:     logger,
		ListenAddr: net.JoinHostPort(deps.cfg.TOCHost, deps.cfg.TOCPort),
		BOSProxy: toc.OSCARProxy{
			AdminService: foodgroup.NewAdminService(
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.inMemorySessionManager,
				deps.logger,
			),
			AuthService: foodgroup.NewAuthService(
				deps.cfg,
				deps.inMemorySessionManager,
				deps.chatSessionManager,
				deps.sqLiteUserStore,
				deps.hmacCookieBaker,
				deps.chatSessionManager,
				deps.sqLiteUserStore,
				nil,
				deps.rateLimitClasses,
			),
			BuddyListRegistry: deps.sqLiteUserStore,
			BuddyService: foodgroup.NewBuddyService(
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
			),
			CookieBaker:      deps.hmacCookieBaker,
			DirSearchService: foodgroup.NewODirService(logger, deps.sqLiteUserStore),
			ICBMService: foodgroup.NewICBMService(
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.snacRateLimits,
			),
			LocateService: foodgroup.NewLocateService(
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
			),
			Logger: logger,
			OServiceServiceBOS: foodgroup.NewOServiceServiceForBOS(
				deps.cfg,
				deps.inMemorySessionManager,
				logger,
				deps.hmacCookieBaker,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
				deps.rateLimitClasses,
				deps.snacRateLimits,
			),
			PermitDenyService: foodgroup.NewPermitDenyService(
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.inMemorySessionManager,
			),
			TOCConfigStore: deps.sqLiteUserStore,
			ChatService:    foodgroup.NewChatService(deps.chatSessionManager),
			OServiceServiceChat: foodgroup.NewOServiceServiceForChat(
				deps.cfg,
				logger,
				sessionManager,
				deps.sqLiteUserStore,
				deps.chatSessionManager,
				deps.sqLiteUserStore,
				sessionManager,
				deps.sqLiteUserStore,
				deps.rateLimitClasses,
				deps.snacRateLimits,
			),
			ChatNavService:    foodgroup.NewChatNavService(logger, deps.sqLiteUserStore),
			SNACRateLimits:    deps.snacRateLimits,
			HTTPIPRateLimiter: toc.NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
		},
		LoginIPRateLimiter: toc.NewIPRateLimiter(rate.Every(10*time.Minute), 10, 20*time.Minute),
	}
}
