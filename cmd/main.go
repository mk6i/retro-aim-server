package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/mkaminski/goaim/handler"
	"github.com/mkaminski/goaim/state"

	"github.com/kelseyhightower/envconfig"
	"github.com/mkaminski/goaim/server"
)

func main() {
	var cfg server.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to process app config: %s", err.Error())
		os.Exit(1)
	}

	feedbagStore, err := state.NewSQLiteUserStore(cfg.DBPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to create feedbag store: %s", err.Error())
		os.Exit(1)
	}

	logger := server.NewLogger(cfg)
	sessionManager := state.NewInMemorySessionManager(logger)
	chatRegistry := state.NewChatRegistry()

	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		server.StartManagementAPI(feedbagStore, logger)
		wg.Done()
	}()
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "BOS")
		authHandler := handler.NewAuthService(cfg, sessionManager, sessionManager, feedbagStore, feedbagStore, chatRegistry)
		buddyHandler := handler.NewBuddyService()
		oserviceHandler := handler.NewOServiceService(cfg, sessionManager, feedbagStore)
		oserviceBOSHandler := handler.NewOServiceServiceForBOS(*oserviceHandler, chatRegistry)
		locateHandler := handler.NewLocateService(sessionManager, feedbagStore, feedbagStore)
		newChatSessMgr := func() handler.SessionManager { return state.NewInMemorySessionManager(logger) }
		chatNavHandler := handler.NewChatNavService(logger, chatRegistry, state.NewChatRoom, newChatSessMgr)
		feedbagHandler := handler.NewFeedbagService(sessionManager, feedbagStore)
		icbmHandler := handler.NewICBMService(sessionManager, feedbagStore)

		server.BOSService{
			AuthHandler:       authHandler,
			OServiceBOSRouter: server.NewOServiceRouterForBOS(logger, oserviceHandler, oserviceBOSHandler),
			Config:            cfg,
			BOSRouter: server.BOSRootRouter{
				AlertRouter:       server.NewAlertRouter(logger),
				BuddyRouter:       server.NewBuddyRouter(logger, buddyHandler),
				ChatNavRouter:     server.NewChatNavRouter(chatNavHandler, logger),
				FeedbagRouter:     server.NewFeedbagRouter(logger, feedbagHandler),
				ICBMRouter:        server.NewICBMRouter(logger, icbmHandler),
				LocateRouter:      server.NewLocateRouter(locateHandler, logger),
				OServiceBOSRouter: server.NewOServiceRouterForBOS(logger, oserviceHandler, oserviceBOSHandler),
				Config:            cfg,
				RouteLogger: server.RouteLogger{
					Logger: logger,
				},
			},
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "CHAT")
		authHandler := handler.NewAuthService(cfg, sessionManager, sessionManager, feedbagStore, feedbagStore, chatRegistry)
		oserviceHandler := handler.NewOServiceService(cfg, sessionManager, feedbagStore)
		chatHandler := handler.NewChatService(chatRegistry)
		oserviceChatHandler := handler.NewOServiceServiceForChat(*oserviceHandler, chatRegistry)

		server.ChatService{
			AuthHandler:        authHandler,
			OServiceChatRouter: server.NewOServiceRouterForChat(logger, oserviceHandler, oserviceChatHandler),
			ChatRouter:         server.NewChatRouter(logger, chatHandler),
			Config:             cfg,
			RouteLogger: server.RouteLogger{
				Logger: logger,
			},
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "AUTH")
		authHandler := handler.NewAuthService(cfg, sessionManager, nil, feedbagStore, feedbagStore, chatRegistry)

		server.AuthService{
			AuthHandler: authHandler,
			Config:      cfg,
			RouteLogger: server.RouteLogger{
				Logger: logger,
			},
		}.Start()
		wg.Done()
	}(logger)

	wg.Wait()
}
