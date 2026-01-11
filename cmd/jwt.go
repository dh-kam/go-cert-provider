package cmd

import (
	"github.com/spf13/cobra"
)

// jwtCmd represents the jwt command
var jwtCmd = &cobra.Command{
	Use:   "jwt",
	Short: "JWT token management commands",
	Long: `Manage JWT tokens for authentication and authorization.

This command provides subcommands for creating secret keys, generating tokens,
and verifying token validity.`,
}

func init() {
	rootCmd.AddCommand(jwtCmd)
}
