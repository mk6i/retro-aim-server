package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrDupAPIKey is returned when attempting to insert a duplicate API key.
	ErrDupAPIKey = errors.New("API key already exists")
	// ErrNoAPIKey is returned when an API key is not found.
	ErrNoAPIKey = errors.New("API key not found")
)

// WebAPIKey represents a Web API authentication key.
type WebAPIKey struct {
	DevID          string     `json:"dev_id"`
	DevKey         string     `json:"dev_key"`
	AppName        string     `json:"app_name"`
	CreatedAt      time.Time  `json:"created_at"`
	LastUsed       *time.Time `json:"last_used,omitempty"`
	IsActive       bool       `json:"is_active"`
	RateLimit      int        `json:"rate_limit"`
	AllowedOrigins []string   `json:"allowed_origins"`
	Capabilities   []string   `json:"capabilities"`
}

// WebAPIKeyUpdate represents fields that can be updated for an API key.
type WebAPIKeyUpdate struct {
	AppName        *string   `json:"app_name,omitempty"`
	IsActive       *bool     `json:"is_active,omitempty"`
	RateLimit      *int      `json:"rate_limit,omitempty"`
	AllowedOrigins *[]string `json:"allowed_origins,omitempty"`
	Capabilities   *[]string `json:"capabilities,omitempty"`
}

// CreateAPIKey inserts a new API key into the database.
func (f SQLiteUserStore) CreateAPIKey(ctx context.Context, key WebAPIKey) error {
	originsJSON, err := json.Marshal(key.AllowedOrigins)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed origins: %w", err)
	}

	capabilitiesJSON, err := json.Marshal(key.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	q := `
		INSERT INTO web_api_keys (dev_id, dev_key, app_name, created_at, is_active, rate_limit, allowed_origins, capabilities)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (dev_id) DO NOTHING
	`

	result, err := f.db.ExecContext(ctx,
		q,
		key.DevID,
		key.DevKey,
		key.AppName,
		key.CreatedAt.Unix(),
		key.IsActive,
		key.RateLimit,
		string(originsJSON),
		string(capabilitiesJSON),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrDupAPIKey
	}

	return nil
}

// GetAPIKeyByDevKey retrieves an API key by its dev_key value.
func (f *SQLiteUserStore) GetAPIKeyByDevKey(ctx context.Context, devKey string) (*WebAPIKey, error) {
	q := `
		SELECT dev_id, dev_key, app_name, created_at, last_used, is_active, rate_limit, allowed_origins, capabilities
		FROM web_api_keys
		WHERE dev_key = ? AND is_active = 1
	`

	var key WebAPIKey
	var createdAt, lastUsed sql.NullInt64
	var originsJSON, capabilitiesJSON string

	err := f.db.QueryRowContext(ctx, q, devKey).Scan(
		&key.DevID,
		&key.DevKey,
		&key.AppName,
		&createdAt,
		&lastUsed,
		&key.IsActive,
		&key.RateLimit,
		&originsJSON,
		&capabilitiesJSON,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNoAPIKey
	}
	if err != nil {
		return nil, err
	}

	key.CreatedAt = time.Unix(createdAt.Int64, 0)
	if lastUsed.Valid {
		t := time.Unix(lastUsed.Int64, 0)
		key.LastUsed = &t
	}

	if err := json.Unmarshal([]byte(originsJSON), &key.AllowedOrigins); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed origins: %w", err)
	}

	if err := json.Unmarshal([]byte(capabilitiesJSON), &key.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	return &key, nil
}

// GetAPIKeyByDevID retrieves an API key by its dev_id value.
func (f SQLiteUserStore) GetAPIKeyByDevID(ctx context.Context, devID string) (*WebAPIKey, error) {
	q := `
		SELECT dev_id, dev_key, app_name, created_at, last_used, is_active, rate_limit, allowed_origins, capabilities
		FROM web_api_keys
		WHERE dev_id = ?
	`

	var key WebAPIKey
	var createdAt, lastUsed sql.NullInt64
	var originsJSON, capabilitiesJSON string

	err := f.db.QueryRowContext(ctx, q, devID).Scan(
		&key.DevID,
		&key.DevKey,
		&key.AppName,
		&createdAt,
		&lastUsed,
		&key.IsActive,
		&key.RateLimit,
		&originsJSON,
		&capabilitiesJSON,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNoAPIKey
	}
	if err != nil {
		return nil, err
	}

	key.CreatedAt = time.Unix(createdAt.Int64, 0)
	if lastUsed.Valid {
		t := time.Unix(lastUsed.Int64, 0)
		key.LastUsed = &t
	}

	if err := json.Unmarshal([]byte(originsJSON), &key.AllowedOrigins); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed origins: %w", err)
	}

	if err := json.Unmarshal([]byte(capabilitiesJSON), &key.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	return &key, nil
}

