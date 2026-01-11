package cert

import (
	"github.com/dh-kam/go-cert-provider/cert/providers/porkbun"
	"github.com/dh-kam/go-cert-provider/cert/registry"
	"github.com/spf13/cobra"
)

var (
	// Package-level instances to avoid re-initialization
	globalProviderRegistry *registry.CertificateProviderRegistry
	globalBootstrapManager *registry.BootstrapManager
)

// InitializeCertificateSystem creates and configures the certificate provider system
func InitializeCertificateSystem(cmd *cobra.Command) (*registry.CertificateProviderRegistry, *registry.BootstrapManager, error) {
	// Return existing instances if already initialized
	if globalProviderRegistry != nil && globalBootstrapManager != nil {
		return globalProviderRegistry, globalBootstrapManager, nil
	}

	// Create registry
	globalProviderRegistry = registry.NewCertificateProviderRegistry()
	
	// Create bootstrap manager
	globalBootstrapManager = registry.NewBootstrapManager(globalProviderRegistry)
	
	// Register all provider bootstraps
	globalBootstrapManager.RegisterBootstrap(porkbun.NewBootstrap())
	// Future providers can be registered here:
	// globalBootstrapManager.RegisterBootstrap(cloudflare.NewBootstrap())
	// globalBootstrapManager.RegisterBootstrap(route53.NewBootstrap())
	
	return globalProviderRegistry, globalBootstrapManager, nil
}
