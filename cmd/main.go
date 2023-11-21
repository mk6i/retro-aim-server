package main

import (
	"fmt"
	"github.com/mkaminski/goaim/handler"
	"github.com/mkaminski/goaim/state"
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

	fm, err := state.NewSQLiteFeedbagStore(cfg.DBPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to create feedbag store: %s", err.Error())
		os.Exit(1)
	}

	logger := server.NewLogger(cfg)
	sm := state.NewSessionManager(logger)
	cr := state.NewChatRegistry()

	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		server.StartManagementAPI(fm, logger)
		wg.Done()
	}()
	go func() {
		authHandler := handler.NewAuthService(cfg, sm, fm, fm, cr)
		buddyHandler := handler.NewBuddyService()
		oserviceHandler := handler.NewOServiceService(cfg, sm, fm)
		oserviceBOSHandler := handler.NewOServiceServiceForBOS(*oserviceHandler, cr)
		locateHandler := handler.NewLocateService(sm, fm, fm)
		newChatSessMgr := func() handler.ChatSessionManager { return state.NewSessionManager(logger) }
		chatNavHandler := handler.NewChatNavService(logger, cr, handler.NewChatRoom, newChatSessMgr)
		feedbagHandler := handler.NewFeedbagService(sm, fm)
		icbmHandler := handler.NewICBMService(sm, fm)

		bosService := server.BOSService{
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
		}
		server.ListenBOS(cfg, bosService, logger.With("svc", "BOS"))
		wg.Done()
	}()
	go func() {
		authHandler := handler.NewAuthService(cfg, sm, fm, fm, cr)
		oserviceHandler := handler.NewOServiceService(cfg, sm, fm)
		chatHandler := handler.NewChatService(cr)
		oserviceChatHandler := handler.NewOServiceServiceForChat(*oserviceHandler, cr)

		chatService := server.ChatService{
			AuthHandler:        authHandler,
			OServiceChatRouter: server.NewOServiceRouterForChat(logger, oserviceHandler, oserviceChatHandler),
			ChatRouter:         server.NewChatRouter(logger, chatHandler),
			Config:             cfg,
			RouteLogger: server.RouteLogger{
				Logger: logger,
			},
		}
		server.ListenChat(cfg, chatService, logger.With("svc", "CHAT"))
		wg.Done()
	}()
	go func() {
		authHandler := handler.NewAuthService(cfg, sm, fm, fm, cr)
		server.ListenBUCPLogin(cfg, err, logger, authHandler)
		wg.Done()
	}()

	wg.Wait()
}
