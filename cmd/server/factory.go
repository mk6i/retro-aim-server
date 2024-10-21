package main

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/kelseyhightower/envconfig"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/foodgroup"
	"github.com/mk6i/retro-aim-server/server/http"
	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/server/oscar/handler"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
)

// Container groups together common dependencies.
type Container struct {
	adjListBuddyListStore  *state.AdjListBuddyListStore
	cfg                    config.Config
	chatSessionManager     *state.InMemoryChatSessionManager
	hmacCookieBaker        state.HMACCookieBaker
	inMemorySessionManager *state.InMemorySessionManager
	logger                 *slog.Logger
	sqLiteUserStore        *state.SQLiteUserStore
}

// MakeCommonDeps creates common dependencies used by the food group services.
func MakeCommonDeps() (Container, error) {
	c := Container{}

	err := envconfig.Process("", &c.cfg)
	if err != nil {
		return c, fmt.Errorf("unable to process app config: %s\n", err.Error())
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
	c.adjListBuddyListStore = state.NewAdjListBuddyListStore()

	return c, nil
}

// Admin creates an OSCAR server for the Admin food group.
func Admin(deps Container) oscar.AdminServer {
	logger := deps.logger.With("svc", "ADMIN")

	buddyService := foodgroup.NewBuddyService(deps.inMemorySessionManager, deps.sqLiteUserStore, deps.adjListBuddyListStore)
	adminService := foodgroup.NewAdminService(deps.inMemorySessionManager, deps.sqLiteUserStore, buddyService, deps.inMemorySessionManager)
	authService := foodgroup.NewAuthService(deps.cfg, deps.inMemorySessionManager, deps.chatSessionManager,
		deps.sqLiteUserStore, deps.adjListBuddyListStore, deps.hmacCookieBaker, deps.inMemorySessionManager,
		deps.sqLiteUserStore, deps.chatSessionManager, deps.sqLiteUserStore)
	oServiceService := foodgroup.NewOServiceServiceForAdmin(deps.cfg, logger, buddyService)

	return oscar.AdminServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewAdminRouter(handler.Handlers{
			AdminHandler:    handler.NewAdminHandler(logger, adminService),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		Logger:         logger,
		OnlineNotifier: oServiceService,
		ListenAddr:     net.JoinHostPort("", deps.cfg.AdminPort),
	}
}

// Alert creates an OSCAR server for the Alert food group.
func Alert(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "ALERT")

	sessionManager := state.NewInMemorySessionManager(logger)
	authService := foodgroup.NewAuthService(deps.cfg, sessionManager, deps.chatSessionManager, deps.sqLiteUserStore,
		deps.adjListBuddyListStore, deps.hmacCookieBaker, sessionManager, deps.sqLiteUserStore, deps.chatSessionManager, deps.sqLiteUserStore)
	oServiceService := foodgroup.NewOServiceServiceForAlert(deps.cfg, logger, sessionManager, deps.adjListBuddyListStore, deps.sqLiteUserStore)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewAlertRouter(handler.Handlers{
			AlertHandler:    handler.NewAlertHandler(logger),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		Logger:         logger,
		OnlineNotifier: oServiceService,
		ListenAddr:     net.JoinHostPort("", deps.cfg.AlertPort),
	}
}

// Auth creates an OSCAR server for the Auth food group.
func Auth(deps Container) oscar.AuthServer {
	logger := deps.logger.With("svc", "AUTH")

	authHandler := foodgroup.NewAuthService(deps.cfg, deps.inMemorySessionManager, deps.chatSessionManager,
		deps.sqLiteUserStore, deps.adjListBuddyListStore, deps.hmacCookieBaker, nil, nil, deps.chatSessionManager, deps.sqLiteUserStore)

	return oscar.AuthServer{
		AuthService: authHandler,
		Config:      deps.cfg,
		Logger:      logger,
	}
}

// BART creates an OSCAR server for the BART food group.
func BART(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "BART")

	sessionManager := state.NewInMemorySessionManager(logger)
	bartService := foodgroup.NewBARTService(logger, deps.sqLiteUserStore, sessionManager, deps.sqLiteUserStore, deps.adjListBuddyListStore)
	authService := foodgroup.NewAuthService(deps.cfg, sessionManager, deps.chatSessionManager, deps.sqLiteUserStore,
		deps.adjListBuddyListStore, deps.hmacCookieBaker, sessionManager, deps.sqLiteUserStore, deps.chatSessionManager, deps.sqLiteUserStore)
	oServiceService := foodgroup.NewOServiceServiceForBART(deps.cfg, logger, sessionManager, deps.adjListBuddyListStore, deps.sqLiteUserStore)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewBARTRouter(handler.Handlers{
			BARTHandler:     handler.NewBARTHandler(logger, bartService),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		ListenAddr:     net.JoinHostPort("", deps.cfg.BARTPort),
		Logger:         logger,
		OnlineNotifier: oServiceService,
	}
}

// BOS creates an OSCAR server for the BOS food group.
func BOS(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "BOS")

	authService := foodgroup.NewAuthService(deps.cfg, deps.inMemorySessionManager, deps.chatSessionManager,
		deps.sqLiteUserStore, deps.adjListBuddyListStore, deps.hmacCookieBaker, deps.inMemorySessionManager,
		deps.sqLiteUserStore, deps.chatSessionManager, deps.sqLiteUserStore)
	bartService := foodgroup.NewBARTService(logger, deps.sqLiteUserStore, deps.inMemorySessionManager,
		deps.sqLiteUserStore, deps.adjListBuddyListStore)
	buddyService := foodgroup.NewBuddyService(deps.inMemorySessionManager, deps.sqLiteUserStore, deps.adjListBuddyListStore)
	chatNavService := foodgroup.NewChatNavService(logger, deps.sqLiteUserStore)
	feedbagService := foodgroup.NewFeedbagService(logger, deps.inMemorySessionManager, deps.sqLiteUserStore,
		deps.sqLiteUserStore, deps.adjListBuddyListStore)
	permitDenyService := foodgroup.NewPermitDenyService()
	icbmService := foodgroup.NewICBMService(deps.inMemorySessionManager, deps.sqLiteUserStore,
		deps.adjListBuddyListStore, deps.sqLiteUserStore)
	icqService := foodgroup.NewICQService(deps.inMemorySessionManager, deps.sqLiteUserStore, deps.sqLiteUserStore,
		logger, deps.inMemorySessionManager, deps.sqLiteUserStore)
	locateService := foodgroup.NewLocateService(deps.inMemorySessionManager, deps.sqLiteUserStore, deps.sqLiteUserStore,
		deps.adjListBuddyListStore)
	oServiceService := foodgroup.NewOServiceServiceForBOS(deps.cfg, deps.inMemorySessionManager,
		deps.adjListBuddyListStore, logger, deps.hmacCookieBaker, deps.sqLiteUserStore, deps.sqLiteUserStore)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
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
		}),
		Logger:         logger,
		OnlineNotifier: oServiceService,
		ListenAddr:     net.JoinHostPort("", deps.cfg.BOSPort),
	}
}

