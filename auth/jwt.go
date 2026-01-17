package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	UserID         string   `json:"user_id"`
	Description    string   `json:"description"`
	AllowedDomains []string `json:"allowed_domains"`
	jwt.RegisteredClaims
}

// ParseJWT parses and validates a JWT token with secret verification
func ParseJWT(tokenString, secret string) (*JWTClaims, error) {
	if secret != "" {
		// Use secret verification if provided
		return ValidateJWTWithSecret(tokenString, secret)
	}

	// Parse without verification (for testing only)
	return parseJWTWithoutVerification(tokenString)
}

// parseJWTWithoutVerification parses JWT without signature verification (for testing)
func parseJWTWithoutVerification(tokenString string) (*JWTClaims, error) {
	// Parse the token without verification
	// In production, you should always verify the signature
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("JWT token is expired")
	}

	if claims.UserID == "" {
		return nil, fmt.Errorf("user_id is required in JWT")
	}

	if claims.Description == "" {
		return nil, fmt.Errorf("description is required in JWT")
	}

	return claims, nil
}

// ValidateJWTWithSecret validates JWT with a secret key (for production use)
func ValidateJWTWithSecret(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate JWT: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	return claims, nil
}

// CreateJWT creates a new JWT token with the specified claims
func CreateJWT(userID, description string, expiresAt time.Time, allowedDomains []string, secret string) (string, error) {
	issuedAt := time.Now()

	claims := &JWTClaims{
		UserID:         userID,
		Description:    description,
		AllowedDomains: allowedDomains,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			Issuer:    "go-cert-provider",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}
