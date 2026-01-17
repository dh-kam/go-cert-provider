package registry

import (
	"fmt"
	"sync"

	"github.com/dh-kam/go-cert-provider/cert/domain"
)

// CertificateProviderRegistry manages all registered certificate providers
type CertificateProviderRegistry struct {
	providers map[string]domain.CertificateProvider // key: provider name
	domainMap map[string]domain.CertificateProvider // key: domain name
	mu        sync.RWMutex
}

// NewCertificateProviderRegistry creates a new registry
func NewCertificateProviderRegistry() *CertificateProviderRegistry {
	return &CertificateProviderRegistry{
		providers: make(map[string]domain.CertificateProvider),
		domainMap: make(map[string]domain.CertificateProvider),
	}
}

// Register registers a new certificate provider
func (r *CertificateProviderRegistry) Register(provider domain.CertificateProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	providerName := provider.GetProviderName()

	if _, exists := r.providers[providerName]; exists {
		return fmt.Errorf("provider %s is already registered", providerName)
	}

	if err := provider.ValidateConfiguration(); err != nil {
		return fmt.Errorf("provider %s configuration invalid: %w", providerName, err)
	}

	r.providers[providerName] = provider

	for _, domain := range provider.GetDomains() {
		if existingProvider, exists := r.domainMap[domain]; exists {
			return fmt.Errorf("domain %s is already managed by provider %s",
				domain, existingProvider.GetProviderName())
		}
		r.domainMap[domain] = provider
	}

	return nil
}

// GetProviderForDomain returns the provider managing the specified domain
func (r *CertificateProviderRegistry) GetProviderForDomain(domain string) (domain.CertificateProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.domainMap[domain]
	if !exists {
		return nil, fmt.Errorf("no provider found for domain: %s", domain)
	}

	return provider, nil
}

// GetProvider returns a provider by name
func (r *CertificateProviderRegistry) GetProvider(providerName string) (domain.CertificateProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	return provider, nil
}

// ListProviders returns all registered provider names
func (r *CertificateProviderRegistry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListDomains returns all managed domains
func (r *CertificateProviderRegistry) ListDomains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domains := make([]string, 0, len(r.domainMap))
	for domain := range r.domainMap {
		domains = append(domains, domain)
	}
	return domains
}

// RetrieveCertificate retrieves the certificate for the specified domain
func (r *CertificateProviderRegistry) RetrieveCertificate(domain string) ([]byte, []byte, error) {
	provider, err := r.GetProviderForDomain(domain)
	if err != nil {
		return nil, nil, err
	}

	return provider.RetrieveCertificate(domain)
}

// GetDomainInfo returns detailed information about a specific domain
func (r *CertificateProviderRegistry) GetDomainInfo(domainName string) *domain.Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.domainMap[domainName]
	if !exists {
		return nil
	}

	return provider.GetDomainInfo(domainName)
}

// ListAllDomainInfo returns detailed information for all managed domains
func (r *CertificateProviderRegistry) ListAllDomainInfo() []domain.Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allInfos []domain.Info

	for _, provider := range r.providers {
		infos := provider.ListDomainInfo()
		allInfos = append(allInfos, infos...)
	}

	return allInfos
}
