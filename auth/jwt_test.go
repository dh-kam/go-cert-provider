package auth

import (
	"testing"
	"time"
)

func TestCreateJWT(t *testing.T) {
	secretKey := "test-secret-key-32-bytes-long!!"
	userID := "test-user"
	description := "Test User"
	expiresAt := time.Now().Add(24 * time.Hour)
	allowedDomains := []string{"example.com", "test.com"}

	token, err := CreateJWT(userID, description, expiresAt, allowedDomains, secretKey)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Token should have 3 parts (header.payload.signature)
	parts := 0
	for _, c := range token {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("JWT should have 3 parts separated by 2 dots, got %d dots", parts)
	}
}

func TestParseJWT(t *testing.T) {
	secretKey := "test-secret-key-32-bytes-long!!"
	userID := "test-user"
	description := "Test User"
	expiresAt := time.Now().Add(24 * time.Hour)
	allowedDomains := []string{"example.com", "test.com"}

	token, err := CreateJWT(userID, description, expiresAt, allowedDomains, secretKey)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	claims, err := ParseJWT(token, secretKey)
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v", err)
	}

	// Verify claims
	if claims.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, claims.UserID)
	}

	if claims.Description != description {
		t.Errorf("Expected description %s, got %s", description, claims.Description)
	}

	if len(claims.AllowedDomains) != len(allowedDomains) {
		t.Errorf("Expected %d domains, got %d", len(allowedDomains), len(claims.AllowedDomains))
	}

	for i, domain := range allowedDomains {
		if claims.AllowedDomains[i] != domain {
			t.Errorf("Expected domain %s at index %d, got %s", domain, i, claims.AllowedDomains[i])
		}
	}
}

func TestParseJWT_ExpiredToken(t *testing.T) {
	secretKey := "test-secret-key-32-bytes-long!!"
	userID := "test-user"
	description := "Test User"
	expiresAt := time.Now().Add(-1 * time.Hour)
	allowedDomains := []string{"example.com"}

	token, err := CreateJWT(userID, description, expiresAt, allowedDomains, secretKey)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	_, err = ParseJWT(token, secretKey)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestParseJWT_InvalidToken(t *testing.T) {
	secretKey := "test-secret-key-32-bytes-long!!"

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "invalid.token"},
		{"random string", "this-is-not-a-jwt-token"},
		{"wrong parts", "one.two.three.four"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseJWT(tt.token, secretKey)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestParseJWT_WrongSecretKey(t *testing.T) {
	correctKey := "correct-secret-key-32-bytes!!"
	wrongKey := "wrong-secret-key-32-bytes-long"
	userID := "test-user"
	description := "Test User"
	expiresAt := time.Now().Add(24 * time.Hour)
	allowedDomains := []string{"example.com"}

	token, err := CreateJWT(userID, description, expiresAt, allowedDomains, correctKey)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	_, err = ParseJWT(token, wrongKey)
	if err == nil {
		t.Error("Expected error for wrong secret key, got nil")
	}
}

func TestCreateJWT_EmptyFields(t *testing.T) {
	secretKey := "test-secret-key-32-bytes-long!!"
	expiresAt := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name           string
		userID         string
		description    string
		allowedDomains []string
		shouldFail     bool
	}{
		{"empty userID", "", "Description", []string{"example.com"}, false},
		{"empty description", "user", "", []string{"example.com"}, false},
		{"empty domains", "user", "Description", []string{}, false},
		{"nil domains", "user", "Description", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := CreateJWT(tt.userID, tt.description, expiresAt, tt.allowedDomains, secretKey)
			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v", tt.name, err)
				}
				if token == "" {
					t.Errorf("Expected token for %s, got empty string", tt.name)
				}
			}
		})
	}
}

func TestJWT_RoundTrip(t *testing.T) {
	secretKey := "test-secret-key-32-bytes-long!!"

	testCases := []struct {
		name           string
		userID         string
		description    string
		expiresIn      time.Duration
		allowedDomains []string
	}{
		{
			name:           "single domain",
			userID:         "user1",
			description:    "Single Domain User",
			expiresIn:      1 * time.Hour,
			allowedDomains: []string{"example.com"},
		},
		{
			name:           "multiple domains",
			userID:         "user2",
			description:    "Multi Domain User",
			expiresIn:      24 * time.Hour,
			allowedDomains: []string{"example.com", "test.com", "*.example.com"},
		},
		{
			name:           "wildcard domain",
			userID:         "user3",
			description:    "Wildcard User",
			expiresIn:      7 * 24 * time.Hour,
			allowedDomains: []string{"*.example.com", "*.test.com"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expiresAt := time.Now().Add(tc.expiresIn)

			token, err := CreateJWT(tc.userID, tc.description, expiresAt, tc.allowedDomains, secretKey)
			if err != nil {
				t.Fatalf("Failed to generate JWT: %v", err)
			}

			claims, err := ParseJWT(token, secretKey)
			if err != nil {
				t.Fatalf("Failed to parse JWT: %v", err)
			}

			if claims.UserID != tc.userID {
				t.Errorf("UserID mismatch: expected %s, got %s", tc.userID, claims.UserID)
			}

			if claims.Description != tc.description {
				t.Errorf("Description mismatch: expected %s, got %s", tc.description, claims.Description)
			}

			if len(claims.AllowedDomains) != len(tc.allowedDomains) {
				t.Fatalf("Domain count mismatch: expected %d, got %d", len(tc.allowedDomains), len(claims.AllowedDomains))
			}

			for i, expected := range tc.allowedDomains {
				if claims.AllowedDomains[i] != expected {
					t.Errorf("Domain mismatch at index %d: expected %s, got %s", i, expected, claims.AllowedDomains[i])
				}
			}

			timeDiff := claims.ExpiresAt.Time.Sub(expiresAt).Abs()
			if timeDiff > time.Second {
				t.Errorf("ExpiresAt time difference too large: %v", timeDiff)
			}
		})
	}
}
