package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dh-kam/go-cert-provider/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
)

type createJwtTokenOptions struct {
	userID         string
	description    string
	allowedDomains string
	expiresAt      string
	jwtSecretKey   string
}

var createTokenCmd = &cobra.Command{
	Use:   "create-token",
	Short: "Create a new JWT token",
	Long:  "Create a new JWT token with specified user ID, description, and allowed domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		options, ok := cmd.Context().Value(KeyForOptions).(*createJwtTokenOptions)
		if !ok {
			return fmt.Errorf("failed to get command options from context")
		}

		if options.userID == "" {
			return fmt.Errorf("user-id is required")
		}

		if options.allowedDomains == "" {
			return fmt.Errorf("allowed-domains is required")
		}

		allowedDomainsList := strings.Split(options.allowedDomains, ",")
		for i, domain := range allowedDomainsList {
			allowedDomainsList[i] = strings.TrimSpace(domain)
		}

		jwtSecretKey := options.jwtSecretKey
		if jwtSecretKey == "" {
			jwtSecretKey = os.Getenv("JWT_SECRET_KEY")
		}
		if jwtSecretKey == "" {
			return fmt.Errorf("jwt secret key is required; use --jwt-secret-key flag or set JWT_SECRET_KEY environment variable")
		}

		var expiresAt time.Time
		if options.expiresAt != "" {
			var err error

			// Try parsing as duration first (e.g., "2y", "3months", "5d")
			if duration, durationErr := utils.ParseDurationString(options.expiresAt); durationErr == nil {
				expiresAt = time.Now().Add(duration)
			} else {
				// Try parsing as date/time formats
				formats := []string{
					utils.DateTimeFormat,
					time.RFC3339,
					"2006-01-02T15:04:05",
					"2006-01-02",
				}

				for _, format := range formats {
					switch format {
					case utils.DateTimeFormat:
						expiresAt, err = utils.ParseDateTime(options.expiresAt)
					case "2006-01-02":
						// For date-only format, set time to 23:59:59
						dateOnly, parseErr := time.ParseInLocation(format, options.expiresAt, time.Local)
						if parseErr == nil {
							expiresAt = time.Date(dateOnly.Year(), dateOnly.Month(), dateOnly.Day(), 23, 59, 59, 0, time.Local)
							err = nil
						} else {
							err = parseErr
						}
					default:
						expiresAt, err = time.ParseInLocation(format, options.expiresAt, time.Local)
					}
					if err == nil {
						break
					}
				}

				if err != nil {
					return fmt.Errorf("invalid expires-at format; use duration (e.g., '2y', '3months', '5d') or date/time format (YYYY-MM-DD HH:mm:ss, YYYY-MM-DD)")
				}
			}
		} else {
			expiresAt = time.Now().Add(365 * 24 * time.Hour)
		}

		issuedAt := time.Now()

		claims := jwt.MapClaims{
			"user_id":         options.userID,
			"description":     options.description,
			"allowed_domains": allowedDomainsList,
			"exp":             expiresAt.Unix(),
			"iat":             issuedAt.Unix(),
			"nbf":             issuedAt.Unix(),
			"iss":             "go-cert-provider",
			"sub":             options.userID,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtSecretKey))
		if err != nil {
			return fmt.Errorf("failed to create JWT token: %w", err)
		}

		greenStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

		fmt.Printf("JWT Token created successfully:\n\n")
		fmt.Printf("Token:\n")
		fmt.Println(greenStyle.Render(tokenString))
		fmt.Printf("\nClaims:\n")
		fmt.Printf("  User ID: %s\n", options.userID)
		fmt.Printf("  Description: %s\n", options.description)
		fmt.Printf("  Allowed Domains: %s\n", strings.Join(allowedDomainsList, ", "))
		fmt.Printf("  Expires At: %s\n", utils.FormatDateTime(expiresAt))
		fmt.Printf("  Issued At: %s\n", utils.FormatDateTime(issuedAt))

		return nil
	},
}

func init() {
	opts := &createJwtTokenOptions{}

	flags := createTokenCmd.Flags()
	flags.StringVar(&opts.userID, "user-id", "", "User ID (required)")
	flags.StringVar(&opts.description, "description", "", "Token description")
	flags.StringVar(&opts.allowedDomains, "allowed-domains", "", "Comma-separated list of allowed domains (required)")
	flags.StringVar(&opts.expiresAt, "expires-at", "", "Token expiration time: duration (2y, 3months, 5d) or date (YYYY-MM-DD HH:mm:ss, YYYY-MM-DD) (default: 1 year)")
	flags.StringVar(&opts.jwtSecretKey, "jwt-secret-key", "", "JWT secret key (overrides JWT_SECRET_KEY env var)")

	if err := createTokenCmd.MarkFlagRequired("user-id"); err != nil {
		panic(err)
	}
	if err := createTokenCmd.MarkFlagRequired("allowed-domains"); err != nil {
		panic(err)
	}

	ctx := context.WithValue(context.Background(), KeyForOptions, opts)
	createTokenCmd.SetContext(ctx)

	jwtCmd.AddCommand(createTokenCmd)
}
