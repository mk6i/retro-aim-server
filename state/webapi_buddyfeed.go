package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// BuddyFeed represents a user's feed configuration.
type BuddyFeed struct {
	ID          int64     `json:"id"`
	ScreenName  string    `json:"screenName"`
	FeedType    string    `json:"feedType"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	PublishedAt time.Time `json:"publishedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	IsActive    bool      `json:"isActive"`
}

// BuddyFeedItem represents an individual feed entry.
type BuddyFeedItem struct {
	ID          int64     `json:"id"`
	FeedID      int64     `json:"feedId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	GUID        string    `json:"guid"`
	Author      string    `json:"author"`
	Categories  []string  `json:"categories"`
	PublishedAt time.Time `json:"publishedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

// BuddyFeedSubscription represents a feed subscription.
type BuddyFeedSubscription struct {
	ID                   int64      `json:"id"`
	SubscriberScreenName string     `json:"subscriberScreenName"`
	FeedID               int64      `json:"feedId"`
	SubscribedAt         time.Time  `json:"subscribedAt"`
	LastCheckedAt        *time.Time `json:"lastCheckedAt"`
}

// RSS/Atom feed structures for XML output
type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language,omitempty"`
	PubDate     string    `xml:"pubDate,omitempty"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Author      string   `xml:"author,omitempty"`
	Categories  []string `xml:"category,omitempty"`
	GUID        string   `xml:"guid,omitempty"`
	PubDate     string   `xml:"pubDate"`
}

type AtomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"title"`
	Link    AtomLink    `xml:"link"`
	Updated string      `xml:"updated"`
	Author  AtomAuthor  `xml:"author,omitempty"`
	ID      string      `xml:"id"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

type AtomEntry struct {
	Title     string      `xml:"title"`
	Link      AtomLink    `xml:"link"`
	ID        string      `xml:"id"`
	Updated   string      `xml:"updated"`
	Published string      `xml:"published,omitempty"`
	Author    AtomAuthor  `xml:"author,omitempty"`
	Summary   string      `xml:"summary,omitempty"`
	Content   AtomContent `xml:"content,omitempty"`
}

type AtomContent struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

// BuddyFeedManager manages buddy feed operations.
type BuddyFeedManager struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewBuddyFeedManager creates a new buddy feed manager.
func NewBuddyFeedManager(db *sql.DB, logger *slog.Logger) *BuddyFeedManager {
	return &BuddyFeedManager{
		db:     db,
		logger: logger,
	}
}

// CreateFeed creates a new buddy feed.
func (m *BuddyFeedManager) CreateFeed(ctx context.Context, feed BuddyFeed) (*BuddyFeed, error) {
	now := time.Now()
	query := `
		INSERT INTO buddy_feeds (
			screen_name, feed_type, title, description, link,
			published_at, created_at, updated_at, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	var id int64
	err := m.db.QueryRowContext(ctx, query,
		feed.ScreenName, feed.FeedType, feed.Title, feed.Description, feed.Link,
		feed.PublishedAt.Unix(), now.Unix(), now.Unix(), feed.IsActive,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to create feed: %w", err)
	}

	feed.ID = id
	feed.CreatedAt = now
	feed.UpdatedAt = now

	return &feed, nil
}

// GetUserFeed retrieves the feed for a specific user.
func (m *BuddyFeedManager) GetUserFeed(ctx context.Context, screenName string, format string) (interface{}, error) {
	// Get feed configuration
	var feed BuddyFeed
	query := `
		SELECT id, screen_name, feed_type, title, description, link,
		       published_at, created_at, updated_at, is_active
		FROM buddy_feeds
		WHERE screen_name = ? AND is_active = 1
		ORDER BY published_at DESC
		LIMIT 1
	`

	err := m.db.QueryRowContext(ctx, query, screenName).Scan(
		&feed.ID, &feed.ScreenName, &feed.FeedType, &feed.Title,
		&feed.Description, &feed.Link, &feed.PublishedAt,
		&feed.CreatedAt, &feed.UpdatedAt, &feed.IsActive,
	)

	if err == sql.ErrNoRows {
		// No feed configured for this user
		return m.generateEmptyFeed(screenName, format), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user feed: %w", err)
	}

	// Get feed items
	items, err := m.GetFeedItems(ctx, feed.ID, 50) // Limit to 50 most recent items
	if err != nil {
		return nil, fmt.Errorf("failed to get feed items: %w", err)
	}

	// Generate feed in requested format
	return m.generateFeed(feed, items, format)
}

// GetBuddyListFeed retrieves aggregated feed for a user's buddy list.
func (m *BuddyFeedManager) GetBuddyListFeed(ctx context.Context, buddies []IdentScreenName, format string, limit int) (interface{}, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}

	// Get all feed items from buddies
	var allItems []BuddyFeedItem
	for _, buddy := range buddies {
		items, err := m.GetUserFeedItems(ctx, buddy.String(), 20) // Get up to 20 items per buddy
		if err != nil {
			m.logger.WarnContext(ctx, "failed to get feed items for buddy",
				"buddy", buddy.String(), "error", err)
			continue
		}
		allItems = append(allItems, items...)
	}

	// Sort by published date (newest first)
	// Note: In production, this should be done in the database query
	// This is simplified for the example

	// Create aggregate feed
	feed := BuddyFeed{
		Title:       "Buddy List Feed",
		Description: "Aggregated feed from your buddy list",
		Link:        "/buddyfeed/getBuddylist",
		PublishedAt: time.Now(),
	}

	// Limit items
	if len(allItems) > limit {
		allItems = allItems[:limit]
	}

	return m.generateFeed(feed, allItems, format)
}

