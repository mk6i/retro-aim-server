package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/foodgroup"
	"github.com/mk6i/retro-aim-server/server/http"
	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/server/oscar/handler"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
)

// Default build fields are populated by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	cfgFile := flag.String("config", "settings.env", "Path to config file")
	showHelp := flag.Bool("help", false, "Display help")
	showVersion := flag.Bool("version", false, "Display build information")

	flag.Parse()

	switch {
	case *showVersion:
		fmt.Printf("%-10s %s\n", "version:", version)
		fmt.Printf("%-10s %s\n", "commit:", commit)
		fmt.Printf("%-10s %s\n", "date:", date)
		os.Exit(0)
	case *showHelp:
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Optionally populate environment variables with config file
	if err := godotenv.Load(*cfgFile); err != nil {
		fmt.Printf("Config file not found (%s). Moving on...\n", *cfgFile)
	} else {
		fmt.Printf("Successfully loaded config file (%s)\n", *cfgFile)
	}
}

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
	chatSessionManager := state.NewInMemoryChatSessionManager(logger)
	adjListBuddyListStore := state.NewAdjListBuddyListStore()

	wg := sync.WaitGroup{}
	wg.Add(7)

	go func() {
		bld := config.Build{
			Version: version,
			Commit:  commit,
			Date:    date,
		}
		http.StartManagementAPI(bld, cfg, feedbagStore, sessionManager, feedbagStore, feedbagStore, chatSessionManager, sessionManager, feedbagStore, feedbagStore, feedbagStore, feedbagStore, logger)
		wg.Done()
	}()
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "BOS")
		authService := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, sessionManager, feedbagStore, chatSessionManager, feedbagStore)
		bartService := foodgroup.NewBARTService(logger, feedbagStore, sessionManager, feedbagStore, adjListBuddyListStore)
		buddyService := foodgroup.NewBuddyService(sessionManager, feedbagStore, adjListBuddyListStore)
		chatNavService := foodgroup.NewChatNavService(logger, feedbagStore)
		feedbagService := foodgroup.NewFeedbagService(logger, sessionManager, feedbagStore, feedbagStore, adjListBuddyListStore)
		foodgroupService := foodgroup.NewPermitDenyService()
		icbmService := foodgroup.NewICBMService(sessionManager, feedbagStore, adjListBuddyListStore, feedbagStore)
		icqService := foodgroup.NewICQService(sessionManager, feedbagStore, feedbagStore, logger, sessionManager, feedbagStore)
		locateService := foodgroup.NewLocateService(sessionManager, feedbagStore, feedbagStore, adjListBuddyListStore)
		oServiceService := foodgroup.NewOServiceServiceForBOS(cfg, sessionManager, adjListBuddyListStore, logger, cookieBaker, feedbagStore, feedbagStore)

		oscar.BOSServer{
			AuthService: authService,
			Config:      cfg,
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
				PermitDenyHandler: handler.NewPermitDenyHandler(logger, foodgroupService),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceService,
			ListenAddr:     net.JoinHostPort("", cfg.BOSPort),
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "CHAT")
		sessionManager := state.NewInMemorySessionManager(logger)
		authService := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, sessionManager, feedbagStore, chatSessionManager, feedbagStore)
		chatService := foodgroup.NewChatService(chatSessionManager)
		oServiceService := foodgroup.NewOServiceServiceForChat(cfg, logger, sessionManager, adjListBuddyListStore, feedbagStore, feedbagStore, chatSessionManager)

		oscar.ChatServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewChatRouter(handler.Handlers{
				ChatHandler:     handler.NewChatHandler(logger, chatService),
				OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceService,
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "CHAT_NAV")
		sessionManager := state.NewInMemorySessionManager(logger)
		authService := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, sessionManager, feedbagStore, chatSessionManager, feedbagStore)
		chatNavService := foodgroup.NewChatNavService(logger, feedbagStore)
		oServiceService := foodgroup.NewOServiceServiceForChatNav(cfg, logger, sessionManager, adjListBuddyListStore, feedbagStore)

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
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "ALERT")
		sessionManager := state.NewInMemorySessionManager(logger)
		authService := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, sessionManager, feedbagStore, chatSessionManager, feedbagStore)
		oServiceService := foodgroup.NewOServiceServiceForAlert(cfg, logger, sessionManager, adjListBuddyListStore, feedbagStore)

		oscar.BOSServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewAlertRouter(handler.Handlers{
				AlertHandler:    handler.NewAlertHandler(logger),
				OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceService,
			ListenAddr:     net.JoinHostPort("", cfg.AlertPort),
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "ADMIN")
		buddyService := foodgroup.NewBuddyService(sessionManager, feedbagStore, adjListBuddyListStore)
		adminService := foodgroup.NewAdminService(sessionManager, feedbagStore, buddyService, sessionManager)
		authService := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, sessionManager, feedbagStore, chatSessionManager, feedbagStore)
		oServiceService := foodgroup.NewOServiceServiceForAdmin(cfg, logger, buddyService)

		oscar.AdminServer{
			AuthService: authService,
			Config:      cfg,
			Handler: handler.NewAdminRouter(handler.Handlers{
				AdminHandler:    handler.NewAdminHandler(logger, adminService),
				OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			}),
			Logger:         logger,
			OnlineNotifier: oServiceService,
			ListenAddr:     net.JoinHostPort("", cfg.AdminPort),
		}.Start()
		wg.Done()
	}(logger)
	go func(logger *slog.Logger) {
		logger = logger.With("svc", "BART")
		sessionManager := state.NewInMemorySessionManager(logger)
		bartService := foodgroup.NewBARTService(logger, feedbagStore, sessionManager, feedbagStore, adjListBuddyListStore)
		authService := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, sessionManager, feedbagStore, chatSessionManager, feedbagStore)
		oServiceService := foodgroup.NewOServiceServiceForBART(cfg, logger, sessionManager, adjListBuddyListStore, feedbagStore)

		oscar.BOSServer{
			AuthService: authService,
			Config:      cfg,
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
		authHandler := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, nil, nil, chatSessionManager, feedbagStore)

		oscar.AuthServer{
			AuthService: authHandler,
			Config:      cfg,
			Logger:      logger,
		}.Start()
		wg.Done()
	}(logger)

	wg.Wait()
}
