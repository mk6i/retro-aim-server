package main

import (
	"fmt"
	"github.com/mkaminski/goaim/handler"
	"os"
	"sync"

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

	fm, err := server.NewSQLiteFeedbagStore(cfg.DBPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to create feedbag store: %s", err.Error())
		os.Exit(1)
	}

	logger := server.NewLogger(cfg)
	sm := server.NewSessionManager(logger)
	cr := server.NewChatRegistry()

	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		server.StartManagementAPI(fm, logger)
		wg.Done()
	}()
	go func() {
		authHandler := handler.NewAuthService(sm, fm, fm, cfg)
		buddyHandler := handler.NewBuddyService()
		oserviceHandler := handler.NewOServiceService(cfg, sm, fm)
		oserviceBOSHandler := handler.NewOServiceServiceForBOS(*oserviceHandler, cr)
		locateHandler := handler.NewLocateService(sm, fm, fm)
		chatNavHandler := handler.NewChatNavService(logger, cr)
		feedbagHandler := handler.NewFeedbagService(sm, fm)
		icbmHandler := handler.NewICBMService(sm, fm)

		router := server.BOSServiceRouter{
			AlertRouter:       server.NewAlertRouter(logger),
			AuthHandler:       authHandler,
			BuddyRouter:       server.NewBuddyRouter(logger, buddyHandler),
			ChatNavRouter:     server.NewChatNavRouter(chatNavHandler, logger),
			FeedbagRouter:     server.NewFeedbagRouter(logger, feedbagHandler),
			ICBMRouter:        server.NewICBMRouter(logger, icbmHandler),
			LocateRouter:      server.NewLocateRouter(locateHandler, logger),
			OServiceBOSRouter: server.NewOServiceRouterForBOS(logger, oserviceHandler, oserviceBOSHandler),
			Cfg:               cfg,
			RouteLogger: server.RouteLogger{
				Logger: logger,
			},
			NewChatSessMgr: func() server.ChatSessionManager { return server.NewSessionManager(logger) },
		}

		server.ListenBOS(cfg, router, authHandler, logger.With("svc", "BOS"))
		wg.Done()
	}()
	go func() {
		authHandler := handler.NewAuthService(sm, fm, fm, cfg)
		oserviceHandler := handler.NewOServiceService(cfg, sm, fm)
		chatHandler := handler.NewChatService()
		oserviceChatHandler := handler.NewOServiceServiceForChat(*oserviceHandler)
		router := server.NewChatServiceRouter(logger, cfg, oserviceHandler, *chatHandler, oserviceChatHandler)
		server.ListenChat(cfg, router, cr, authHandler, logger.With("svc", "CHAT"))
		wg.Done()
	}()
	go func() {
		authHandler := handler.NewAuthService(sm, fm, fm, cfg)
		server.ListenBUCPLogin(cfg, err, logger, authHandler)
		wg.Done()
	}()

	wg.Wait()
}
