package cmd

import (
	"github.com/spf13/cobra"
)

// domainCmd represents the domain command
var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Domain management commands",
	Long: `Manage domains from various service providers.

This command provides subcommands for listing and managing domains
that are configured with certificate providers.`,
}

func init() {
	rootCmd.AddCommand(domainCmd)
}
