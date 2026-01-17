package porkbun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiBaseURL = "https://api.porkbun.com/api/json/v3"
)

// Client represents a Porkbun API client
type Client struct {
	apiKey     string
	secretKey  string
	httpClient *http.Client
}

// NewClient creates a new Porkbun API client
func NewClient(apiKey, secretKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		secretKey:  secretKey,
		httpClient: &http.Client{},
	}
}

// Domain represents a domain from Porkbun API
type Domain struct {
	Domain     string `json:"domain"`
	Status     string `json:"status"`
	TLD        string `json:"tld"`
	CreateDate string `json:"createDate"`
	ExpireDate string `json:"expireDate"`
}

// ListDomainsResponse represents the response from domain list API
type ListDomainsResponse struct {
	Status  string   `json:"status"`
	Domains []Domain `json:"domains"`
}

// SSLResponse represents the response from SSL retrieve API
type SSLResponse struct {
	Status           string `json:"status"`
	CertificateChain string `json:"certificatechain"`
	PrivateKey       string `json:"privatekey"`
	PublicKey        string `json:"publickey"`
}

// PingResponse represents the response from ping API
type PingResponse struct {
	Status string `json:"status"`
	YourIP string `json:"yourIp"`
}

// authRequest is the base request structure with authentication
type authRequest struct {
	SecretAPIKey string `json:"secretapikey"`
	APIKey       string `json:"apikey"`
}

// makeRequest makes an authenticated request to Porkbun API
func (c *Client) makeRequest(endpoint string, result interface{}) error {
	reqBody := authRequest{
		SecretAPIKey: c.secretKey,
		APIKey:       c.apiKey,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := apiBaseURL + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API returned status %d (failed to read body: %w)", resp.StatusCode, err)
		}
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// Ping tests the API connection and returns the client's IP address
func (c *Client) Ping() (*PingResponse, error) {
	var result PingResponse
	if err := c.makeRequest("/ping", &result); err != nil {
		return nil, err
	}

	if result.Status != "SUCCESS" {
		return nil, fmt.Errorf("ping failed: %s", result.Status)
	}

	return &result, nil
}

// ListDomains retrieves all domains in the account
func (c *Client) ListDomains() ([]Domain, error) {
	var result ListDomainsResponse
	if err := c.makeRequest("/domain/listAll", &result); err != nil {
		return nil, err
	}

	if result.Status != "SUCCESS" {
		return nil, fmt.Errorf("list domains failed: %s", result.Status)
	}

	return result.Domains, nil
}

// RetrieveSSL retrieves the SSL certificate for a domain
func (c *Client) RetrieveSSL(domain string) (*SSLResponse, error) {
	var result SSLResponse
	endpoint := fmt.Sprintf("/ssl/retrieve/%s", domain)

	if err := c.makeRequest(endpoint, &result); err != nil {
		return nil, err
	}

	if result.Status != "SUCCESS" {
		return nil, fmt.Errorf("SSL retrieval failed: %s", result.Status)
	}

	return &result, nil
}
