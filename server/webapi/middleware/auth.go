package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"

	"github.com/mk6i/retro-aim-server/state"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// ContextKeyAPIKey is the context key for storing the validated API key.
	ContextKeyAPIKey contextKey = "api_key"
	// ContextKeyDevID is the context key for storing the developer ID.
	ContextKeyDevID contextKey = "dev_id"
)

// APIKeyValidator defines methods for validating Web API keys.
type APIKeyValidator interface {
	// GetAPIKeyByDevKey retrieves and validates an API key by its dev_key value.
	GetAPIKeyByDevKey(ctx context.Context, devKey string) (*state.WebAPIKey, error)
	// UpdateLastUsed updates the last_used timestamp for an API key.
	UpdateLastUsed(ctx context.Context, devKey string) error
}

// RateLimiter manages per-devID rate limiting for the Web API.
type RateLimiter struct {
	limiters *cache.Cache
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with automatic cleanup.
func NewRateLimiter() *RateLimiter {
	// Create cache with 5 minute expiration and 10 minute cleanup interval
	c := cache.New(5*time.Minute, 10*time.Minute)
	return &RateLimiter{
		limiters: c,
	}
}

// Allow checks if a request from the given devID is allowed based on rate limits.
func (r *RateLimiter) Allow(devID string, limit int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get or create limiter for this devID
	var limiter *rate.Limiter
	if val, found := r.limiters.Get(devID); found {
		limiter = val.(*rate.Limiter)
	} else {
		// Create new limiter with burst equal to limit (allows initial burst)
		limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(limit)), limit)
		r.limiters.Set(devID, limiter, cache.DefaultExpiration)
	}

	return limiter.Allow()
}

// AuthMiddleware provides authentication and rate limiting for Web API endpoints.
type AuthMiddleware struct {
	Validator   APIKeyValidator
	RateLimiter *RateLimiter
	Logger      *slog.Logger
}

// NewAuthMiddleware creates a new authentication middleware instance.
func NewAuthMiddleware(validator APIKeyValidator, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		Validator:   validator,
		RateLimiter: NewRateLimiter(),
		Logger:      logger,
	}
}

// Authenticate is an HTTP middleware that validates API keys and enforces rate limits.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from 'k' parameter (query or form)
		apiKey := r.URL.Query().Get("k")
		if apiKey == "" {
			// Try form value for POST requests
			apiKey = r.FormValue("k")
		}

		if apiKey == "" {
			m.sendErrorResponse(w, http.StatusBadRequest, "required parameter 'k' is missing")
			return
		}

		// Validate API key
		ctx := r.Context()
		key, err := m.Validator.GetAPIKeyByDevKey(ctx, apiKey)
		if err != nil {
			if err == state.ErrNoAPIKey {
				m.Logger.DebugContext(ctx, "invalid API key attempted", "key", apiKey[:min(8, len(apiKey))]+"...")
				m.sendErrorResponse(w, http.StatusForbidden, "invalid API key")
				return
			}
			m.Logger.ErrorContext(ctx, "error validating API key", "err", err.Error())
			m.sendErrorResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		// Check if key is active
		if !key.IsActive {
			m.Logger.DebugContext(ctx, "inactive API key used", "dev_id", key.DevID)
			m.sendErrorResponse(w, http.StatusForbidden, "API key is inactive")
			return
		}

		// Check rate limit
		if !m.RateLimiter.Allow(key.DevID, key.RateLimit) {
			m.Logger.WarnContext(ctx, "rate limit exceeded", "dev_id", key.DevID, "limit", key.RateLimit)
			m.sendErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		// Update last used timestamp asynchronously
		go func() {
			if err := m.Validator.UpdateLastUsed(context.Background(), apiKey); err != nil {
				m.Logger.Error("failed to update last_used timestamp", "err", err.Error())
			}
		}()

		// Add API key info to context for use in handlers
		ctx = context.WithValue(ctx, ContextKeyAPIKey, key)
		ctx = context.WithValue(ctx, ContextKeyDevID, key.DevID)

		// Log the API request
		m.Logger.InfoContext(ctx, "API request authenticated",
			"dev_id", key.DevID,
			"app_name", key.AppName,
			"method", r.Method,
			"path", r.URL.Path,
		)

		// Pass to next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORSMiddleware handles CORS headers based on allowed origins for the API key.
func (m *AuthMiddleware) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from context (set by Authenticate middleware)
		key, ok := r.Context().Value(ContextKeyAPIKey).(*state.WebAPIKey)
		if !ok {
			// If no API key in context, this middleware is being used incorrectly
			m.Logger.Error("CORSMiddleware called without authentication context")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		if m.isOriginAllowed(origin, key.AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CapabilitiesMiddleware checks if the API key has the required capability for an endpoint.
func (m *AuthMiddleware) CapabilitiesMiddleware(requiredCapability string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from context
			key, ok := r.Context().Value(ContextKeyAPIKey).(*state.WebAPIKey)
			if !ok {
				m.Logger.Error("CapabilitiesMiddleware called without authentication context")
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// If no capabilities are defined, allow all (backward compatibility)
			if len(key.Capabilities) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Check if required capability is present
			hasCapability := false
			for _, cap := range key.Capabilities {
				if cap == requiredCapability || cap == "*" {
					hasCapability = true
					break
				}
			}

			if !hasCapability {
				m.Logger.WarnContext(r.Context(), "capability check failed",
					"dev_id", key.DevID,
					"required", requiredCapability,
					"available", key.Capabilities,
				)
				m.sendErrorResponse(w, http.StatusForbidden, fmt.Sprintf("missing required capability: %s", requiredCapability))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if an origin is in the allowed list.
func (m *AuthMiddleware) isOriginAllowed(origin string, allowedOrigins []string) bool {
	// If no origins specified, allow all (for backward compatibility/development)
	if len(allowedOrigins) == 0 {
		return true
	}

	origin = strings.ToLower(origin)
	for _, allowed := range allowedOrigins {
		allowed = strings.ToLower(allowed)
		
		// Exact match
		if origin == allowed {
			return true
		}
		
		// Wildcard support (e.g., "*.example.com")
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
		
		// Allow all origins (development only)
		if allowed == "*" {
			m.Logger.Warn("wildcard origin (*) used - should not be used in production")
			return true
		}
	}

	return false
}

// sendErrorResponse sends a JSON error response.
func (m *AuthMiddleware) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error": message,
		"code":  statusCode,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		m.Logger.Error("failed to encode error response", "err", err.Error())
	}
}

// GetAPIKeyFromContext retrieves the API key from the request context.
func GetAPIKeyFromContext(ctx context.Context) (*state.WebAPIKey, bool) {
	key, ok := ctx.Value(ContextKeyAPIKey).(*state.WebAPIKey)
	return key, ok
}

// GetDevIDFromContext retrieves the developer ID from the request context.
func GetDevIDFromContext(ctx context.Context) (string, bool) {
	devID, ok := ctx.Value(ContextKeyDevID).(string)
	return devID, ok
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

