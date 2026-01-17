package domain

import (
	"time"

	"github.com/spf13/cobra"
)

// Info contains detailed information about a domain
type Info struct {
	Name       string    // Domain name
	Status     string    // Domain status (ACTIVE, EXPIRED, etc.)
	Provider   string    // Provider name
	CreateDate time.Time // When the domain was created
	ExpireDate time.Time // When the domain expires
	AutoRenew  bool      // Whether auto-renewal is enabled
}

// CertificateProvider is the interface that all domain service providers must implement
// to provide wildcard SSL certificates for their managed domains
type CertificateProvider interface {
	// GetProviderName returns the unique name of the provider (e.g., "porkbun", "cloudflare")
	GetProviderName() string

	// GetDomains returns the list of domains this provider manages
	GetDomains() []string

	// GetDomainInfo returns detailed information about a specific domain
	// Returns nil if domain is not managed by this provider
	GetDomainInfo(domain string) *Info

	// ListDomainInfo returns detailed information for all managed domains
	ListDomainInfo() []Info

	// RetrieveCertificate retrieves the SSL certificate for the specified domain
	// Returns certificate chain, private key, and error
	RetrieveCertificate(domain string) (certChain []byte, privateKey []byte, err error)

	// ValidateConfiguration validates the provider's configuration
	ValidateConfiguration() error
}

// ProviderBootstrap is the interface for bootstrapping providers
// Each provider implementation should have a corresponding bootstrap that knows
// how to initialize the provider from environment variables and command-line options
type ProviderBootstrap interface {
	// GetProviderName returns the name of the provider this bootstrap creates
	GetProviderName() string

	// RegisterFlags registers command-line flags specific to this provider
	RegisterFlags(cmd *cobra.Command)

	// IsConfigured checks if the provider is configured (via env vars or flags)
	// Returns true if all required configuration is present
	IsConfigured() bool

	// CreateProvider creates and returns a configured provider instance
	// Returns error if configuration is invalid
	CreateProvider() (CertificateProvider, error)
}
