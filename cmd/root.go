package cmd

import (
	"fmt"

	"github.com/dh-kam/go-cert-provider/cert"
	"github.com/dh-kam/go-cert-provider/cert/registry"
	"github.com/spf13/cobra"
)

// Shared state for commands that need provider access
type globalState struct {
	providerRegistry *registry.CertificateProviderRegistry
	bootstrapManager *registry.BootstrapManager
}

var (
	appState *globalState

	rootCmd = &cobra.Command{
		Use:   "go-cert-provider",
		Short: "Certificate provider service with JWT authentication",
		Long: `A service that provides TLS certificates from domain providers (Porkbun) 
to authorized users via JWT authentication.

This tool allows users to retrieve certificates without exposing provider API keys,
using JWT tokens for authentication and authorization.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip provider initialization for commands that don't need it
			cmdPath := cmd.CommandPath()
			skipProviderInit := false

			// Commands that don't need provider initialization
			skipCommands := []string{
				"go-cert-provider jwt",
				"go-cert-provider version",
				"go-cert-provider help",
				"go-cert-provider completion",
			}

			for _, skipCmd := range skipCommands {
				if len(cmdPath) >= len(skipCmd) && cmdPath[:len(skipCmd)] == skipCmd {
					skipProviderInit = true
					break
				}
			}

			if skipProviderInit {
				return nil
			}

			// Initialize certificate provider system
			providerRegistry, bootstrapManager, err := cert.InitializeCertificateSystem(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialize certificate system: %w", err)
			}

			// Initialize all configured providers
			if err := bootstrapManager.InitializeProviders(); err != nil {
				return fmt.Errorf("failed to initialize providers: %w", err)
			}

			// Store in global state for subcommands to use
			appState = &globalState{
				providerRegistry: providerRegistry,
				bootstrapManager: bootstrapManager,
			}

			return nil
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Initialize certificate system to register provider flags
	_, bootstrapManager, err := cert.InitializeCertificateSystem(rootCmd)
	if err != nil {
		fmt.Fprintf(rootCmd.OutOrStderr(), "Warning: failed to initialize certificate system: %v\n", err)
	} else {
		// Register all provider flags as persistent flags at root level
		// These will be available to all subcommands
		bootstrapManager.RegisterFlags(rootCmd)
	}
}
