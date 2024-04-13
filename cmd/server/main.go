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

	feedbagStore, err := state.NewSQLiteUserStore(cfg.DBPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to create feedbag store: %s", err.Error())
		os.Exit(1)
	}

	logger := middleware.NewLogger(cfg)
	sessionManager := state.NewInMemorySessionManager(logger)
	chatRegistry := state.NewChatRegistry()

	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		http.StartManagementAPI(feedbagStore, sessionManager, logger)
		wg.Done()
	}()
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "BOS")
		authService := foodgroup.NewAuthService(cfg, sessionManager, sessionManager, feedbagStore, feedbagStore, chatRegistry)
		bartService := foodgroup.NewBARTService(logger, feedbagStore, sessionManager, feedbagStore)
		buddyService := foodgroup.NewBuddyService()
		oServiceService := foodgroup.NewOServiceService(cfg, sessionManager, feedbagStore)
		oServiceServiceForBOS := foodgroup.NewOServiceServiceForBOS(*oServiceService, chatRegistry)
		locateService := foodgroup.NewLocateService(sessionManager, feedbagStore, feedbagStore)
		newChatSessMgr := func() foodgroup.SessionManager { return state.NewInMemorySessionManager(logger) }
		chatNavService := foodgroup.NewChatNavService(logger, chatRegistry, state.NewChatRoom, newChatSessMgr)
		feedbagService := foodgroup.NewFeedbagService(logger, sessionManager, feedbagStore, feedbagStore)
		icbmService := foodgroup.NewICBMService(sessionManager, feedbagStore)

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
		authService := foodgroup.NewAuthService(cfg, sessionManager, sessionManager, feedbagStore, feedbagStore, chatRegistry)
		oServiceService := foodgroup.NewOServiceService(cfg, sessionManager, feedbagStore)
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
		logger = logger.With("svc", "AUTH")
		authHandler := foodgroup.NewAuthService(cfg, sessionManager, nil, feedbagStore, feedbagStore, chatRegistry)

		oscar.BUCPAuthService{
			AuthService: authHandler,
			Config:      cfg,
			Logger:      logger,
		}.Start()
		wg.Done()
	}(logger)

	wg.Wait()
}
