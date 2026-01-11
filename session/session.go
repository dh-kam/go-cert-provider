package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// UserSession represents a user session with authentication info
type UserSession struct {
	SessionID      string    `json:"session_id"`
	UserID         string    `json:"user_id"`
	Description    string    `json:"description"`
	ExpireDate     time.Time `json:"expire_date"`
	AllowedDomains []string  `json:"allowed_domains"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

// SessionManager manages user sessions in memory
type SessionManager struct {
	sessions map[string]*UserSession
	mutex    sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*UserSession),
	}
	
	// Start cleanup routine for expired sessions
	go sm.cleanupExpiredSessions()
	
	return sm
}

// CreateSession creates a new session and returns session ID
func (sm *SessionManager) CreateSession(userID, description string, expireDate time.Time, allowedDomains []string) string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sessionID := uuid.New().String()
	now := time.Now()

	// Session expires in 30 minutes or at JWT expiry, whichever comes first
	sessionExpiry := now.Add(30 * time.Minute)
	if expireDate.Before(sessionExpiry) {
		sessionExpiry = expireDate
	}

	session := &UserSession{
		SessionID:      sessionID,
		UserID:         userID,
		Description:    description,
		ExpireDate:     sessionExpiry,
		AllowedDomains: allowedDomains,
		CreatedAt:      now,
		LastAccessedAt: now,
	}

	sm.sessions[sessionID] = session
	return sessionID
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*UserSession, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpireDate) {
		go sm.DeleteSession(sessionID)
		return nil, false
	}

	session.LastAccessedAt = time.Now()
	return session, true
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	delete(sm.sessions, sessionID)
}

// CleanupExpiredSessions manually triggers cleanup of expired sessions (for testing)
func (sm *SessionManager) CleanupExpiredSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	for sessionID, session := range sm.sessions {
		if now.After(session.ExpireDate) {
			delete(sm.sessions, sessionID)
		}
	}
}
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mutex.Lock()
		now := time.Now()
		for sessionID, session := range sm.sessions {
			if now.After(session.ExpireDate) {
				delete(sm.sessions, sessionID)
			}
		}
		sm.mutex.Unlock()
	}
}

// Global session manager instance
var globalSessionManager *SessionManager

// GetGlobalSessionManager returns the global session manager instance
func GetGlobalSessionManager() *SessionManager {
	if globalSessionManager == nil {
		globalSessionManager = NewSessionManager()
	}
	return globalSessionManager
}
