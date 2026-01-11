package config

import (
	"fmt"
	"os"
	"strconv"
)

var (
	// Version is the current version of the application (set at build time)
	Version = "dev"
	// BuildTime is the build timestamp (set at build time)
	BuildTime = "unknown"
	// GitCommit is the git commit hash (set at build time)
	GitCommit = "unknown"
)

const (
	// DefaultPort is the default port number
	DefaultPort = 5000
)

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port int
	Addr string
}

// NewServerConfig creates a new server configuration
func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{
		Port: DefaultPort,
		Addr: "localhost",
	}

	// Check environment variables first
	if portStr := os.Getenv("LISTEN_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = int(port)
		}
	}

	if addr := os.Getenv("LISTEN_ADDR"); addr != "" {
		cfg.Addr = addr
	}

	return cfg
}

// SetPort sets the port number
func (c *ServerConfig) SetPort(port int) {
	c.Port = port
}

// SetAddr sets the listen address
func (c *ServerConfig) SetAddr(addr string) {
	c.Addr = addr
}

// GetListenAddr returns the full listen address string
func (c *ServerConfig) GetListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Addr, c.Port)
} 