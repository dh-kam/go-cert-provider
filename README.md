# go-cert-provider

[![CI](https://github.com/dh-kam/go-cert-provider/actions/workflows/ci.yml/badge.svg)](https://github.com/dh-kam/go-cert-provider/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/dh-kam/go-cert-provider/branch/main/graph/badge.svg)](https://codecov.io/gh/dh-kam/go-cert-provider)
[![Go Report Card](https://goreportcard.com/badge/github.com/dh-kam/go-cert-provider)](https://goreportcard.com/report/github.com/dh-kam/go-cert-provider)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Certificate Provider Service with JWT Authentication

## Overview

This service provides TLS certificates from domain providers (Porkbun) to authorized users through JWT-based authentication. Users can retrieve certificates without having direct access to the provider's API keys.

### Problem It Solves

- **Secure Certificate Distribution**: Share certificates without exposing provider API credentials
- **JWT-Based Authorization**: Control who can access which domain certificates
- **Centralized Management**: Single point of certificate management for multiple domains
- **GraphQL API**: Easy integration with modern applications

### Use Case

Perfect for teams or organizations where:
- Multiple users/services need certificates from Porkbun domains
- You don't want to share Porkbun API keys with everyone
- You need audit trails and access control for certificate retrieval
- You want a GraphQL API for certificate management

## Features

- ** JWT Authentication**: Secure access control with token-based authentication
- ** Multi-Domain Support**: Manage certificates for multiple domains from a single service
- ** Provider Abstraction**: Clean architecture supporting multiple certificate providers (currently Porkbun)
- ** Auto-Discovery**: Automatically discover domains from provider account
- ** GraphQL API**: Modern API for certificate retrieval and management
- ** Health Check**: Built-in health monitoring endpoint
- ** Domain-Level Authorization**: JWT tokens specify which domains users can access

## Quick Start

### Prerequisites

- Go 1.25 or higher
- Porkbun account with API access
- Valid domain(s) registered with Porkbun

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd go-cert-provider

# Build (debug + release for current platform)
make

# Or build release only
make release

# Or build for specific platform
make linux-amd64-release
make windows-x86_64-release
make darwin-arm64-release

# Build all platforms
make build-all-release
```

### Configuration

#### Auto-Discovery (Recommended)

By default, the system will automatically discover all active domains from your Porkbun account:

```bash
# Only API credentials needed - domains auto-discovered
export PORKBUN_API_KEY="your-api-key"
export PORKBUN_SECRET_KEY="your-secret-key"

# Server will automatically manage all active domains in your account
./build/current/debug/go-cert-provider certs serve
```

#### Manual Domain Specification

You can manually specify which domains to manage:

```bash
# Porkbun Configuration
export PORKBUN_API_KEY="your-api-key"
export PORKBUN_SECRET_KEY="your-secret-key"
export PORKBUN_DOMAINS="example.com,*.example.com,test.com"

# JWT Secret (for authentication)
export JWT_SECRET_KEY="your-secret-key"

# Server Configuration
export LISTEN_ADDR="localhost"
export LISTEN_PORT="5000"
```

#### Using Command-Line Flags

All provider flags are available globally and can be used with any command:

```bash
# Start server with provider configuration
./build/current/debug/go-cert-provider certs serve \
  --porkbun-api-key "your-api-key" \
  --porkbun-secret-key "your-secret-key" \
  --porkbun-domains "example.com,*.example.com" \
  --jwt-secret-key "your-secret-key" \
  --listen-port 5000

# List domains with provider configuration
./build/current/debug/go-cert-provider domain list \
  --porkbun-api-key "your-api-key" \
  --porkbun-secret-key "your-secret-key"

# Retrieve certificate with provider configuration
./build/current/debug/go-cert-provider certs retrieve example.com \
  --porkbun-api-key "your-api-key" \
  --porkbun-secret-key "your-secret-key" \
  --porkbun-domains "example.com,test.com"
```

### Running the Server

The server requires two essential configurations:
1. **JWT Secret Key** - For token authentication
2. **At least one domain** - For certificate management

```bash
# Generate JWT secret key first
./build/current/debug/go-cert-provider jwt create-secret-key
export JWT_SECRET_KEY="your-generated-secret-key"

# Start the server with provider configuration
./build/current/debug/go-cert-provider certs serve \
  --porkbun-api-key "your-api-key" \
  --porkbun-secret-key "your-secret-key"

# Or with manually specified domains
./build/current/debug/go-cert-provider certs serve \
  --porkbun-api-key "your-api-key" \
  --porkbun-secret-key "your-secret-key" \
  --porkbun-domains "example.com,test.com"

# The server will start on http://localhost:5000
# GraphQL Playground: http://localhost:5000/
# GraphQL Endpoint: http://localhost:5000/graphql
# Health Check: http://localhost:5000/health
```

### Retrieving Certificates

```bash
# Retrieve certificate to stdout
./build/current/debug/go-cert-provider certs retrieve example.com

# Save certificate to files
./build/current/debug/go-cert-provider certs retrieve example.com --output-dir ./certs

# Save as separate files
./build/current/debug/go-cert-provider certs retrieve example.com \
  --output-dir ./certs \
  --separate-files
```

## Available Commands

```bash
# Domain management
./build/current/debug/go-cert-provider domain --help

# List all managed domains
./build/current/debug/go-cert-provider domain list
./build/current/debug/go-cert-provider domain list --detail
./build/current/debug/go-cert-provider domain list --output json

# Certificate management
./build/current/debug/go-cert-provider certs --help

# Start GraphQL server
./build/current/debug/go-cert-provider certs serve

# Retrieve certificate for a domain
./build/current/debug/go-cert-provider certs retrieve example.com

# JWT token management
./build/current/debug/go-cert-provider jwt --help

# Create JWT secret key
./build/current/debug/go-cert-provider jwt create-secret-key

# Create JWT token
./build/current/debug/go-cert-provider jwt create-token \
  --user-id "user123" \
  --description "API access for user" \
  --expires-at "2y" \
  --allowed-domains "example.com,test.com"

# Verify JWT token
./build/current/debug/go-cert-provider jwt verify-token "your-jwt-token"
```

## Adding a New Provider

1. Create a new provider package in `cert/providers/<provider-name>/`
2. Implement `domain.CertificateProvider` interface
3. Implement `domain.ProviderBootstrap` interface
4. Register the bootstrap in `cert/init.go`

## GraphQL API

### Authentication

```graphql
mutation Login {
  login(input: { apiKey: "your-jwt-token" }) {
    success
    message
    user {
      id
      description
    }
  }
}
```

### Queries

```graphql
# Health check
query Health {
  health {
    status
    timestamp
  }
}

# Version information
query Version {
  version {
    version
    buildTime
    gitCommit
  }
}

# Current user
query Me {
  me {
    id
    description
  }
}
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./cert/providers/porkbun/...
go test ./cert/registry/...

# Run with verbose output
go test -v ./...
```

### Building

```bash
# Using Makefile
make build

# Using go build directly
go build -o build/go-cert-provider ./main.go
```

## Environment Variables Reference

### Server Configuration
- `LISTEN_ADDR`: Server listen address (default: "localhost")
- `LISTEN_PORT`: Server listen port (default: 5000)
- `JWT_SECRET_KEY`: JWT secret key for authentication

### Porkbun Provider
- `PORKBUN_API_KEY`: Porkbun API key
- `PORKBUN_SECRET_KEY`: Porkbun secret key
- `PORKBUN_DOMAINS`: Comma-separated list of domains

## License

nullcode@gmail.com

## Contributing

Contributions are welcome! Please read the architecture documentation before submitting PRs.
