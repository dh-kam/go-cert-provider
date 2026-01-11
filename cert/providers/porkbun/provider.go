package porkbun

import (
	"fmt"
	"strings"

	"github.com/dh-kam/go-cert-provider/cert/domain"
)

var _ domain.CertificateProvider = (*Provider)(nil)

// Provider implements domain.CertificateProvider for Porkbun domain service
type Provider struct {
	apiKey      string
	secretKey   string
	domains     []string
	domainInfos map[string]*domain.DomainInfo // Map of domain name to info
	client      *Client
}

// NewProvider creates a new Porkbun certificate provider
func NewProvider(apiKey, secretKey string, domains []string) *Provider {
	return &Provider{
		apiKey:      apiKey,
		secretKey:   secretKey,
		domains:     domains,
		domainInfos: make(map[string]*domain.DomainInfo),
		client:      NewClient(apiKey, secretKey),
	}
}

// SetDomainInfos sets the domain information (called by bootstrap)
func (p *Provider) SetDomainInfos(infos []domain.DomainInfo) {
	p.domainInfos = make(map[string]*domain.DomainInfo)
	for i := range infos {
		p.domainInfos[infos[i].Name] = &infos[i]
	}
}

// GetProviderName returns the provider name
func (p *Provider) GetProviderName() string {
	return "porkbun"
}

// GetDomains returns the list of domains this provider manages
func (p *Provider) GetDomains() []string {
	return p.domains
}

// GetDomainInfo returns detailed information about a specific domain
func (p *Provider) GetDomainInfo(domainName string) *domain.DomainInfo {
	info, exists := p.domainInfos[domainName]
	if !exists {
		// Return basic info if detailed info not available
		for _, d := range p.domains {
			if d == domainName {
				return &domain.DomainInfo{
					Name:     domainName,
					Provider: p.GetProviderName(),
					Status:   "UNKNOWN",
				}
			}
		}
		return nil
	}
	return info
}

// ListDomainInfo returns detailed information for all managed domains
func (p *Provider) ListDomainInfo() []domain.DomainInfo {
	infos := make([]domain.DomainInfo, 0, len(p.domains))
	for _, domainName := range p.domains {
		if info := p.GetDomainInfo(domainName); info != nil {
			infos = append(infos, *info)
		}
	}
	return infos
}

// RetrieveCertificate retrieves the SSL certificate for the specified domain
func (p *Provider) RetrieveCertificate(domain string) ([]byte, []byte, error) {
	// Check if domain is managed by this provider
	found := false
	for _, d := range p.domains {
		if d == domain {
			found = true
			break
		}
	}
	if !found {
		return nil, nil, fmt.Errorf("domain %s is not managed by this provider", domain)
	}

	// Retrieve certificate from Porkbun API
	sslResp, err := p.client.RetrieveSSL(domain)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve SSL certificate: %w", err)
	}

	// Convert to byte slices
	certChain := []byte(sslResp.CertificateChain)
	privateKey := []byte(sslResp.PrivateKey)

	return certChain, privateKey, nil
}

// ValidateConfiguration validates the provider's configuration
func (p *Provider) ValidateConfiguration() error {
	var missingFields []string

	if p.apiKey == "" {
		missingFields = append(missingFields, "api-key")
	}

	if p.secretKey == "" {
		missingFields = append(missingFields, "secret-key")
	}

	// Note: domains can be empty - they will be auto-discovered

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required Porkbun fields: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// GetAPIKey returns the API key (for internal use)
func (p *Provider) GetAPIKey() string {
	return p.apiKey
}

// GetSecretKey returns the secret key (for internal use)
func (p *Provider) GetSecretKey() string {
	return p.secretKey
}
