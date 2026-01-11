package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dh-kam/go-cert-provider/auth"
	"github.com/dh-kam/go-cert-provider/utils"
	"github.com/spf13/cobra"
)

type verifyJwtTokenOptions struct {
	jwtSecretKey string
}

var verifyTokenCmd = &cobra.Command{
	Use:   "verify-token [token]",
	Short: "Verify a JWT token",
	Long:  "Verify a JWT token and display its claims",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := args[0]
		
		options, ok := cmd.Context().Value(KeyForOptions).(*verifyJwtTokenOptions)
		if !ok {
			return fmt.Errorf("failed to get command options from context")
		}

		jwtSecretKey := options.jwtSecretKey
		if jwtSecretKey == "" {
			jwtSecretKey = os.Getenv("JWT_SECRET_KEY")
		}

		claims, err := auth.ParseJWT(token, jwtSecretKey)
		if err != nil {
			fmt.Printf("❌ Token verification failed: %v\n", err)
			return nil
		}

		fmt.Printf("✅ Token verification successful!\n\n")
		fmt.Printf("Claims:\n")
		fmt.Printf("  User ID: %s\n", claims.UserID)
		fmt.Printf("  Description: %s\n", claims.Description)
		fmt.Printf("  Allowed Domains: %s\n", strings.Join(claims.AllowedDomains, ", "))
		
		if claims.ExpiresAt != nil {
			fmt.Printf("  Expires At: %s\n", utils.FormatDateTime(claims.ExpiresAt.Time))
			
			if time.Now().After(claims.ExpiresAt.Time) {
				fmt.Printf("  Status: ⚠️  EXPIRED\n")
			} else {
				timeLeft := time.Until(claims.ExpiresAt.Time)
				fmt.Printf("  Status: ✅ Valid (expires in %s)\n", utils.FormatDuration(timeLeft))
			}
		}
		
		if claims.IssuedAt != nil {
			fmt.Printf("  Issued At: %s\n", utils.FormatDateTime(claims.IssuedAt.Time))
		}
		
		if claims.NotBefore != nil {
			fmt.Printf("  Not Before: %s\n", utils.FormatDateTime(claims.NotBefore.Time))
		}
		
		if claims.Issuer != "" {
			fmt.Printf("  Issuer: %s\n", claims.Issuer)
		}
		
		if claims.Subject != "" {
			fmt.Printf("  Subject: %s\n", claims.Subject)
		}

		return nil
	},
}

func init() {
	opts := &verifyJwtTokenOptions{}

	verifyTokenCmd.Flags().StringVar(&opts.jwtSecretKey, "jwt-secret-key", "", "JWT secret key (overrides JWT_SECRET_KEY env var)")

	ctx := context.WithValue(context.Background(), KeyForOptions, opts)
	verifyTokenCmd.SetContext(ctx)

	jwtCmd.AddCommand(verifyTokenCmd)
}
