package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mk6i/retro-aim-server/server/kerberos"
	"golang.org/x/time/rate"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/foodgroup"
	"github.com/mk6i/retro-aim-server/server/http"
	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/server/toc"
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
	Listeners              []config.Listener
}

// MakeCommonDeps creates common dependencies used by the food group services.
func MakeCommonDeps() (Container, error) {
	c := Container{}

	err := envconfig.Process("", &c.cfg)
	if err != nil {
		return c, fmt.Errorf("unable to process app config: %s\n", err.Error())
	}

	c.Listeners, err = config.ParseListenersCfg(c.cfg.BOSListeners, c.cfg.BOSAdvertisedHosts, c.cfg.KerberosListeners)
	if err != nil {
		return c, fmt.Errorf("unable to parse listener config: %s\n", err.Error())
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
func OSCAR(deps Container) *oscar.Server {
	logger := deps.logger.With("svc", "OSCAR")

	adminService := foodgroup.NewAdminService(
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.sqLiteUserStore,
		deps.inMemorySessionManager,
		deps.inMemorySessionManager,
		deps.logger,
	)
	authService := foodgroup.NewAuthService(
		deps.cfg,
		deps.inMemorySessionManager,
		deps.inMemorySessionManager,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.hmacCookieBaker,
		deps.chatSessionManager,
		deps.sqLiteUserStore,
		deps.rateLimitClasses,
	)
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

	if err := deps.sqLiteUserStore.ClearBuddyListRegistry(context.Background()); err != nil {
		panic(err)
	}

	return oscar.NewServer(
		authService,
		deps.sqLiteUserStore,
		deps.chatSessionManager,
		buddyService,
		logger,
		oServiceService,
		oscar.Handler{
			AdminService:      adminService,
			BARTService:       bartService,
			BuddyService:      buddyService,
			ChatNavService:    chatNavService,
			ChatService:       chatService,
			FeedbagService:    feedbagService,
			ICBMService:       icbmService,
			ICQService:        icqService,
			LocateService:     locateService,
			ODirService:       oDirService,
			OServiceService:   oServiceService,
			PermitDenyService: permitDenyService,
			StatsService:      statsService,
			UserLookupService: userLookupService,
			RouteLogger: middleware.RouteLogger{
				Logger: logger,
			},
		}.Handle,
		oServiceService,
		deps.snacRateLimits,
		oscar.NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
		deps.Listeners,
	)
}

// KerberosAPI creates an HTTP server for the Kerberos server.
func KerberosAPI(deps Container) *kerberos.Server {
	logger := deps.logger.With("svc", "kerberos")
	authService := foodgroup.NewAuthService(deps.cfg, deps.inMemorySessionManager, deps.inMemorySessionManager, deps.chatSessionManager, deps.sqLiteUserStore, deps.hmacCookieBaker, deps.chatSessionManager, deps.sqLiteUserStore, deps.rateLimitClasses)
	return kerberos.NewKerberosServer(deps.Listeners, logger, authService)
}

// MgmtAPI creates an HTTP server for the management API.
func MgmtAPI(deps Container) *http.Server {
	bld := config.Build{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	logger := deps.logger.With("svc", "API")
	return http.NewManagementAPI(bld, deps.cfg.APIListener, deps.sqLiteUserStore, deps.inMemorySessionManager, deps.sqLiteUserStore,
		deps.sqLiteUserStore, deps.chatSessionManager, deps.sqLiteUserStore, deps.inMemorySessionManager,
		deps.sqLiteUserStore, deps.sqLiteUserStore, deps.sqLiteUserStore, deps.sqLiteUserStore, logger)
}

// TOC creates a TOC server.
func TOC(deps Container) *toc.Server {
	logger := deps.logger.With("svc", "TOC")
	return toc.NewServer(
		strings.Split(deps.cfg.TOCListeners, ","),
		logger,
		toc.OSCARProxy{
			AdminService: foodgroup.NewAdminService(
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.inMemorySessionManager,
				deps.logger,
			),
			AuthService: foodgroup.NewAuthService(
				deps.cfg,
				deps.inMemorySessionManager,
				deps.inMemorySessionManager,
				deps.chatSessionManager,
				deps.sqLiteUserStore,
				deps.hmacCookieBaker,
				deps.chatSessionManager,
				deps.sqLiteUserStore,
				deps.rateLimitClasses,
			),
			BuddyListRegistry: deps.sqLiteUserStore,
			BuddyService: foodgroup.NewBuddyService(
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
			),
			CookieBaker:      deps.hmacCookieBaker,
			DirSearchService: foodgroup.NewODirService(logger, deps.sqLiteUserStore),
			ICBMService: foodgroup.NewICBMService(
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.snacRateLimits,
			),
			LocateService: foodgroup.NewLocateService(
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
			),
			Logger: logger,
			OServiceService: foodgroup.NewOServiceService(
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
			),
			PermitDenyService: foodgroup.NewPermitDenyService(
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.sqLiteUserStore,
				deps.inMemorySessionManager,
				deps.inMemorySessionManager,
			),
			TOCConfigStore:    deps.sqLiteUserStore,
			ChatService:       foodgroup.NewChatService(deps.chatSessionManager),
			ChatNavService:    foodgroup.NewChatNavService(logger, deps.sqLiteUserStore),
			SNACRateLimits:    deps.snacRateLimits,
			HTTPIPRateLimiter: toc.NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
		},
		toc.NewIPRateLimiter(rate.Every(1*time.Minute), 10, 1*time.Minute),
	)
}
