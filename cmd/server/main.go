package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"

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
		fmt.Printf("Config file (%s) not found, defaulting to env vars for app config...\n", *cfgFile)
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

	var g errgroup.Group

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g.Go(func() error {
		startMgmtAPI(ctx, cfg, feedbagStore, sessionManager, chatSessionManager, logger)
		return nil
	})
	g.Go(func() error {
		startBOS(ctx, logger, cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker)
		return nil
	})
	g.Go(func() error {
		startChat(ctx, logger, cfg, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker)
		return nil
	})
	g.Go(func() error {
		startChatNav(ctx, logger, cfg, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker)
		return nil
	})
	g.Go(func() error {
		startAlert(ctx, logger, cfg, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker)
		return nil
	})
	g.Go(func() error {
		startAdmin(ctx, logger, sessionManager, feedbagStore, adjListBuddyListStore, cfg, chatSessionManager, cookieBaker)
		return nil
	})
	g.Go(func() error {
		startBART(ctx, logger, feedbagStore, adjListBuddyListStore, cfg, chatSessionManager, cookieBaker)
		return nil
	})
	g.Go(func() error {
		startAuth(ctx, logger, cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker)
		return nil
	})
	g.Go(func() error {
		startODir(ctx, logger, cfg, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker)
		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("An error occurred: %v\n", err)
	}
}

func startODir(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	chatSessionManager *state.InMemoryChatSessionManager,
	feedbagStore *state.SQLiteUserStore,
	adjListBuddyListStore *state.AdjListBuddyListStore,
	cookieBaker state.HMACCookieBaker,
) {
	logger = logger.With("svc", "ODIR")
	sessionManager := state.NewInMemorySessionManager(logger)
	authService := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, sessionManager, feedbagStore, chatSessionManager, feedbagStore)
	oServiceService := foodgroup.NewOServiceServiceForODir(cfg, logger)

	oDirService := foodgroup.NewODirService(logger, feedbagStore)
	oscar.BOSServer{
		AuthService: authService,
		Config:      cfg,
		Handler: handler.NewODirRouter(handler.Handlers{
			OServiceHandler: handler.NewOServiceHandler(logger, oServiceService),
			ODirHandler:     handler.NewODirHandler(logger, oDirService),
		}),
		Logger:         logger,
		OnlineNotifier: oServiceService,
		ListenAddr:     net.JoinHostPort("", cfg.ODirPort),
	}.Start(ctx)
}

func startAuth(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	sessionManager *state.InMemorySessionManager,
	chatSessionManager *state.InMemoryChatSessionManager,
	feedbagStore *state.SQLiteUserStore,
	adjListBuddyListStore *state.AdjListBuddyListStore,
	cookieBaker state.HMACCookieBaker,
) {
	logger = logger.With("svc", "AUTH")
	authHandler := foodgroup.NewAuthService(cfg, sessionManager, chatSessionManager, feedbagStore, adjListBuddyListStore, cookieBaker, nil, nil, chatSessionManager, feedbagStore)

	oscar.AuthServer{
		AuthService: authHandler,
		Config:      cfg,
		Logger:      logger,
	}.Start(ctx)
}

func startBART(
	ctx context.Context,
	logger *slog.Logger,
	feedbagStore *state.SQLiteUserStore,
	adjListBuddyListStore *state.AdjListBuddyListStore,
	cfg config.Config,
	chatSessionManager *state.InMemoryChatSessionManager,
	cookieBaker state.HMACCookieBaker,
) {
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
	}.Start(ctx)
}

func startAdmin(
	ctx context.Context,
	logger *slog.Logger,
	sessionManager *state.InMemorySessionManager,
	feedbagStore *state.SQLiteUserStore,
	adjListBuddyListStore *state.AdjListBuddyListStore,
	cfg config.Config,
	chatSessionManager *state.InMemoryChatSessionManager,
	cookieBaker state.HMACCookieBaker,
) {
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
	}.Start(ctx)
}

func startAlert(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	chatSessionManager *state.InMemoryChatSessionManager,
	feedbagStore *state.SQLiteUserStore,
	adjListBuddyListStore *state.AdjListBuddyListStore,
	cookieBaker state.HMACCookieBaker,
) {
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
	}.Start(ctx)
}

func startChatNav(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	chatSessionManager *state.InMemoryChatSessionManager,
	feedbagStore *state.SQLiteUserStore,
	adjListBuddyListStore *state.AdjListBuddyListStore,
	cookieBaker state.HMACCookieBaker,
) {
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
	}.Start(ctx)
}

func startChat(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	chatSessionManager *state.InMemoryChatSessionManager,
	feedbagStore *state.SQLiteUserStore,
	adjListBuddyListStore *state.AdjListBuddyListStore,
	cookieBaker state.HMACCookieBaker,
) {
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
	}.Start(ctx)
}

func startBOS(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	sessionManager *state.InMemorySessionManager,
	chatSessionManager *state.InMemoryChatSessionManager,
	feedbagStore *state.SQLiteUserStore,
	adjListBuddyListStore *state.AdjListBuddyListStore,
	cookieBaker state.HMACCookieBaker,
) {
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
	}.Start(ctx)
}

func startMgmtAPI(
	ctx context.Context,
	cfg config.Config,
	feedbagStore *state.SQLiteUserStore,
	sessionManager *state.InMemorySessionManager,
	chatSessionManager *state.InMemoryChatSessionManager,
	logger *slog.Logger,
) {
	bld := config.Build{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	http.StartManagementAPI(ctx, bld, cfg, feedbagStore, sessionManager, feedbagStore, feedbagStore, chatSessionManager,
		feedbagStore, sessionManager, feedbagStore, feedbagStore, feedbagStore, feedbagStore, logger)
}
