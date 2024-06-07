package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/foodgroup"
	"github.com/mk6i/retro-aim-server/server/http"
	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/server/oscar/handler"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg config.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to process app config: %s\n", err.Error())
		os.Exit(1)
	}

	feedbagStore, err := state.NewSQLiteUserStore(cfg.DBPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to create feedbag store: %s\n", err.Error())
		os.Exit(1)
	}

	cookieBaker, err := state.NewHMACCookieBaker()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to create HMAC cookie baker: %s\n", err.Error())
		os.Exit(1)
	}

	logger := middleware.NewLogger(cfg)
	sessionManager := state.NewInMemorySessionManager(logger)
	chatRegistry := state.NewChatRegistry()
	adjListBuddyListStore := state.NewAdjListBuddyListStore()

	wg := sync.WaitGroup{}
	wg.Add(7)

	go func() {
		http.StartManagementAPI(cfg, feedbagStore, sessionManager, logger)
		wg.Done()
	}()
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "BOS")
		buddyService := foodgroup.NewBuddyService(sessionManager, feedbagStore, adjListBuddyListStore)
		authService := foodgroup.NewAuthService(cfg, sessionManager, feedbagStore, chatRegistry, adjListBuddyListStore, cookieBaker, buddyService)
		bartService := foodgroup.NewBARTService(logger, feedbagStore, buddyService)
		oServiceService := foodgroup.NewOServiceServiceForBOS(cfg, sessionManager, adjListBuddyListStore, logger, cookieBaker, buddyService, chatRegistry)
		locateService := foodgroup.NewLocateService(sessionManager, feedbagStore, feedbagStore, buddyService)
		newChatSessMgr := func() foodgroup.SessionManager { return state.NewInMemorySessionManager(logger) }
		chatNavService := foodgroup.NewChatNavService(logger, chatRegistry, state.NewChatRoom, newChatSessMgr)
		feedbagService := foodgroup.NewFeedbagService(logger, sessionManager, feedbagStore, feedbagStore, buddyService)
		icbmService := foodgroup.NewICBMService(sessionManager, feedbagStore, buddyService)
		foodgroupService := foodgroup.NewPermitDenyService()

		oscar.BOSServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewBOSRouter(handler.Handlers{
				AlertHandler:      handler.NewAlertHandler(logger),
				BARTHandler:       handler.NewBARTHandler(logger, bartService),
				BuddyHandler:      handler.NewBuddyHandler(logger, buddyService),
				ChatNavHandler:    handler.NewChatNavHandler(chatNavService, logger),
				FeedbagHandler:    handler.NewFeedbagHandler(logger, feedbagService),
				ICBMHandler:       handler.NewICBMHandler(logger, icbmService),
				LocateHandler:     handler.NewLocateHandler(locateService, logger),
				OServiceHandler:   handler.NewOServiceHandler(logger, oServiceService),
				PermitDenyHandler: handler.NewPermitDenyHandler(logger, foodgroupService),
			}),
			CookieCracker:  cookieBaker,
			Logger:         logger,
			OnlineNotifier: oServiceService,
			ListenAddr:     net.JoinHostPort("", cfg.BOSPort),
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "CHAT")
		sessionManager := state.NewInMemorySessionManager(logger)
		buddyService := foodgroup.NewBuddyService(nil, feedbagStore, adjListBuddyListStore)
		authService := foodgroup.NewAuthService(cfg, sessionManager, feedbagStore, chatRegistry, adjListBuddyListStore, cookieBaker, buddyService)
		chatService := foodgroup.NewChatService(chatRegistry)
		oServiceService := foodgroup.NewOServiceServiceForChat(cfg, logger, buddyService, chatRegistry)

		oscar.ChatServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewChatRouter(handler.Handlers{
				ChatHandler:     handler.NewChatHandler(logger, chatService),
				OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceService,
			CookieCracker:  cookieBaker,
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "CHAT_NAV")
		sessionManager := state.NewInMemorySessionManager(logger)
		buddyService := foodgroup.NewBuddyService(sessionManager, feedbagStore, adjListBuddyListStore)
		authService := foodgroup.NewAuthService(cfg, sessionManager, feedbagStore, chatRegistry, adjListBuddyListStore, cookieBaker, buddyService)
		oServiceService := foodgroup.NewOServiceServiceForChatNav(cfg, logger, buddyService)
		newChatSessMgr := func() foodgroup.SessionManager { return state.NewInMemorySessionManager(logger) }
		chatNavService := foodgroup.NewChatNavService(logger, chatRegistry, state.NewChatRoom, newChatSessMgr)

		oscar.BOSServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewChatNavRouter(handler.Handlers{
				ChatNavHandler:  handler.NewChatNavHandler(chatNavService, logger),
				OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceService,
			ListenAddr:     net.JoinHostPort("", cfg.ChatNavPort),
			CookieCracker:  cookieBaker,
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "ALERT")
		sessionManager := state.NewInMemorySessionManager(logger)
		buddyService := foodgroup.NewBuddyService(sessionManager, feedbagStore, adjListBuddyListStore)
		authService := foodgroup.NewAuthService(cfg, sessionManager, feedbagStore, chatRegistry, adjListBuddyListStore, cookieBaker, buddyService)
		oServiceService := foodgroup.NewOServiceServiceForAlert(cfg, logger, buddyService)

		oscar.BOSServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewAlertRouter(handler.Handlers{
				AlertHandler:    handler.NewAlertHandler(logger),
				OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			}),
			CookieCracker:  cookieBaker,
			Logger:         logger,
			OnlineNotifier: oServiceService,
			ListenAddr:     net.JoinHostPort("", cfg.AlertPort),
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "BART")
		sessionManager := state.NewInMemorySessionManager(logger)
		buddyService := foodgroup.NewBuddyService(sessionManager, feedbagStore, adjListBuddyListStore)
		bartService := foodgroup.NewBARTService(logger, feedbagStore, buddyService)
		authService := foodgroup.NewAuthService(cfg, sessionManager, feedbagStore, chatRegistry, adjListBuddyListStore, cookieBaker, buddyService)
		oServiceService := foodgroup.NewOServiceServiceForBART(cfg, logger, buddyService)

		oscar.BOSServer{
			AuthService:   authService,
			Config:        cfg,
			CookieCracker: cookieBaker,
			Handler: handler.NewBARTRouter(handler.Handlers{
				BARTHandler:     handler.NewBARTHandler(logger, bartService),
				OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			}),
			ListenAddr:     net.JoinHostPort("", cfg.BARTPort),
			Logger:         logger,
			OnlineNotifier: oServiceService,
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "AUTH")
		authHandler := foodgroup.NewAuthService(cfg, sessionManager, feedbagStore, chatRegistry, adjListBuddyListStore, cookieBaker, nil)

		oscar.AuthServer{
			AuthService:   authHandler,
			Config:        cfg,
			Logger:        logger,
			CookieCracker: cookieBaker,
		}.Start()
		wg.Done()
	}(logger)

	wg.Wait()
}