// Chat creates an OSCAR server for the Chat food group.
func Chat(deps Container) oscar.ChatServer {
	logger := deps.logger.With("svc", "CHAT")

	sessionManager := state.NewInMemorySessionManager(logger)
	authService := foodgroup.NewAuthService(deps.cfg, sessionManager, deps.chatSessionManager, deps.sqLiteUserStore,
		deps.adjListBuddyListStore, deps.hmacCookieBaker, sessionManager, deps.sqLiteUserStore, deps.chatSessionManager, deps.sqLiteUserStore)
	chatService := foodgroup.NewChatService(deps.chatSessionManager)
	oServiceService := foodgroup.NewOServiceServiceForChat(deps.cfg, logger, sessionManager, deps.adjListBuddyListStore,
		deps.sqLiteUserStore, deps.sqLiteUserStore, deps.chatSessionManager)

	return oscar.ChatServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewChatRouter(handler.Handlers{
			ChatHandler:     handler.NewChatHandler(logger, chatService),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		Logger:         logger,
		OnlineNotifier: oServiceService,
	}
}

// ChatNav creates an OSCAR server for the ChatNav food group.
func ChatNav(deps Container) oscar.BOSServer {
	logger := deps.logger.With("svc", "CHAT_NAV")

	sessionManager := state.NewInMemorySessionManager(logger)
	authService := foodgroup.NewAuthService(deps.cfg, sessionManager, deps.chatSessionManager, deps.sqLiteUserStore,
		deps.adjListBuddyListStore, deps.hmacCookieBaker, sessionManager, deps.sqLiteUserStore, deps.chatSessionManager,
		deps.sqLiteUserStore)
	chatNavService := foodgroup.NewChatNavService(logger, deps.sqLiteUserStore)
	oServiceService := foodgroup.NewOServiceServiceForChatNav(deps.cfg, logger, sessionManager, deps.adjListBuddyListStore, deps.sqLiteUserStore)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewChatNavRouter(handler.Handlers{
			ChatNavHandler:  handler.NewChatNavHandler(chatNavService, logger),
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
		}),
		Logger:         logger,
		OnlineNotifier: oServiceService,
		ListenAddr:     net.JoinHostPort("", deps.cfg.ChatNavPort),
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
	authService := foodgroup.NewAuthService(deps.cfg, sessionManager, deps.chatSessionManager, deps.sqLiteUserStore,
		deps.adjListBuddyListStore, deps.hmacCookieBaker, sessionManager, deps.sqLiteUserStore, deps.chatSessionManager,
		deps.sqLiteUserStore)
	oServiceService := foodgroup.NewOServiceServiceForODir(deps.cfg, logger)
	oDirService := foodgroup.NewODirService(logger, deps.sqLiteUserStore)

	return oscar.BOSServer{
		AuthService: authService,
		Config:      deps.cfg,
		Handler: handler.NewODirRouter(handler.Handlers{
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			ODirHandler:     handler.NewODirHandler(logger, oDirService),
		}),
		Logger:         logger,
		OnlineNotifier: oServiceService,
		ListenAddr:     net.JoinHostPort("", deps.cfg.ODirPort),
	}
}
