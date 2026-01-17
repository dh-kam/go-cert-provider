package porkbun

import (
	"testing"
)

func TestProviderImplementsInterface(t *testing.T) {
	provider := NewProvider("test-api-key", "test-secret", []string{"example.com"})

	if provider.GetProviderName() != "porkbun" {
		t.Errorf("Expected provider name 'porkbun', got '%s'", provider.GetProviderName())
	}

	domains := provider.GetDomains()
	if len(domains) != 1 || domains[0] != "example.com" {
		t.Errorf("Expected domains [example.com], got %v", domains)
	}
}

func TestProviderValidation(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		secretKey string
		domains   []string
		wantError bool
	}{
		{
			name:      "valid configuration",
			apiKey:    "test-api-key",
			secretKey: "test-secret",
			domains:   []string{"example.com"},
			wantError: false,
		},
		{
			name:      "missing api key",
			apiKey:    "",
			secretKey: "test-secret",
			domains:   []string{"example.com"},
			wantError: true,
		},
		{
			name:      "missing secret key",
			apiKey:    "test-api-key",
			secretKey: "",
			domains:   []string{"example.com"},
			wantError: true,
		},
		{
			name:      "missing domains",
			apiKey:    "test-api-key",
			secretKey: "test-secret",
			domains:   []string{},
			wantError: false, // Empty domains are allowed for auto-discovery
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(tt.apiKey, tt.secretKey, tt.domains)
			err := provider.ValidateConfiguration()

			if tt.wantError && err == nil {
				t.Error("Expected validation error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestParseDomains(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single domain",
			input:    "example.com",
			expected: []string{"example.com"},
		},
		{
			name:     "multiple domains",
			input:    "example.com,test.com,api.example.com",
			expected: []string{"example.com", "test.com", "api.example.com"},
		},
		{
			name:     "domains with spaces",
			input:    "example.com, test.com , api.example.com",
			expected: []string{"example.com", "test.com", "api.example.com"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "trailing comma",
			input:    "example.com,test.com,",
			expected: []string{"example.com", "test.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDomains(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d domains, got %d", len(tt.expected), len(result))
				return
			}

			for i, domain := range result {
				if domain != tt.expected[i] {
					t.Errorf("Expected domain[%d] = '%s', got '%s'", i, tt.expected[i], domain)
				}
			}
		})
	}
}