// GetUserFeedItems retrieves feed items for a specific user.
func (m *BuddyFeedManager) GetUserFeedItems(ctx context.Context, screenName string, limit int) ([]BuddyFeedItem, error) {
	query := `
		SELECT i.id, i.feed_id, i.title, i.description, i.link, i.guid,
		       i.author, i.categories, i.published_at, i.created_at
		FROM buddy_feed_items i
		JOIN buddy_feeds f ON i.feed_id = f.id
		WHERE f.screen_name = ? AND f.is_active = 1
		ORDER BY i.published_at DESC
		LIMIT ?
	`

	rows, err := m.db.QueryContext(ctx, query, screenName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query feed items: %w", err)
	}
	defer rows.Close()

	var items []BuddyFeedItem
	for rows.Next() {
		var item BuddyFeedItem
		var categoriesJSON sql.NullString
		var publishedAt, createdAt int64

		err := rows.Scan(
			&item.ID, &item.FeedID, &item.Title, &item.Description,
			&item.Link, &item.GUID, &item.Author, &categoriesJSON,
			&publishedAt, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed item: %w", err)
		}

		item.PublishedAt = time.Unix(publishedAt, 0)
		item.CreatedAt = time.Unix(createdAt, 0)

		if categoriesJSON.Valid {
			json.Unmarshal([]byte(categoriesJSON.String), &item.Categories)
		}

		items = append(items, item)
	}

	return items, nil
}

// GetFeedItems retrieves items for a specific feed.
func (m *BuddyFeedManager) GetFeedItems(ctx context.Context, feedID int64, limit int) ([]BuddyFeedItem, error) {
	query := `
		SELECT id, feed_id, title, description, link, guid,
		       author, categories, published_at, created_at
		FROM buddy_feed_items
		WHERE feed_id = ?
		ORDER BY published_at DESC
		LIMIT ?
	`

	rows, err := m.db.QueryContext(ctx, query, feedID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query feed items: %w", err)
	}
	defer rows.Close()

	var items []BuddyFeedItem
	for rows.Next() {
		var item BuddyFeedItem
		var categoriesJSON sql.NullString
		var publishedAt, createdAt int64

		err := rows.Scan(
			&item.ID, &item.FeedID, &item.Title, &item.Description,
			&item.Link, &item.GUID, &item.Author, &categoriesJSON,
			&publishedAt, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed item: %w", err)
		}

		item.PublishedAt = time.Unix(publishedAt, 0)
		item.CreatedAt = time.Unix(createdAt, 0)

		if categoriesJSON.Valid {
			json.Unmarshal([]byte(categoriesJSON.String), &item.Categories)
		}

		items = append(items, item)
	}

	return items, nil
}

