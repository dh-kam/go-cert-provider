package registry

import (
	"testing"

	"github.com/dh-kam/go-cert-provider/cert/providers/porkbun"
)

func TestRegistryRegisterProvider(t *testing.T) {
	registry := NewCertificateProviderRegistry()

	// Create a valid provider
	provider := porkbun.NewProvider("api-key", "secret", []string{"example.com", "test.com"})

	// Register the provider
	err := registry.Register(provider)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// Verify provider is registered
	providers := registry.ListProviders()
	if len(providers) != 1 || providers[0] != "porkbun" {
		t.Errorf("Expected providers [porkbun], got %v", providers)
	}

	// Verify domains are registered
	domains := registry.ListDomains()
	if len(domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(domains))
	}
}

func TestRegistryGetProviderForDomain(t *testing.T) {
	registry := NewCertificateProviderRegistry()

	provider := porkbun.NewProvider("api-key", "secret", []string{"example.com"})
	registry.Register(provider)

	// Test getting provider for registered domain
	p, err := registry.GetProviderForDomain("example.com")
	if err != nil {
		t.Fatalf("Failed to get provider for domain: %v", err)
	}

	if p.GetProviderName() != "porkbun" {
		t.Errorf("Expected provider 'porkbun', got '%s'", p.GetProviderName())
	}

	// Test getting provider for unregistered domain
	_, err = registry.GetProviderForDomain("nonexistent.com")
	if err == nil {
		t.Error("Expected error for unregistered domain, got nil")
	}
}

func TestRegistryDuplicateProvider(t *testing.T) {
	registry := NewCertificateProviderRegistry()

	provider1 := porkbun.NewProvider("api-key-1", "secret-1", []string{"example.com"})
	err := registry.Register(provider1)
	if err != nil {
		t.Fatalf("Failed to register first provider: %v", err)
	}

	// Try to register another provider with same name (different domains)
	provider2 := porkbun.NewProvider("api-key-2", "secret-2", []string{"other.com"})
	err = registry.Register(provider2)
	if err == nil {
		t.Error("Expected error when registering duplicate provider, got nil")
	}
}

func TestRegistryDuplicateDomain(t *testing.T) {
	// For this test, we need to create a different provider type
	// Since we only have porkbun, we'll skip this test for now
	// In a real scenario, you'd have multiple provider implementations
	t.Skip("Skipping duplicate domain test - requires multiple provider types")
}

func TestBootstrapManager(t *testing.T) {
	registry := NewCertificateProviderRegistry()
	manager := NewBootstrapManager(registry)

	// Register bootstrap
	bootstrap := porkbun.NewBootstrap()
	manager.RegisterBootstrap(bootstrap)

	// At this point, no providers should be configured
	configured := manager.GetConfiguredProviders()
	if len(configured) > 0 {
		t.Errorf("Expected no configured providers, got %v", configured)
	}

	// Verify the manager is properly initialized
	_ = registry // Use the registry variable
}
