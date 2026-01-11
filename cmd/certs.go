package cmd

import (
	"github.com/spf13/cobra"
)

// certsCmd represents the certs command
var certsCmd = &cobra.Command{
	Use:   "certs",
	Short: "Certificate management commands",
	Long: `Manage SSL/TLS certificates from various domain service providers.

This command provides subcommands for serving certificates via GraphQL API
and retrieving certificates directly from the command line.`,
}

func init() {
	rootCmd.AddCommand(certsCmd)
}
