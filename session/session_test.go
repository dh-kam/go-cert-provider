package session

import (
	"testing"
	"time"
)

func TestSessionManager_CreateAndGet(t *testing.T) {
	manager := NewSessionManager()

	userID := "test-user"
	description := "Test User"
	expiresAt := time.Now().Add(1 * time.Hour)
	allowedDomains := []string{"example.com", "test.com"}

	sessionID := manager.CreateSession(userID, description, expiresAt, allowedDomains)

	if sessionID == "" {
		t.Fatal("Session ID should not be empty")
	}

	session, exists := manager.GetSession(sessionID)
	if !exists {
		t.Fatal("Session should exist")
	}

	// Verify session data
	if session.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, session.UserID)
	}

	if session.Description != description {
		t.Errorf("Expected description %s, got %s", description, session.Description)
	}

	if len(session.AllowedDomains) != len(allowedDomains) {
		t.Errorf("Expected %d domains, got %d", len(allowedDomains), len(session.AllowedDomains))
	}

	for i, domain := range allowedDomains {
		if session.AllowedDomains[i] != domain {
			t.Errorf("Expected domain %s at index %d, got %s", domain, i, session.AllowedDomains[i])
		}
	}

	expectedExpiry := time.Now().Add(30 * time.Minute)
	if expiresAt.Before(expectedExpiry) {
		expectedExpiry = expiresAt
	}
	timeDiff := session.ExpireDate.Sub(expectedExpiry).Abs()
	if timeDiff > time.Second {
		t.Errorf("ExpireDate time difference too large: %v", timeDiff)
	}
}

func TestSessionManager_DeleteSession(t *testing.T) {
	manager := NewSessionManager()

	sessionID := manager.CreateSession("user1", "User One", time.Now().Add(1*time.Hour), []string{"example.com"})
	_, exists := manager.GetSession(sessionID)
	if !exists {
		t.Fatal("Session should exist before deletion")
	}

	// Delete session
	manager.DeleteSession(sessionID)

	_, exists = manager.GetSession(sessionID)
	if exists {
		t.Error("Session should not exist after deletion")
	}
}

func TestSessionManager_ExpiredSession(t *testing.T) {
	manager := NewSessionManager()

	sessionID := manager.CreateSession(
		"expired-user",
		"Expired User",
		time.Now().Add(-1*time.Hour),
		[]string{"example.com"},
	)
	_, exists := manager.GetSession(sessionID)
	if exists {
		t.Error("Expired session should not be returned")
	}
}

func TestSessionManager_CleanupExpiredSessions(t *testing.T) {
	manager := NewSessionManager()

	validID := manager.CreateSession("valid-user", "Valid", time.Now().Add(1*time.Hour), []string{"example.com"})
	expiredID := manager.CreateSession("expired-user", "Expired", time.Now().Add(-1*time.Hour), []string{"test.com"})

	manager.CleanupExpiredSessions()
	_, validExists := manager.GetSession(validID)
	if !validExists {
		t.Error("Valid session should exist after cleanup")
	}

	manager.mutex.RLock()
	_, expiredStillExists := manager.sessions[expiredID]
	manager.mutex.RUnlock()

	if expiredStillExists {
		t.Error("Expired session should be removed after cleanup")
	}
}

func TestSessionManager_MultipleSessions(t *testing.T) {
	manager := NewSessionManager()

	sessions := make(map[string]string)
	expiresAt := time.Now().Add(1 * time.Hour)

	for i := 0; i < 10; i++ {
		userID := string(rune('a' + i))
		sessionID := manager.CreateSession(userID, "User "+userID, expiresAt, []string{"example.com"})
		sessions[sessionID] = userID
	}

	for sessionID, expectedUserID := range sessions {
		session, exists := manager.GetSession(sessionID)
		if !exists {
			t.Errorf("Session %s should exist", sessionID)
			continue
		}

		if session.UserID != expectedUserID {
			t.Errorf("Expected userID %s, got %s", expectedUserID, session.UserID)
		}
	}
}

func TestSessionManager_GetNonExistentSession(t *testing.T) {
	manager := NewSessionManager()

	_, exists := manager.GetSession("non-existent-session-id")
	if exists {
		t.Error("Non-existent session should not exist")
	}
}

func TestSessionManager_DeleteNonExistentSession(t *testing.T) {
	manager := NewSessionManager()

	// Should not panic when deleting non-existent session
	manager.DeleteSession("non-existent-session-id")
}

func TestGlobalSessionManager(t *testing.T) {
	manager1 := GetGlobalSessionManager()
	if manager1 == nil {
		t.Fatal("Global session manager should not be nil")
	}

	manager2 := GetGlobalSessionManager()
	if manager1 != manager2 {
		t.Error("Global session manager should return the same instance")
	}

	sessionID := manager1.CreateSession("test", "Test", time.Now().Add(1*time.Hour), []string{"example.com"})
	_, exists := manager2.GetSession(sessionID)
	if !exists {
		t.Error("Session should exist in global manager")
	}

	manager1.DeleteSession(sessionID)
}

func TestSessionManager_UniqueSessionIDs(t *testing.T) {
	manager := NewSessionManager()

	expiresAt := time.Now().Add(1 * time.Hour)
	sessionIDs := make(map[string]bool)

	for i := 0; i < 100; i++ {
		sessionID := manager.CreateSession("same-user", "Same User", expiresAt, []string{"example.com"})
		
		if sessionIDs[sessionID] {
			t.Fatalf("Duplicate session ID detected: %s", sessionID)
		}
		sessionIDs[sessionID] = true
	}
}

func TestSessionManager_ConcurrentAccess(t *testing.T) {
	manager := NewSessionManager()
	expiresAt := time.Now().Add(1 * time.Hour)

	done := make(chan bool)

	go func() {
		for i := 0; i < 50; i++ {
			manager.CreateSession("user-goroutine1", "User 1", expiresAt, []string{"example.com"})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			manager.CreateSession("user-goroutine2", "User 2", expiresAt, []string{"test.com"})
		}
		done <- true
	}()

	<-done
	<-done

	manager.CleanupExpiredSessions()
}

func TestSessionManager_EmptyFields(t *testing.T) {
	manager := NewSessionManager()
	expiresAt := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name           string
		userID         string
		description    string
		allowedDomains []string
	}{
		{"empty userID", "", "Description", []string{"example.com"}},
		{"empty description", "user", "", []string{"example.com"}},
		{"empty domains", "user", "Description", []string{}},
		{"nil domains", "user", "Description", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID := manager.CreateSession(tt.userID, tt.description, expiresAt, tt.allowedDomains)
			
			if sessionID == "" {
				t.Error("Session ID should not be empty even with empty fields")
			}

			session, exists := manager.GetSession(sessionID)
			if !exists {
				t.Error("Session should exist even with empty fields")
			}

			if session.UserID != tt.userID {
				t.Errorf("UserID mismatch: expected %q, got %q", tt.userID, session.UserID)
			}

			manager.DeleteSession(sessionID)
		})
	}
}
