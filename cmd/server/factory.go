package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/time/rate"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/foodgroup"
	"github.com/mk6i/retro-aim-server/server/http"
	"github.com/mk6i/retro-aim-server/server/kerberos"
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

	if err := validateConfigMigration(); err != nil {
		return c, fmt.Errorf("unable to validate config migration: %s\n", err.Error())
	}

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

func validateConfigMigration() error {
	// Old environment variables that should be removed
	oldEnvVars := []string{
		"API_HOST",
		"API_PORT",
		"KERBEROS_PORT",
		"ALERT_PORT",
		"AUTH_PORT",
		"BART_PORT",
		"BOS_PORT",
		"CHAT_NAV_PORT",
		"CHAT_PORT",
		"ADMIN_PORT",
		"ODIR_PORT",
		"OSCAR_HOST",
		"TOC_HOST",
		"TOC_PORT",
	}

	// New environment variables that should be present
	newEnvVars := []string{
		"API_LISTENER",
		"OSCAR_ADVERTISED_LISTENERS",
		"OSCAR_LISTENERS",
		"KERBEROS_LISTENERS",
		"TOC_LISTENERS",
	}

	var oldEnvVarsFound []string
	var newEnvVarsMissing []string

	// Check for old environment variables that should be removed
	for _, envVar := range oldEnvVars {
		if os.Getenv(envVar) != "" {
			oldEnvVarsFound = append(oldEnvVarsFound, envVar)
		}
	}

	// Check for new environment variables that should be present
	for _, envVar := range newEnvVars {
		if os.Getenv(envVar) == "" {
			newEnvVarsMissing = append(newEnvVarsMissing, envVar)
		}
	}

	// If there are any issues, return an error with details
	if len(oldEnvVarsFound) > 0 || len(newEnvVarsMissing) > 0 {
		var errorMsg strings.Builder
		errorMsg.WriteString("Retro AIM Server v0.19.0 introduced some breaking configuration changes that you need to fix.\n")

		if len(oldEnvVarsFound) > 0 {
			errorMsg.WriteString("\nOld environment variables that must be removed:\n\n")
			for _, envVar := range oldEnvVarsFound {
				errorMsg.WriteString(fmt.Sprintf("  - %s\n", envVar))
			}
		}

		if len(newEnvVarsMissing) > 0 {
			errorMsg.WriteString("\nNew environment variables that must be provided:\n\n")
			for _, envVar := range newEnvVarsMissing {
				errorMsg.WriteString(fmt.Sprintf("  - %s\n", envVar))
			}

			// Generate export commands based on old environment variables
			errorMsg.WriteString("\nCopy/paste this updated configuration into your settings file:\n\n")

			if contains(newEnvVarsMissing, "API_LISTENER") {
				apiHost := getEnvOrDefault("API_HOST", "127.0.0.1")
				apiPort := getEnvOrDefault("API_PORT", "8080")
				errorMsg.WriteString(fmt.Sprintf("export API_LISTENER=%s:%s\n", apiHost, apiPort))
			}

			if contains(newEnvVarsMissing, "OSCAR_ADVERTISED_LISTENERS") {
				oscarHost := getEnvOrDefault("OSCAR_HOST", "127.0.0.1")
				authPort := getEnvOrDefault("AUTH_PORT", "5190")
				errorMsg.WriteString(fmt.Sprintf("export OSCAR_ADVERTISED_LISTENERS=EXTERNAL://%s:%s\n", oscarHost, authPort))
			}

			if contains(newEnvVarsMissing, "OSCAR_LISTENERS") {
				authPort := getEnvOrDefault("AUTH_PORT", "5190")
				errorMsg.WriteString(fmt.Sprintf("export OSCAR_LISTENERS=EXTERNAL://0.0.0.0:%s\n", authPort))
			}

			if contains(newEnvVarsMissing, "KERBEROS_LISTENERS") {
				kerberosPort := getEnvOrDefault("KERBEROS_PORT", "1088")
				errorMsg.WriteString(fmt.Sprintf("export KERBEROS_LISTENERS=EXTERNAL://0.0.0.0:%s\n", kerberosPort))
			}

			if contains(newEnvVarsMissing, "TOC_LISTENERS") {
				tocHost := getEnvOrDefault("TOC_HOST", "0.0.0.0")
				tocPort := getEnvOrDefault("TOC_PORT", "9898")
				errorMsg.WriteString(fmt.Sprintf("export TOC_LISTENERS=%s:%s\n", tocHost, tocPort))
			}
		}

		return errors.New(errorMsg.String())
	}

	return nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to get environment variable or return default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
