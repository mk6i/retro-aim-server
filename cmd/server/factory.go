package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/time/rate"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/foodgroup"
	"github.com/mk6i/retro-aim-server/server/http"
	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
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

// OSCAR creates an OSCAR server for the OSCAR food group.
func OSCAR(deps Container) oscar.Server {
	logger := deps.logger.With("svc", "OSCAR")

	authService := foodgroup.NewAuthService(deps.cfg, deps.inMemorySessionManager, deps.inMemorySessionManager, deps.chatSessionManager, deps.sqLiteUserStore, deps.hmacCookieBaker, deps.chatSessionManager, deps.sqLiteUserStore, deps.rateLimitClasses)
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
	chatService := foodgroup.NewChatService(deps.chatSessionManager)
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
	oServiceService := foodgroup.NewOServiceService(
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
		deps.chatSessionManager,
	)
	userLookupService := foodgroup.NewUserLookupService(deps.sqLiteUserStore)
	statsService := foodgroup.NewStatsService()
	oDirService := foodgroup.NewODirService(logger, deps.sqLiteUserStore)

	return oscar.Server{
		AuthService:        authService,
		BuddyListRegistry:  deps.sqLiteUserStore,
		DepartureNotifier:  buddyService,
		ChatSessionManager: deps.chatSessionManager,
		Handler: oscar.Router{
			AlertHandler:      oscar.NewAlertHandler(logger),
			BARTHandler:       oscar.NewBARTHandler(logger, bartService),
			BuddyHandler:      oscar.NewBuddyHandler(logger, buddyService),
			ChatHandler:       oscar.NewChatHandler(logger, chatService),
			ChatNavHandler:    oscar.NewChatNavHandler(chatNavService, logger),
			FeedbagHandler:    oscar.NewFeedbagHandler(logger, feedbagService),
			ICBMHandler:       oscar.NewICBMHandler(logger, icbmService),
			ICQHandler:        oscar.NewICQHandler(logger, icqService),
			LocateHandler:     oscar.NewLocateHandler(locateService, logger),
			ODirHandler:       oscar.NewODirHandler(logger, oDirService),
			OServiceHandler:   oscar.NewOServiceHandler(logger, oServiceService),
			PermitDenyHandler: oscar.NewPermitDenyHandler(logger, permitDenyService),
			StatsHandler:      oscar.NewStatsHandler(logger, statsService),
			UserLookupHandler: oscar.NewUserLookupHandler(logger, userLookupService),
		}.Handle,
		IPRateLimiter: oscar.NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
		Listeners: []oscar.Listener{
			{
				Hostname:      "0.0.0.0",
				Port:          "5190",
				AdvertiseHost: "127.0.0.1",
				AdvertisePort: "5190",
			},
			{
				Hostname:      "0.0.0.0",
				Port:          "5191",
				AdvertiseHost: "127.0.0.1",
				AdvertisePort: "5191",
			},
		},
		Logger:           logger,
		OnlineNotifier:   oServiceService,
		RateLimitUpdater: oServiceService,
		SNACRateLimits:   deps.snacRateLimits,
	}
}

// KerberosAPI creates an HTTP server for the Kerberos server.
func KerberosAPI(deps Container) *oscar.KerberosServer {
	authService := foodgroup.NewAuthService(deps.cfg, deps.inMemorySessionManager, deps.inMemorySessionManager, deps.chatSessionManager, deps.sqLiteUserStore, deps.hmacCookieBaker, deps.chatSessionManager, deps.sqLiteUserStore, deps.rateLimitClasses)
	return oscar.NewKerberosServer(deps.cfg, deps.logger, authService)
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

// TOC creates a TOC server.
//func TOC(deps Container) toc.Server {
//	logger := deps.logger.With("svc", "TOC")
//	sessionManager := state.NewInMemorySessionManager(logger)
//	return toc.Server{
//		Logger:     logger,
//		ListenAddr: net.JoinHostPort(deps.cfg.TOCHost, deps.cfg.TOCPort),
//		BOSProxy: toc.OSCARProxy{
//			AdminService: foodgroup.NewAdminService(
//				deps.sqLiteUserStore,
//				deps.sqLiteUserStore,
//				deps.sqLiteUserStore,
//				deps.inMemorySessionManager,
//				deps.inMemorySessionManager,
//				deps.logger,
//			),
//			AuthService: foodgroup.NewAuthService(
//				deps.cfg,
//				deps.inMemorySessionManager,
//				deps.chatSessionManager,
//				deps.sqLiteUserStore,
//				deps.hmacCookieBaker,
//				deps.chatSessionManager,
//				deps.sqLiteUserStore,
//				nil,
//				deps.rateLimitClasses,
//			),
//			BuddyListRegistry: deps.sqLiteUserStore,
//			BuddyService: foodgroup.NewBuddyService(
//				deps.inMemorySessionManager,
//				deps.sqLiteUserStore,
//				deps.sqLiteUserStore,
//				deps.inMemorySessionManager,
//				deps.sqLiteUserStore,
//			),
//			CookieBaker:      deps.hmacCookieBaker,
//			DirSearchService: foodgroup.NewODirService(logger, deps.sqLiteUserStore),
//			ICBMService: foodgroup.NewICBMService(
//				deps.sqLiteUserStore,
//				deps.inMemorySessionManager,
//				deps.sqLiteUserStore,
//				deps.sqLiteUserStore,
//				deps.inMemorySessionManager,
//				deps.snacRateLimits,
//			),
//			LocateService: foodgroup.NewLocateService(
//				deps.sqLiteUserStore,
//				deps.inMemorySessionManager,
//				deps.sqLiteUserStore,
//				deps.sqLiteUserStore,
//				deps.inMemorySessionManager,
//			),
//			Logger: logger,
//			OServiceServiceBOS: foodgroup.NewOServiceServiceForBOS(
//				deps.cfg,
//				deps.inMemorySessionManager,
//				logger,
//				deps.hmacCookieBaker,
//				deps.sqLiteUserStore,
//				deps.sqLiteUserStore,
//				deps.inMemorySessionManager,
//				deps.sqLiteUserStore,
//				deps.rateLimitClasses,
//				deps.snacRateLimits,
//			),
//			PermitDenyService: foodgroup.NewPermitDenyService(
//				deps.sqLiteUserStore,
//				deps.sqLiteUserStore,
//				deps.sqLiteUserStore,
//				deps.inMemorySessionManager,
//				deps.inMemorySessionManager,
//			),
//			TOCConfigStore: deps.sqLiteUserStore,
//			ChatService:    foodgroup.NewChatService(deps.chatSessionManager),
//			OServiceServiceChat: foodgroup.NewOServiceServiceForChat(
//				deps.cfg,
//				logger,
//				sessionManager,
//				deps.sqLiteUserStore,
//				deps.chatSessionManager,
//				deps.sqLiteUserStore,
//				sessionManager,
//				deps.sqLiteUserStore,
//				deps.rateLimitClasses,
//				deps.snacRateLimits,
//			),
//			ChatNavService:    foodgroup.NewChatNavService(logger, deps.sqLiteUserStore),
//			SNACRateLimits:    deps.snacRateLimits,
//			HTTPIPRateLimiter: toc.NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
//		},
//		LoginIPRateLimiter: toc.NewIPRateLimiter(rate.Every(10*time.Minute), 10, 20*time.Minute),
//	}
//}
