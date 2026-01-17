package registry

import (
	"fmt"

	"github.com/dh-kam/go-cert-provider/cert/domain"
	"github.com/spf13/cobra"
)

// BootstrapManager manages provider bootstraps
type BootstrapManager struct {
	bootstraps []domain.ProviderBootstrap
	registry   *CertificateProviderRegistry
}

// NewBootstrapManager creates a new bootstrap manager
func NewBootstrapManager(registry *CertificateProviderRegistry) *BootstrapManager {
	return &BootstrapManager{
		bootstraps: make([]domain.ProviderBootstrap, 0),
		registry:   registry,
	}
}

// RegisterBootstrap registers a provider bootstrap
func (bm *BootstrapManager) RegisterBootstrap(bootstrap domain.ProviderBootstrap) {
	bm.bootstraps = append(bm.bootstraps, bootstrap)
}

// RegisterFlags registers all provider flags to the command
func (bm *BootstrapManager) RegisterFlags(cmd *cobra.Command) {
	for _, bootstrap := range bm.bootstraps {
		bootstrap.RegisterFlags(cmd)
	}
}

// InitializeProviders initializes all configured providers and registers them
func (bm *BootstrapManager) InitializeProviders() error {
	configuredCount := 0

	for _, bootstrap := range bm.bootstraps {
		if !bootstrap.IsConfigured() {
			continue
		}

		provider, err := bootstrap.CreateProvider()
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %w",
				bootstrap.GetProviderName(), err)
		}

		if err := bm.registry.Register(provider); err != nil {
			return fmt.Errorf("failed to register provider %s: %w",
				bootstrap.GetProviderName(), err)
		}

		configuredCount++
	}

	if configuredCount == 0 {
		return fmt.Errorf("no certificate providers configured")
	}

	return nil
}

// GetConfiguredProviders returns a list of configured provider names
func (bm *BootstrapManager) GetConfiguredProviders() []string {
	configured := make([]string, 0)

	for _, bootstrap := range bm.bootstraps {
		if bootstrap.IsConfigured() {
			configured = append(configured, bootstrap.GetProviderName())
		}
	}

	return configured
}
