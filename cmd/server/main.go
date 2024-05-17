package main

import (
	"fmt"
	"log/slog"
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
		_, _ = fmt.Fprintf(os.Stderr, "unable to process app config: %s", err.Error())
		os.Exit(1)
	}

	var feedbagStore *state.SQLUserStore
	switch cfg.DBType {
	case "sqlite3":
		feedbagStore, err = state.NewSQLiteUserStore(cfg.DBPath)
	case "postgres":
		feedbagStore, err = state.NewPostgresUserStore(cfg.DBPath)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unsupported database type: %s", cfg.DBType)
		os.Exit(1)
	}

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to create feedbag store: %s", err.Error())
		os.Exit(1)
	}

	logger := middleware.NewLogger(cfg)
	sessionManager := state.NewInMemorySessionManager(logger)
	chatRegistry := state.NewChatRegistry()
	adjListBuddyListStore := state.NewAdjListBuddyListStore()

	wg := sync.WaitGroup{}
	wg.Add(6)

	go func() {
		http.StartManagementAPI(feedbagStore, sessionManager, logger)
		wg.Done()
	}()
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "BOS")
		authService := foodgroup.NewAuthService(cfg, sessionManager, sessionManager, feedbagStore, feedbagStore, chatRegistry, adjListBuddyListStore)
		bartService := foodgroup.NewBARTService(logger, feedbagStore, sessionManager, feedbagStore, adjListBuddyListStore)
		buddyService := foodgroup.NewBuddyService(sessionManager, feedbagStore, adjListBuddyListStore)
		oServiceService := foodgroup.NewOServiceService(cfg, sessionManager, feedbagStore, adjListBuddyListStore, logger)
		oServiceServiceForBOS := foodgroup.NewOServiceServiceForBOS(*oServiceService, chatRegistry)
		locateService := foodgroup.NewLocateService(sessionManager, feedbagStore, feedbagStore, adjListBuddyListStore)
		newChatSessMgr := func() foodgroup.SessionManager { return state.NewInMemorySessionManager(logger) }
		chatNavService := foodgroup.NewChatNavService(logger, chatRegistry, state.NewChatRoom, newChatSessMgr)
		feedbagService := foodgroup.NewFeedbagService(logger, sessionManager, feedbagStore, feedbagStore, adjListBuddyListStore)
		icbmService := foodgroup.NewICBMService(sessionManager, feedbagStore, adjListBuddyListStore)

		oscar.BOSServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewBOSRouter(handler.Handlers{
				AlertHandler:       handler.NewAlertHandler(logger),
				BARTHandler:        handler.NewBARTHandler(logger, bartService),
				BuddyHandler:       handler.NewBuddyHandler(logger, buddyService),
				ChatNavHandler:     handler.NewChatNavHandler(chatNavService, logger),
				FeedbagHandler:     handler.NewFeedbagHandler(logger, feedbagService),
				ICBMHandler:        handler.NewICBMHandler(logger, icbmService),
				LocateHandler:      handler.NewLocateHandler(locateService, logger),
				OServiceBOSHandler: handler.NewOServiceHandlerForBOS(logger, oServiceService, oServiceServiceForBOS),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceServiceForBOS,
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "CHAT")
		authService := foodgroup.NewAuthService(cfg, sessionManager, sessionManager, feedbagStore, feedbagStore, chatRegistry, adjListBuddyListStore)
		oServiceService := foodgroup.NewOServiceService(cfg, sessionManager, feedbagStore, adjListBuddyListStore, logger)
		chatService := foodgroup.NewChatService(chatRegistry)
		oServiceServiceForChat := foodgroup.NewOServiceServiceForChat(*oServiceService, chatRegistry)

		oscar.ChatServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewChatRouter(handler.Handlers{
				ChatHandler:         handler.NewChatHandler(logger, chatService),
				OServiceChatHandler: handler.NewOServiceHandlerForChat(logger, oServiceService, oServiceServiceForChat),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceServiceForChat,
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "CHAT_NAV")
		authService := foodgroup.NewAuthService(cfg, sessionManager, sessionManager, feedbagStore, feedbagStore, chatRegistry, adjListBuddyListStore)
		oServiceService := foodgroup.NewOServiceService(cfg, sessionManager, feedbagStore, adjListBuddyListStore, logger)
		oServiceServiceForChatNav := foodgroup.NewOServiceServiceForChatNav(*oServiceService, chatRegistry)
		newChatSessMgr := func() foodgroup.SessionManager { return state.NewInMemorySessionManager(logger) }
		chatNavService := foodgroup.NewChatNavService(logger, chatRegistry, state.NewChatRoom, newChatSessMgr)

		oscar.ChatNavServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewChatNavRouter(handler.Handlers{
				ChatNavHandler:         handler.NewChatNavHandler(chatNavService, logger),
				OServiceChatNavHandler: handler.NewOServiceHandlerForChatNav(logger, oServiceService, oServiceServiceForChatNav),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceServiceForChatNav,
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "ALERT")
		authService := foodgroup.NewAuthService(cfg, sessionManager, sessionManager, feedbagStore, feedbagStore, chatRegistry, adjListBuddyListStore)
		oServiceService := foodgroup.NewOServiceService(cfg, sessionManager, feedbagStore, adjListBuddyListStore, logger)
		oServiceServiceForAlert := foodgroup.NewOServiceServiceForAlert(*oServiceService)

		oscar.AlertServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewAlertRouter(handler.Handlers{
				AlertHandler:         handler.NewAlertHandler(logger),
				OServiceAlertHandler: handler.NewOServiceHandlerForAlert(logger, oServiceService, oServiceServiceForAlert),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceServiceForAlert,
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "AUTH")
		authHandler := foodgroup.NewAuthService(cfg, sessionManager, nil, feedbagStore, feedbagStore, chatRegistry, adjListBuddyListStore)

		oscar.AuthServer{
			AuthService: authHandler,
			Config:      cfg,
			Logger:      logger,
		}.Start()
		wg.Done()
	}(logger)

	wg.Wait()
}