// AddFeedItem adds a new item to a feed.
func (m *BuddyFeedManager) AddFeedItem(ctx context.Context, feedID int64, item BuddyFeedItem) (*BuddyFeedItem, error) {
	categoriesJSON, _ := json.Marshal(item.Categories)
	now := time.Now()

	query := `
		INSERT INTO buddy_feed_items (
			feed_id, title, description, link, guid,
			author, categories, published_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	var id int64
	err := m.db.QueryRowContext(ctx, query,
		feedID, item.Title, item.Description, item.Link, item.GUID,
		item.Author, string(categoriesJSON), item.PublishedAt.Unix(), now.Unix(),
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to add feed item: %w", err)
	}

	item.ID = id
	item.FeedID = feedID
	item.CreatedAt = now

	// Update feed's updated_at timestamp
	updateQuery := `UPDATE buddy_feeds SET updated_at = ? WHERE id = ?`
	m.db.ExecContext(ctx, updateQuery, now.Unix(), feedID)

	return &item, nil
}

// PushFeed handles feed submission from a user.
func (m *BuddyFeedManager) PushFeed(ctx context.Context, screenName string, feedData map[string]interface{}) error {
	// Extract feed data
	title, _ := feedData["title"].(string)
	description, _ := feedData["description"].(string)
	link, _ := feedData["link"].(string)
	feedType, _ := feedData["type"].(string)

	if feedType == "" {
		feedType = "status" // Default to status update
	}

	// Get or create feed for user
	var feedID int64
	query := `SELECT id FROM buddy_feeds WHERE screen_name = ? AND is_active = 1 LIMIT 1`
	err := m.db.QueryRowContext(ctx, query, screenName).Scan(&feedID)

	if err == sql.ErrNoRows {
		// Create new feed
		feed := BuddyFeed{
			ScreenName:  screenName,
			FeedType:    feedType,
			Title:       fmt.Sprintf("%s's Feed", screenName),
			Description: fmt.Sprintf("Updates from %s", screenName),
			Link:        fmt.Sprintf("/buddyfeed/getUser?u=%s", screenName),
			PublishedAt: time.Now(),
			IsActive:    true,
		}

		createdFeed, err := m.CreateFeed(ctx, feed)
		if err != nil {
			return fmt.Errorf("failed to create feed: %w", err)
		}
		feedID = createdFeed.ID
	} else if err != nil {
		return fmt.Errorf("failed to get feed: %w", err)
	}

	// Add feed item
	item := BuddyFeedItem{
		Title:       title,
		Description: description,
		Link:        link,
		Author:      screenName,
		PublishedAt: time.Now(),
		GUID:        fmt.Sprintf("%s-%d", screenName, time.Now().UnixNano()),
	}

	// Extract categories if provided
	if cats, ok := feedData["categories"].([]interface{}); ok {
		for _, cat := range cats {
			if catStr, ok := cat.(string); ok {
				item.Categories = append(item.Categories, catStr)
			}
		}
	}

	_, err = m.AddFeedItem(ctx, feedID, item)
	return err
}

// generateFeed creates a feed in the requested format.
func (m *BuddyFeedManager) generateFeed(feed BuddyFeed, items []BuddyFeedItem, format string) (interface{}, error) {
	format = strings.ToLower(format)

	switch format {
	case "atom":
		return m.generateAtomFeed(feed, items), nil
	case "json":
		return m.generateJSONFeed(feed, items), nil
	case "rss", "":
		return m.generateRSSFeed(feed, items), nil
	default:
		return m.generateRSSFeed(feed, items), nil
	}
}

// generateRSSFeed creates an RSS 2.0 feed.
func (m *BuddyFeedManager) generateRSSFeed(feed BuddyFeed, items []BuddyFeedItem) *RSSFeed {
	rss := &RSSFeed{
		Version: "2.0",
		Channel: RSSChannel{
			Title:       feed.Title,
			Link:        feed.Link,
			Description: feed.Description,
			Language:    "en-US",
			PubDate:     feed.PublishedAt.Format(time.RFC1123Z),
			Items:       make([]RSSItem, 0, len(items)),
		},
	}

	for _, item := range items {
		rssItem := RSSItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			Author:      item.Author,
			Categories:  item.Categories,
			GUID:        item.GUID,
			PubDate:     item.PublishedAt.Format(time.RFC1123Z),
		}
		rss.Channel.Items = append(rss.Channel.Items, rssItem)
	}

	return rss
}

// generateAtomFeed creates an Atom 1.0 feed.
func (m *BuddyFeedManager) generateAtomFeed(feed BuddyFeed, items []BuddyFeedItem) *AtomFeed {
	atom := &AtomFeed{
		Title:   feed.Title,
		Link:    AtomLink{Href: feed.Link, Rel: "alternate"},
		Updated: feed.UpdatedAt.Format(time.RFC3339),
		ID:      feed.Link,
		Author:  AtomAuthor{Name: feed.ScreenName},
		Entries: make([]AtomEntry, 0, len(items)),
	}

	for _, item := range items {
		entry := AtomEntry{
			Title:     item.Title,
			Link:      AtomLink{Href: item.Link},
			ID:        item.GUID,
			Updated:   item.PublishedAt.Format(time.RFC3339),
			Published: item.PublishedAt.Format(time.RFC3339),
			Author:    AtomAuthor{Name: item.Author},
			Summary:   item.Description,
			Content:   AtomContent{Type: "html", Text: item.Description},
		}
		atom.Entries = append(atom.Entries, entry)
	}

	return atom
}

// generateJSONFeed creates a JSON feed.
func (m *BuddyFeedManager) generateJSONFeed(feed BuddyFeed, items []BuddyFeedItem) map[string]interface{} {
	jsonItems := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		jsonItem := map[string]interface{}{
			"id":          item.GUID,
			"title":       item.Title,
			"description": item.Description,
			"link":        item.Link,
			"author":      item.Author,
			"categories":  item.Categories,
			"published":   item.PublishedAt.Unix(),
		}
		jsonItems = append(jsonItems, jsonItem)
	}

	return map[string]interface{}{
		"title":       feed.Title,
		"description": feed.Description,
		"link":        feed.Link,
		"updated":     feed.UpdatedAt.Unix(),
		"items":       jsonItems,
	}
}

// generateEmptyFeed creates an empty feed for users without configured feeds.
func (m *BuddyFeedManager) generateEmptyFeed(screenName string, format string) interface{} {
	feed := BuddyFeed{
		ScreenName:  screenName,
		Title:       fmt.Sprintf("%s's Feed", screenName),
		Description: fmt.Sprintf("No updates from %s", screenName),
		Link:        fmt.Sprintf("/buddyfeed/getUser?u=%s", screenName),
		PublishedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, _ := m.generateFeed(feed, []BuddyFeedItem{}, format)
	return result
}
