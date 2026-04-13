package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dh-kam/go-cert-provider/cert/domain"
	"github.com/dh-kam/go-cert-provider/cert/registry"
	"github.com/dh-kam/go-cert-provider/graph/model"
	"github.com/dh-kam/go-cert-provider/session"
	"github.com/gin-gonic/gin"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{}

type contextKey string

const (
	ContextKeyGin          contextKey = "gin"
	ContextKeyJWTSecret    contextKey = "jwt_secret_key" //nolint:gosec // context key, not a credential
	ContextKeyCertRegistry contextKey = "cert_registry"
)

func getSessionFromContext(ctx context.Context) (*session.UserSession, error) {
	ginCtx, ok := ctx.Value(ContextKeyGin).(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("request context is unavailable")
	}

	sessionID, err := ginCtx.Cookie("session_id")
	if err != nil || sessionID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	userSession, exists := session.GetGlobalManager().GetSession(sessionID)
	if !exists {
		return nil, fmt.Errorf("session not found or expired")
	}

	return userSession, nil
}

func getRegistryFromContext(ctx context.Context) (*registry.CertificateProviderRegistry, error) {
	providerRegistry, ok := ctx.Value(ContextKeyCertRegistry).(*registry.CertificateProviderRegistry)
	if !ok || providerRegistry == nil {
		return nil, fmt.Errorf("certificate registry is unavailable")
	}

	return providerRegistry, nil
}

func isDomainAllowed(allowedDomains []string, candidate string) bool {
	for _, allowed := range allowedDomains {
		if allowed == "*" || allowed == candidate {
			return true
		}

		if !strings.HasPrefix(allowed, "*.") {
			continue
		}

		suffix := strings.TrimPrefix(allowed, "*.")
		if candidate == suffix || strings.HasSuffix(candidate, "."+suffix) {
			return true
		}
	}

	return false
}

func formatOptionalTime(t time.Time) *string {
	if t.IsZero() {
		return nil
	}

	formatted := t.Format(time.RFC3339)
	return &formatted
}

func toDomainModel(info domain.Info) *model.Domain {
	return &model.Domain{
		Name:       info.Name,
		Status:     info.Status,
		Provider:   info.Provider,
		CreateDate: formatOptionalTime(info.CreateDate),
		ExpireDate: formatOptionalTime(info.ExpireDate),
		AutoRenew:  info.AutoRenew,
	}
}

func isSecureRequest(ginCtx *gin.Context) bool {
	if ginCtx.Request.TLS != nil {
		return true
	}

	return strings.EqualFold(ginCtx.GetHeader("X-Forwarded-Proto"), "https")
}
