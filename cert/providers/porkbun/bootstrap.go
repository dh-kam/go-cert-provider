package porkbun

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dh-kam/go-cert-provider/cert/domain"
	"github.com/spf13/cobra"
)

const (
	envAPIKey    = "PORKBUN_API_KEY"    //nolint:gosec // not a credential
	envSecretKey = "PORKBUN_SECRET_KEY" //nolint:gosec // not a credential
	envDomains   = "PORKBUN_DOMAINS"    // Optional: manually specify domains
)

// Bootstrap implements domain.ProviderBootstrap for Porkbun
type Bootstrap struct {
	apiKey    string
	secretKey string
	domains   string // Comma-separated list of domains (optional)
}

// NewBootstrap creates a new Porkbun bootstrap
func NewBootstrap() *Bootstrap {
	return &Bootstrap{}
}

// GetProviderName returns the provider name
func (b *Bootstrap) GetProviderName() string {
	return "porkbun"
}

// RegisterFlags registers command-line flags for Porkbun provider
func (b *Bootstrap) RegisterFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()

	flags.StringVar(&b.apiKey, "porkbun-api-key", "",
		"Porkbun API key (overrides PORKBUN_API_KEY env var)")
	flags.StringVar(&b.secretKey, "porkbun-secret-key", "",
		"Porkbun secret key (overrides PORKBUN_SECRET_KEY env var)")
	flags.StringVar(&b.domains, "porkbun-domains", "",
		"Comma-separated list of domains (optional, if not specified all domains from account will be used)")
}

// IsConfigured checks if the provider is configured
func (b *Bootstrap) IsConfigured() bool {
	apiKey := b.getAPIKey()
	secretKey := b.getSecretKey()

	// Only API key and secret key are required
	// Domains are optional - will be auto-discovered if not specified
	return apiKey != "" && secretKey != ""
}

// CreateProvider creates a configured Porkbun provider instance
func (b *Bootstrap) CreateProvider() (domain.CertificateProvider, error) {
	apiKey := b.getAPIKey()
	secretKey := b.getSecretKey()
	domainsStr := b.getDomains()

	if apiKey == "" {
		return nil, fmt.Errorf("Porkbun API key not configured (set PORKBUN_API_KEY env var or --porkbun-api-key flag)")
	}

	if secretKey == "" {
		return nil, fmt.Errorf("Porkbun secret key not configured (set PORKBUN_SECRET_KEY env var or --porkbun-secret-key flag)")
	}

	var domains []string
	var domainInfos []domain.Info

	if domainsStr != "" {
		// User specified domains manually
		domains = parseDomains(domainsStr)
		if len(domains) == 0 {
			return nil, fmt.Errorf("no valid domains specified for Porkbun")
		}

		// Create basic domain info for manually specified domains
		for _, d := range domains {
			domainInfos = append(domainInfos, domain.Info{
				Name:     d,
				Provider: "porkbun",
				Status:   "CONFIGURED",
			})
		}
	} else {
		// Auto-discover domains from Porkbun account
		client := NewClient(apiKey, secretKey)

		// Test connection first
		if _, err := client.Ping(); err != nil {
			return nil, fmt.Errorf("failed to connect to Porkbun API: %w", err)
		}

		// Retrieve all domains
		porkbunDomains, err := client.ListDomains()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve domains from Porkbun: %w", err)
		}

		if len(porkbunDomains) == 0 {
			return nil, fmt.Errorf("no domains found in Porkbun account")
		}

		// Extract domain names (only ACTIVE domains) and create domain info
		for _, d := range porkbunDomains {
			if d.Status == "ACTIVE" {
				domains = append(domains, d.Domain)

				// Parse dates
				createDate := parseDate(d.CreateDate)
				expireDate := parseDate(d.ExpireDate)

				domainInfos = append(domainInfos, domain.Info{
					Name:       d.Domain,
					Provider:   "porkbun",
					Status:     d.Status,
					CreateDate: createDate,
					ExpireDate: expireDate,
					AutoRenew:  false, // Porkbun API doesn't provide this in listAll
				})
			}
		}

		if len(domains) == 0 {
			return nil, fmt.Errorf("no active domains found in Porkbun account")
		}
	}

	provider := NewProvider(apiKey, secretKey, domains)

	// Set domain info
	provider.SetDomainInfos(domainInfos)

	// Validate configuration
	if err := provider.ValidateConfiguration(); err != nil {
		return nil, fmt.Errorf("Porkbun provider validation failed: %w", err)
	}

	return provider, nil
}

// getAPIKey returns the API key from flag or environment
func (b *Bootstrap) getAPIKey() string {
	if b.apiKey != "" {
		return b.apiKey
	}
	return os.Getenv(envAPIKey)
}

// getSecretKey returns the secret key from flag or environment
func (b *Bootstrap) getSecretKey() string {
	if b.secretKey != "" {
		return b.secretKey
	}
	return os.Getenv(envSecretKey)
}

// getDomains returns the domains string from flag or environment
func (b *Bootstrap) getDomains() string {
	if b.domains != "" {
		return b.domains
	}
	return os.Getenv(envDomains)
}

// parseDomains parses a comma-separated list of domains
func parseDomains(domainsStr string) []string {
	parts := strings.Split(domainsStr, ",")
	domains := make([]string, 0, len(parts))

	for _, part := range parts {
		domain := strings.TrimSpace(part)
		if domain != "" {
			domains = append(domains, domain)
		}
	}

	return domains
}

// parseDate parses Porkbun date format (YYYY-MM-DD HH:MM:SS)
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	// Porkbun format: "2018-08-20 17:52:51"
	t, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		// If parsing fails, return zero time
		return time.Time{}
	}

	return t
}
