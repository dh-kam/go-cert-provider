package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var createSecretKeyCmd = &cobra.Command{
	Use:   "create-secret-key",
	Short: "Generate a random JWT secret key",
	Long:  "Generate a cryptographically secure random string for JWT signing",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			_ = cmd.Usage()
			return fmt.Errorf("this command does not take any arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Generate 32 bytes (256 bits) of random data
		// This is recommended for HMAC-SHA256
		secretBytes := make([]byte, 32)

		_, err := rand.Read(secretBytes)
		if err != nil {
			return fmt.Errorf("failed to generate random secret: %w", err)
		}

		// Encode to base64 for easy use
		secretKey := base64.StdEncoding.EncodeToString(secretBytes)

		// Define styles
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

		greenStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

		usageStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))

		fmt.Println(titleStyle.Render("Generated JWT Secret Key (base64 encoded):"))
		fmt.Println("   ", greenStyle.Render(secretKey))
		fmt.Println()
		fmt.Println(usageStyle.Render("Usage:"))
		fmt.Println("    Environment variable:")
		fmt.Println("        ", greenStyle.Render(fmt.Sprintf("export JWT_SECRET_KEY=\"%s\"", secretKey)))
		fmt.Println("    Command line option:")
		fmt.Println("        ", greenStyle.Render(fmt.Sprintf("--jwt-secret-key \"%s\"", secretKey)))

		return nil
	},
}

func init() {
	jwtCmd.AddCommand(createSecretKeyCmd)
}
