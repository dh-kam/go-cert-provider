package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dh-kam/go-cert-provider/cert/domain"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed domains with details",
	Long: `List all domains managed by configured certificate providers.

This command displays all domains that are available for certificate retrieval
from the configured providers, including status and expiration information.

Examples:
  # List all domains (simple)
  go-cert-provider domain list

  # List with detailed information (provider, status, dates)
  go-cert-provider domain list --detail

  # Output as JSON with details
  go-cert-provider domain list --output json --detail

  # With Porkbun provider (auto-discovery)
  go-cert-provider domain list \
    --porkbun-api-key "your-key" \
    --porkbun-secret-key "your-secret"

  # With manually specified domains
  go-cert-provider domain list \
    --porkbun-api-key "your-key" \
    --porkbun-secret-key "your-secret" \
    --porkbun-domains "example.com,test.com"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get options from command flags
		outputFormat, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
		showDetail, err := cmd.Flags().GetBool("detail")
		if err != nil {
			return err
		}

		// Use global app state (initialized in PersistentPreRunE)
		if appState == nil {
			return fmt.Errorf("certificate system not initialized")
		}

		providerRegistry := appState.providerRegistry

		domains := providerRegistry.ListDomains()

		if len(domains) == 0 {
			fmt.Fprintln(cmd.OutOrStderr(), "No domains found")
			return nil
		}

		sort.Strings(domains)

		switch outputFormat {
		case "json":
			return outputJSON(cmd, domains, providerRegistry, showDetail)
		case "table", "":
			return outputTable(cmd, domains, providerRegistry, showDetail)
		case "simple":
			return outputSimple(cmd, domains)
		default:
			return fmt.Errorf("unsupported output format: %s", outputFormat)
		}
	},
}

func outputSimple(cmd *cobra.Command, domains []string) error {
	for _, domain := range domains {
		fmt.Fprintln(cmd.OutOrStdout(), domain)
	}
	return nil
}

func outputTable(cmd *cobra.Command, domains []string, registry interface{}, showDetail bool) error {
	providerRegistry := appState.providerRegistry

	allDomainInfo := providerRegistry.ListAllDomainInfo()

	infoMap := make(map[string]*domain.Info)
	for i := range allDomainInfo {
		infoMap[allDomainInfo[i].Name] = &allDomainInfo[i]
	}

	if showDetail {
		maxDomainLen := 6 // "DOMAIN"
		for _, d := range domains {
			if len(d) > maxDomainLen {
				maxDomainLen = len(d)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %-8s  %-10s  %-19s  %-19s\n",
			maxDomainLen, "DOMAIN", "PROVIDER", "STATUS", "CREATED", "EXPIRES")
		fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s  %s  %s\n",
			strings.Repeat("-", maxDomainLen),
			strings.Repeat("-", 8),
			strings.Repeat("-", 10),
			strings.Repeat("-", 19),
			strings.Repeat("-", 19))

		for _, domainName := range domains {
			info := infoMap[domainName]
			if info != nil {
				created := formatDate(info.CreateDate)
				expires := formatDate(info.ExpireDate)
				fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %-8s  %-10s  %-19s  %-19s\n",
					maxDomainLen, domainName, info.Provider, info.Status, created, expires)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %-8s  %-10s  %-19s  %-19s\n",
					maxDomainLen, domainName, "unknown", "UNKNOWN", "-", "-")
			}
		}
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "DOMAINS")
		fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 40))
		for _, domainName := range domains {
			fmt.Fprintln(cmd.OutOrStdout(), domainName)
		}
	}

	fmt.Fprintf(cmd.OutOrStderr(), "\nTotal: %d domain(s)\n", len(domains))
	return nil
}

// formatDate formats a time.Time for display
func formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

func outputJSON(cmd *cobra.Command, domains []string, registry interface{}, showDetail bool) error {
	providerRegistry := appState.providerRegistry

	// Get all domain info
	allDomainInfo := providerRegistry.ListAllDomainInfo()

	// Create a map for quick lookup
	infoMap := make(map[string]*domain.Info)
	for i := range allDomainInfo {
		infoMap[allDomainInfo[i].Name] = &allDomainInfo[i]
	}

	if showDetail {
		// Build domain-provider map with full info
		type domainInfoJSON struct {
			Domain     string `json:"domain"`
			Provider   string `json:"provider"`
			Status     string `json:"status"`
			CreateDate string `json:"createDate,omitempty"`
			ExpireDate string `json:"expireDate,omitempty"`
		}

		var domainInfos []domainInfoJSON
		for _, domainName := range domains {
			info := infoMap[domainName]
			if info != nil {
				created := ""
				expires := ""
				if !info.CreateDate.IsZero() {
					created = info.CreateDate.Format(time.RFC3339)
				}
				if !info.ExpireDate.IsZero() {
					expires = info.ExpireDate.Format(time.RFC3339)
				}

				domainInfos = append(domainInfos, domainInfoJSON{
					Domain:     domainName,
					Provider:   info.Provider,
					Status:     info.Status,
					CreateDate: created,
					ExpireDate: expires,
				})
			} else {
				domainInfos = append(domainInfos, domainInfoJSON{
					Domain:   domainName,
					Provider: "unknown",
					Status:   "UNKNOWN",
				})
			}
		}

		fmt.Fprintln(cmd.OutOrStdout(), "{")
		fmt.Fprintf(cmd.OutOrStdout(), "  \"total\": %d,\n", len(domains))
		fmt.Fprintln(cmd.OutOrStdout(), "  \"domains\": [")
		for i, info := range domainInfos {
			comma := ","
			if i == len(domainInfos)-1 {
				comma = ""
			}
			fmt.Fprintf(cmd.OutOrStdout(), "    {\"domain\": \"%s\", \"provider\": \"%s\", \"status\": \"%s\"",
				info.Domain, info.Provider, info.Status)
			if info.CreateDate != "" {
				fmt.Fprintf(cmd.OutOrStdout(), ", \"createDate\": \"%s\"", info.CreateDate)
			}
			if info.ExpireDate != "" {
				fmt.Fprintf(cmd.OutOrStdout(), ", \"expireDate\": \"%s\"", info.ExpireDate)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "}%s\n", comma)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "  ]")
		fmt.Fprintln(cmd.OutOrStdout(), "}")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "{")
		fmt.Fprintf(cmd.OutOrStdout(), "  \"total\": %d,\n", len(domains))
		fmt.Fprintln(cmd.OutOrStdout(), "  \"domains\": [")
		for i, domain := range domains {
			comma := ","
			if i == len(domains)-1 {
				comma = ""
			}
			fmt.Fprintf(cmd.OutOrStdout(), "    \"%s\"%s\n", domain, comma)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "  ]")
		fmt.Fprintln(cmd.OutOrStdout(), "}")
	}

	return nil
}

func init() {
	listCmd.Flags().String("output", "table", "Output format (table, simple, json)")
	listCmd.Flags().Bool("detail", false, "Show detailed information (provider, status, dates)")

	domainCmd.AddCommand(listCmd)
}
