package state

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	// ErrNoWebAPISession is returned when a WebAPI session is not found.
	ErrNoWebAPISession = errors.New("WebAPI session not found")
	// ErrWebAPISessionExpired is returned when a WebAPI session has expired.
	ErrWebAPISessionExpired = errors.New("WebAPI session expired")
)

// WebAPISession represents an active Web AIM API session.
type WebAPISession struct {
	AimSID          string            // Unique session ID for web client
	ScreenName      DisplayScreenName // User identity
	OSCARSession    *Session          // Bridge to existing OSCAR session
	Events          []string          // Subscribed event types
	EventQueue      *EventQueue       // Per-session event queue
	DevID           string            // Developer ID that created this session
	ClientName      string            // Client application name
	ClientVersion   string            // Client application version
	CreatedAt       time.Time         // Session creation time
	LastAccessed    time.Time         // Last activity time
	ExpiresAt       time.Time         // Session expiration time
	FetchTimeout    int               // Long-polling timeout in milliseconds
	TimeToNextFetch int               // Suggested delay before next fetch
	RemoteAddr      string            // Client IP address
}

// IsExpired checks if the session has expired.
func (s *WebAPISession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Touch updates the last accessed time and extends expiration if needed.
func (s *WebAPISession) Touch() {
	s.LastAccessed = time.Now()
	// Extend expiration by 60 minutes from last access
	newExpiry := s.LastAccessed.Add(60 * time.Minute)
	if newExpiry.After(s.ExpiresAt) {
		s.ExpiresAt = newExpiry
	}
}

// IsSubscribedTo checks if the session is subscribed to a specific event type.
func (s *WebAPISession) IsSubscribedTo(eventType string) bool {
	for _, event := range s.Events {
		if event == eventType {
			return true
		}
	}
	return false
}

// WebAPISessionManager manages Web API sessions with thread-safe operations.
type WebAPISessionManager struct {
	sessions      map[string]*WebAPISession          // Keyed by aimsid
	byUser        map[IdentScreenName]*WebAPISession // Keyed by screen name
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// NewWebAPISessionManager creates a new WebAPI session manager.
func NewWebAPISessionManager() *WebAPISessionManager {
	mgr := &WebAPISessionManager{
		sessions:    make(map[string]*WebAPISession),
		byUser:      make(map[IdentScreenName]*WebAPISession),
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine to remove expired sessions
	mgr.cleanupTicker = time.NewTicker(1 * time.Minute)
	go mgr.cleanupExpiredSessions()

	return mgr
}

// CreateSession creates a new WebAPI session.
func (m *WebAPISessionManager) CreateSession(screenName DisplayScreenName, devID string, events []string) (*WebAPISession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if user already has an active session
	identName := screenName.IdentScreenName()
	if existing, exists := m.byUser[identName]; exists {
		// Remove the old session
		delete(m.sessions, existing.AimSID)
	}

	// Generate unique session ID
	aimsid, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &WebAPISession{
		AimSID:          aimsid,
		ScreenName:      screenName,
		Events:          events,
		EventQueue:      NewEventQueue(1000), // Max 1000 events per session
		DevID:           devID,
		CreatedAt:       now,
		LastAccessed:    now,
		ExpiresAt:       now.Add(60 * time.Minute), // 60 minute initial expiry
		FetchTimeout:    60000,                     // 60 seconds default for better stability
		TimeToNextFetch: 500,                       // 500ms suggested delay
	}

	m.sessions[aimsid] = session
	m.byUser[identName] = session

	return session, nil
}

// GetSession retrieves a session by aimsid.
func (m *WebAPISessionManager) GetSession(aimsid string) (*WebAPISession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[aimsid]
	if !exists {
		return nil, ErrNoWebAPISession
	}

	if session.IsExpired() {
		return nil, ErrWebAPISessionExpired
	}

	return session, nil
}

// GetSessionByUser retrieves a session by screen name.
func (m *WebAPISessionManager) GetSessionByUser(screenName IdentScreenName) (*WebAPISession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.byUser[screenName]
	if !exists {
		return nil, ErrNoWebAPISession
	}

	if session.IsExpired() {
		return nil, ErrWebAPISessionExpired
	}

	return session, nil
}

// RemoveSession removes a session by aimsid.
func (m *WebAPISessionManager) RemoveSession(aimsid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[aimsid]
	if !exists {
		return ErrNoWebAPISession
	}

	delete(m.sessions, aimsid)
	delete(m.byUser, session.ScreenName.IdentScreenName())

	// Close the event queue to unblock any waiting fetches
	if session.EventQueue != nil {
		session.EventQueue.Close()
	}

	return nil
}

// TouchSession updates the last accessed time for a session.
func (m *WebAPISessionManager) TouchSession(aimsid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[aimsid]
	if !exists {
		return ErrNoWebAPISession
	}

	session.Touch()
	return nil
}

// GetAllSessions returns all active sessions (for monitoring/admin).
func (m *WebAPISessionManager) GetAllSessions() []*WebAPISession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*WebAPISession, 0, len(m.sessions))
	for _, session := range m.sessions {
		if !session.IsExpired() {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// GetSessionsByScreenName returns all sessions for a given screen name.
func (m *WebAPISessionManager) GetSessionsByScreenName(screenName DisplayScreenName) []*WebAPISession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []*WebAPISession
	identScreenName := screenName.IdentScreenName()

	// Check both the byUser map and iterate through all sessions
	// since a user might have multiple sessions
	for _, session := range m.sessions {
		if session.ScreenName.IdentScreenName() == identScreenName {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// cleanupExpiredSessions periodically removes expired sessions.
func (m *WebAPISessionManager) cleanupExpiredSessions() {
	for {
		select {
		case <-m.cleanupTicker.C:
			m.mu.Lock()
			now := time.Now()
			var toRemove []string

			for aimsid, session := range m.sessions {
				if now.After(session.ExpiresAt) {
					toRemove = append(toRemove, aimsid)
				}
			}

			for _, aimsid := range toRemove {
				session := m.sessions[aimsid]
				delete(m.sessions, aimsid)
				delete(m.byUser, session.ScreenName.IdentScreenName())
				if session.EventQueue != nil {
					session.EventQueue.Close()
				}
			}
			m.mu.Unlock()

		case <-m.stopCleanup:
			m.cleanupTicker.Stop()
			return
		}
	}
}

// Shutdown stops the session manager and cleans up resources.
func (m *WebAPISessionManager) Shutdown() {
	close(m.stopCleanup)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all event queues
	for _, session := range m.sessions {
		if session.EventQueue != nil {
			session.EventQueue.Close()
		}
	}

	// Clear all sessions
	m.sessions = make(map[string]*WebAPISession)
	m.byUser = make(map[IdentScreenName]*WebAPISession)
}

// generateSessionID creates a cryptographically secure session ID.
func generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