// ListAPIKeys retrieves all API keys from the database.
func (f SQLiteUserStore) ListAPIKeys(ctx context.Context) ([]WebAPIKey, error) {
	q := `
		SELECT dev_id, dev_key, app_name, created_at, last_used, is_active, rate_limit, allowed_origins, capabilities
		FROM web_api_keys
		ORDER BY created_at DESC
	`

	rows, err := f.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []WebAPIKey
	for rows.Next() {
		var key WebAPIKey
		var createdAt, lastUsed sql.NullInt64
		var originsJSON, capabilitiesJSON string

		err := rows.Scan(
			&key.DevID,
			&key.DevKey,
			&key.AppName,
			&createdAt,
			&lastUsed,
			&key.IsActive,
			&key.RateLimit,
			&originsJSON,
			&capabilitiesJSON,
		)
		if err != nil {
			return nil, err
		}

		key.CreatedAt = time.Unix(createdAt.Int64, 0)
		if lastUsed.Valid {
			t := time.Unix(lastUsed.Int64, 0)
			key.LastUsed = &t
		}

		if err := json.Unmarshal([]byte(originsJSON), &key.AllowedOrigins); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed origins: %w", err)
		}

		if err := json.Unmarshal([]byte(capabilitiesJSON), &key.Capabilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}

		keys = append(keys, key)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

// UpdateAPIKey updates an existing API key's fields.
func (f SQLiteUserStore) UpdateAPIKey(ctx context.Context, devID string, updates WebAPIKeyUpdate) error {
	// Build dynamic UPDATE query based on provided fields
	var setClauses []string
	var args []interface{}

	if updates.AppName != nil {
		setClauses = append(setClauses, "app_name = ?")
		args = append(args, *updates.AppName)
	}

	if updates.IsActive != nil {
		setClauses = append(setClauses, "is_active = ?")
		args = append(args, *updates.IsActive)
	}

	if updates.RateLimit != nil {
		setClauses = append(setClauses, "rate_limit = ?")
		args = append(args, *updates.RateLimit)
	}

	if updates.AllowedOrigins != nil {
		originsJSON, err := json.Marshal(*updates.AllowedOrigins)
		if err != nil {
			return fmt.Errorf("failed to marshal allowed origins: %w", err)
		}
		setClauses = append(setClauses, "allowed_origins = ?")
		args = append(args, string(originsJSON))
	}

	if updates.Capabilities != nil {
		capabilitiesJSON, err := json.Marshal(*updates.Capabilities)
		if err != nil {
			return fmt.Errorf("failed to marshal capabilities: %w", err)
		}
		setClauses = append(setClauses, "capabilities = ?")
		args = append(args, string(capabilitiesJSON))
	}

	if len(setClauses) == 0 {
		return nil // No updates to perform
	}

	// Add WHERE clause argument
	args = append(args, devID)

	q := fmt.Sprintf(`
		UPDATE web_api_keys
		SET %s
		WHERE dev_id = ?
	`, joinStrings(setClauses, ", "))

	result, err := f.db.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoAPIKey
	}

	return nil
}

// DeleteAPIKey removes an API key from the database.
func (f SQLiteUserStore) DeleteAPIKey(ctx context.Context, devID string) error {
	q := `
		DELETE FROM web_api_keys WHERE dev_id = ?
	`
	result, err := f.db.ExecContext(ctx, q, devID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoAPIKey
	}

	return nil
}

// UpdateLastUsed updates the last_used timestamp for an API key.
func (f *SQLiteUserStore) UpdateLastUsed(ctx context.Context, devKey string) error {
	q := `
		UPDATE web_api_keys
		SET last_used = ?
		WHERE dev_key = ?
	`

	_, err := f.db.ExecContext(ctx, q, time.Now().Unix(), devKey)
	return err
}

// joinStrings is a helper function to join strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
