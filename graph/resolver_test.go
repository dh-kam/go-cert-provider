package graph

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	certdomain "github.com/dh-kam/go-cert-provider/cert/domain"
	"github.com/dh-kam/go-cert-provider/cert/registry"
	"github.com/dh-kam/go-cert-provider/session"
	"github.com/gin-gonic/gin"
)

type fakeProvider struct {
	name        string
	domains     []string
	domainInfos map[string]*certdomain.Info
	certChain   []byte
	privateKey  []byte
}

func (p *fakeProvider) GetProviderName() string {
	return p.name
}

func (p *fakeProvider) GetDomains() []string {
	return p.domains
}

func (p *fakeProvider) GetDomainInfo(domain string) *certdomain.Info {
	return p.domainInfos[domain]
}

func (p *fakeProvider) ListDomainInfo() []certdomain.Info {
	result := make([]certdomain.Info, 0, len(p.domainInfos))
	for _, info := range p.domainInfos {
		result = append(result, *info)
	}

	return result
}

func (p *fakeProvider) RetrieveCertificate(domain string) ([]byte, []byte, error) {
	return p.certChain, p.privateKey, nil
}

func (p *fakeProvider) ValidateConfiguration() error {
	return nil
}

func makeResolverContext(t *testing.T, allowedDomains []string, provider *fakeProvider) context.Context {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest("POST", "/graphql", nil)

	sessionID := session.GetGlobalManager().CreateSession(
		"user-1",
		"test user",
		time.Now().Add(time.Hour),
		allowedDomains,
	)
	t.Cleanup(func() {
		session.GetGlobalManager().DeleteSession(sessionID)
	})

	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	ginCtx.Request = req

	providerRegistry := registry.NewCertificateProviderRegistry()
	if err := providerRegistry.Register(provider); err != nil {
		t.Fatalf("failed to register fake provider: %v", err)
	}

	ctx := context.WithValue(context.Background(), ContextKeyGin, ginCtx)
	ctx = context.WithValue(ctx, ContextKeyCertRegistry, providerRegistry)

	return ctx
}

func TestIsDomainAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		allowedDomains []string
		candidate      string
		want           bool
	}{
		{name: "exact match", allowedDomains: []string{"example.com"}, candidate: "example.com", want: true},
		{name: "wildcard suffix", allowedDomains: []string{"*.example.com"}, candidate: "api.example.com", want: true},
		{name: "wildcard apex", allowedDomains: []string{"*.example.com"}, candidate: "example.com", want: true},
		{name: "global wildcard", allowedDomains: []string{"*"}, candidate: "anything.com", want: true},
		{name: "not allowed", allowedDomains: []string{"test.com"}, candidate: "example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isDomainAllowed(tt.allowedDomains, tt.candidate); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestDomainsFiltersBySessionAllowedDomains(t *testing.T) {
	provider := &fakeProvider{
		name:    "fake",
		domains: []string{"example.com", "test.com"},
		domainInfos: map[string]*certdomain.Info{
			"example.com": {Name: "example.com", Provider: "fake", Status: "ACTIVE"},
			"test.com":    {Name: "test.com", Provider: "fake", Status: "ACTIVE"},
		},
	}

	ctx := makeResolverContext(t, []string{"example.com"}, provider)

	resolver := &queryResolver{&Resolver{}}
	domains, err := resolver.Domains(ctx)
	if err != nil {
		t.Fatalf("domains query failed: %v", err)
	}

	if len(domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(domains))
	}

	if domains[0].Name != "example.com" {
		t.Fatalf("expected example.com, got %s", domains[0].Name)
	}
}

func TestCertificateRequiresAllowedDomain(t *testing.T) {
	provider := &fakeProvider{
		name:        "fake",
		domains:     []string{"example.com"},
		domainInfos: map[string]*certdomain.Info{"example.com": {Name: "example.com", Provider: "fake", Status: "ACTIVE"}},
		certChain:   []byte("cert"),
		privateKey:  []byte("key"),
	}

	ctx := makeResolverContext(t, []string{"test.com"}, provider)

	resolver := &queryResolver{&Resolver{}}
	_, err := resolver.Certificate(ctx, "example.com")
	if err == nil {
		t.Fatal("expected access denied error")
	}
}

func TestCertificateReturnsMaterialForAllowedDomain(t *testing.T) {
	provider := &fakeProvider{
		name:        "fake",
		domains:     []string{"example.com"},
		domainInfos: map[string]*certdomain.Info{"example.com": {Name: "example.com", Provider: "fake", Status: "ACTIVE"}},
		certChain:   []byte("cert"),
		privateKey:  []byte("key"),
	}

	ctx := makeResolverContext(t, []string{"example.com"}, provider)

	resolver := &queryResolver{&Resolver{}}
	result, err := resolver.Certificate(ctx, "example.com")
	if err != nil {
		t.Fatalf("certificate query failed: %v", err)
	}

	if result.Domain != "example.com" || result.CertificateChain != "cert" || result.PrivateKey != "key" {
		t.Fatalf("unexpected certificate payload: %+v", result)
	}
}
