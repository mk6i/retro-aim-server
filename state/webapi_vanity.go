package state

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

// VanityURL represents a user's vanity URL configuration.
type VanityURL struct {
	ScreenName   string     `json:"screenName"`
	VanityURL    string     `json:"vanityUrl"`
	DisplayName  string     `json:"displayName,omitempty"`
	Bio          string     `json:"bio,omitempty"`
	Location     string     `json:"location,omitempty"`
	Website      string     `json:"website,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	IsActive     bool       `json:"isActive"`
	ClickCount   int        `json:"clickCount"`
	LastAccessed *time.Time `json:"lastAccessed,omitempty"`
}

// VanityURLRedirect represents a vanity URL access record.
type VanityURLRedirect struct {
	ID         int64     `json:"id"`
	VanityURL  string    `json:"vanityUrl"`
	AccessedAt time.Time `json:"accessedAt"`
	IPAddress  string    `json:"ipAddress,omitempty"`
	UserAgent  string    `json:"userAgent,omitempty"`
	Referer    string    `json:"referer,omitempty"`
}

// VanityInfo represents the response for vanity URL lookups.
type VanityInfo struct {
	ScreenName  string                 `json:"screenName"`
	VanityURL   string                 `json:"vanityUrl"`
	DisplayName string                 `json:"displayName,omitempty"`
	Bio         string                 `json:"bio,omitempty"`
	Location    string                 `json:"location,omitempty"`
	Website     string                 `json:"website,omitempty"`
	ProfileURL  string                 `json:"profileUrl"`
	IsActive    bool                   `json:"isActive"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// VanityURLManager manages vanity URL operations.
type VanityURLManager struct {
	db       *sql.DB
	logger   *slog.Logger
	baseURL  string   // Base URL for the service (e.g., "https://aim.example.com")
	reserved []string // Reserved URLs that cannot be claimed
}

// NewVanityURLManager creates a new vanity URL manager.
func NewVanityURLManager(db *sql.DB, logger *slog.Logger, baseURL string) *VanityURLManager {
	return &VanityURLManager{
		db:      db,
		logger:  logger,
		baseURL: baseURL,
		reserved: []string{
			"api", "admin", "help", "support", "about", "terms", "privacy",
			"login", "logout", "register", "signup", "signin", "settings",
			"profile", "user", "users", "aim", "aol", "webapi", "oscar",
			"chat", "im", "message", "buddy", "buddies", "feed", "rss",
		},
	}
}

// CreateOrUpdateVanityURL creates or updates a vanity URL for a user.
func (m *VanityURLManager) CreateOrUpdateVanityURL(ctx context.Context, screenName string, vanityURL string, info map[string]interface{}) error {
	// Validate vanity URL
	if err := m.validateVanityURL(vanityURL); err != nil {
		return err
	}

	// Check if URL is reserved
	if m.isReserved(vanityURL) {
		return fmt.Errorf("vanity URL '%s' is reserved", vanityURL)
	}

	// Extract optional fields from info
	displayName, _ := info["displayName"].(string)
	bio, _ := info["bio"].(string)
	location, _ := info["location"].(string)
	website, _ := info["website"].(string)

	now := time.Now()

	// Try to update existing record first
	updateQuery := `
		UPDATE vanity_urls
		SET vanity_url = ?, display_name = ?, bio = ?, location = ?, 
		    website = ?, updated_at = ?, is_active = ?
		WHERE screen_name = ?
	`

	result, err := m.db.ExecContext(ctx, updateQuery,
		vanityURL, displayName, bio, location, website,
		now.Unix(), true, screenName,
	)

	if err != nil {
		return fmt.Errorf("failed to update vanity URL: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		m.logger.InfoContext(ctx, "updated vanity URL",
			"screenName", screenName,
			"vanityURL", vanityURL,
		)
		return nil
	}

	// Insert new record
	insertQuery := `
		INSERT INTO vanity_urls (
			screen_name, vanity_url, display_name, bio, location,
			website, created_at, updated_at, is_active, click_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = m.db.ExecContext(ctx, insertQuery,
		screenName, vanityURL, displayName, bio, location,
		website, now.Unix(), now.Unix(), true, 0,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("vanity URL '%s' is already taken", vanityURL)
		}
		return fmt.Errorf("failed to create vanity URL: %w", err)
	}

	m.logger.InfoContext(ctx, "created vanity URL",
		"screenName", screenName,
		"vanityURL", vanityURL,
	)

	return nil
}

// GetVanityInfo retrieves vanity URL information.
func (m *VanityURLManager) GetVanityInfo(ctx context.Context, vanityURL string) (*VanityInfo, error) {
	// Clean the vanity URL
	vanityURL = strings.ToLower(strings.TrimSpace(vanityURL))

	query := `
		SELECT screen_name, vanity_url, display_name, bio, location,
		       website, created_at, updated_at, is_active, click_count, last_accessed
		FROM vanity_urls
		WHERE vanity_url = ? AND is_active = 1
	`

	var v VanityURL
	var createdAt, updatedAt int64
	var lastAccessed sql.NullInt64

	err := m.db.QueryRowContext(ctx, query, vanityURL).Scan(
		&v.ScreenName, &v.VanityURL, &v.DisplayName, &v.Bio, &v.Location,
		&v.Website, &createdAt, &updatedAt, &v.IsActive, &v.ClickCount, &lastAccessed,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("vanity URL not found: %s", vanityURL)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get vanity info: %w", err)
	}

	v.CreatedAt = time.Unix(createdAt, 0)
	v.UpdatedAt = time.Unix(updatedAt, 0)
	if lastAccessed.Valid {
		t := time.Unix(lastAccessed.Int64, 0)
		v.LastAccessed = &t
	}

	// Create response info
	info := &VanityInfo{
		ScreenName:  v.ScreenName,
		VanityURL:   v.VanityURL,
		DisplayName: v.DisplayName,
		Bio:         v.Bio,
		Location:    v.Location,
		Website:     v.Website,
		ProfileURL:  m.buildProfileURL(v.VanityURL),
		IsActive:    v.IsActive,
		Extra: map[string]interface{}{
			"createdAt":  v.CreatedAt.Unix(),
			"clickCount": v.ClickCount,
		},
	}

	// Update click count and last accessed asynchronously
	go m.recordAccess(context.Background(), vanityURL)

	return info, nil
}

// GetVanityInfoByScreenName retrieves vanity URL info by screen name.
func (m *VanityURLManager) GetVanityInfoByScreenName(ctx context.Context, screenName string) (*VanityInfo, error) {
	query := `
		SELECT screen_name, vanity_url, display_name, bio, location,
		       website, created_at, updated_at, is_active, click_count, last_accessed
		FROM vanity_urls
		WHERE screen_name = ? AND is_active = 1
	`

	var v VanityURL
	var createdAt, updatedAt int64
	var lastAccessed sql.NullInt64

	err := m.db.QueryRowContext(ctx, query, screenName).Scan(
		&v.ScreenName, &v.VanityURL, &v.DisplayName, &v.Bio, &v.Location,
		&v.Website, &createdAt, &updatedAt, &v.IsActive, &v.ClickCount, &lastAccessed,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No vanity URL configured
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get vanity info: %w", err)
	}

	v.CreatedAt = time.Unix(createdAt, 0)
	v.UpdatedAt = time.Unix(updatedAt, 0)
	if lastAccessed.Valid {
		t := time.Unix(lastAccessed.Int64, 0)
		v.LastAccessed = &t
	}

	// Create response info
	info := &VanityInfo{
		ScreenName:  v.ScreenName,
		VanityURL:   v.VanityURL,
		DisplayName: v.DisplayName,
		Bio:         v.Bio,
		Location:    v.Location,
		Website:     v.Website,
		ProfileURL:  m.buildProfileURL(v.VanityURL),
		IsActive:    v.IsActive,
		Extra: map[string]interface{}{
			"createdAt":  v.CreatedAt.Unix(),
			"clickCount": v.ClickCount,
		},
	}

	return info, nil
}

// DeleteVanityURL removes a user's vanity URL.
func (m *VanityURLManager) DeleteVanityURL(ctx context.Context, screenName string) error {
	query := `UPDATE vanity_urls SET is_active = 0, updated_at = ? WHERE screen_name = ?`

	_, err := m.db.ExecContext(ctx, query, time.Now().Unix(), screenName)
	if err != nil {
		return fmt.Errorf("failed to delete vanity URL: %w", err)
	}

	m.logger.InfoContext(ctx, "deleted vanity URL", "screenName", screenName)
	return nil
}

// CheckAvailability checks if a vanity URL is available.
func (m *VanityURLManager) CheckAvailability(ctx context.Context, vanityURL string) (bool, error) {
	// Validate format
	if err := m.validateVanityURL(vanityURL); err != nil {
		return false, err
	}

	// Check if reserved
	if m.isReserved(vanityURL) {
		return false, nil
	}

	// Check database
	query := `SELECT COUNT(*) FROM vanity_urls WHERE vanity_url = ? AND is_active = 1`

	var count int
	err := m.db.QueryRowContext(ctx, query, vanityURL).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check availability: %w", err)
	}

	return count == 0, nil
}

// GetPopularVanityURLs retrieves the most accessed vanity URLs.
func (m *VanityURLManager) GetPopularVanityURLs(ctx context.Context, limit int) ([]VanityInfo, error) {
	query := `
		SELECT screen_name, vanity_url, display_name, bio, location,
		       website, is_active, click_count
		FROM vanity_urls
		WHERE is_active = 1
		ORDER BY click_count DESC
		LIMIT ?
	`

	rows, err := m.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular vanity URLs: %w", err)
	}
	defer rows.Close()

	var results []VanityInfo
	for rows.Next() {
		var info VanityInfo
		var displayName, bio, location, website sql.NullString

		err := rows.Scan(
			&info.ScreenName, &info.VanityURL, &displayName, &bio,
			&location, &website, &info.IsActive, &info.Extra,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vanity info: %w", err)
		}

		if displayName.Valid {
			info.DisplayName = displayName.String
		}
		if bio.Valid {
			info.Bio = bio.String
		}
		if location.Valid {
			info.Location = location.String
		}
		if website.Valid {
			info.Website = website.String
		}

		info.ProfileURL = m.buildProfileURL(info.VanityURL)
		results = append(results, info)
	}

	return results, nil
}

// recordAccess records a vanity URL access.
func (m *VanityURLManager) recordAccess(ctx context.Context, vanityURL string) {
	// Update click count and last accessed time
	updateQuery := `
		UPDATE vanity_urls
		SET click_count = click_count + 1, last_accessed = ?
		WHERE vanity_url = ?
	`

	_, err := m.db.ExecContext(ctx, updateQuery, time.Now().Unix(), vanityURL)
	if err != nil {
		m.logger.Error("failed to record vanity URL access", "error", err, "vanityURL", vanityURL)
	}
}

// LogRedirect logs a vanity URL redirect for analytics.
func (m *VanityURLManager) LogRedirect(ctx context.Context, redirect VanityURLRedirect) error {
	query := `
		INSERT INTO vanity_url_redirects (vanity_url, accessed_at, ip_address, user_agent, referer)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := m.db.ExecContext(ctx, query,
		redirect.VanityURL, redirect.AccessedAt.Unix(),
		redirect.IPAddress, redirect.UserAgent, redirect.Referer,
	)

	if err != nil {
		return fmt.Errorf("failed to log redirect: %w", err)
	}

	return nil
}

// validateVanityURL validates the format of a vanity URL.
func (m *VanityURLManager) validateVanityURL(vanityURL string) error {
	// Clean and lowercase
	vanityURL = strings.ToLower(strings.TrimSpace(vanityURL))

	// Check length
	if len(vanityURL) < 3 || len(vanityURL) > 30 {
		return fmt.Errorf("vanity URL must be between 3 and 30 characters")
	}

	// Check format (alphanumeric, hyphens, underscores only)
	validFormat := regexp.MustCompile(`^[a-z0-9_-]+$`)
	if !validFormat.MatchString(vanityURL) {
		return fmt.Errorf("vanity URL can only contain letters, numbers, hyphens, and underscores")
	}

	// Can't start or end with special characters
	if strings.HasPrefix(vanityURL, "-") || strings.HasPrefix(vanityURL, "_") ||
		strings.HasSuffix(vanityURL, "-") || strings.HasSuffix(vanityURL, "_") {
		return fmt.Errorf("vanity URL cannot start or end with hyphens or underscores")
	}

	return nil
}

// isReserved checks if a vanity URL is in the reserved list.
func (m *VanityURLManager) isReserved(vanityURL string) bool {
	vanityURL = strings.ToLower(vanityURL)
	for _, reserved := range m.reserved {
		if vanityURL == reserved {
			return true
		}
	}
	return false
}

// buildProfileURL builds the full profile URL for a vanity URL.
func (m *VanityURLManager) buildProfileURL(vanityURL string) string {
	if m.baseURL == "" {
		return fmt.Sprintf("/profile/%s", vanityURL)
	}
	return fmt.Sprintf("%s/profile/%s", strings.TrimRight(m.baseURL, "/"), vanityURL)
}
