package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type retrieveCommandOptions struct {
	outputDir      string
	outputFormat   string
	separateFiles  bool
	certFileName   string
	keyFileName    string
	bundleFileName string
}

// retrieveCmd represents the retrieve command
var retrieveCmd = &cobra.Command{
	Use:   "retrieve <domain>",
	Short: "Retrieve SSL certificate for a domain",
	Long: `Retrieve SSL/TLS certificate for the specified domain from the configured provider.

The certificate can be output to stdout or saved to files. By default, the certificate
chain and private key are displayed to stdout.

Examples:
  # Retrieve certificate for example.com (output to stdout)
  go-cert-provider certs retrieve example.com

  # Save certificate to files in current directory
  go-cert-provider certs retrieve example.com --output-dir ./certs

  # Save as separate files
  go-cert-provider certs retrieve example.com \
    --output-dir ./certs \
    --separate-files

  # With Porkbun provider
  go-cert-provider certs retrieve example.com \
    --porkbun-api-key "your-key" \
    --porkbun-secret-key "your-secret" \
    --porkbun-domains "example.com,test.com"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]

		outputDir, _ := cmd.Flags().GetString("output-dir")
		separateFiles, _ := cmd.Flags().GetBool("separate-files")
		certFileName, _ := cmd.Flags().GetString("cert-file")
		keyFileName, _ := cmd.Flags().GetString("key-file")
		bundleFileName, _ := cmd.Flags().GetString("bundle-file")

		// Use global app state (initialized in PersistentPreRunE)
		if appState == nil {
			return fmt.Errorf("certificate system not initialized")
		}

		providerRegistry := appState.providerRegistry

		provider, err := providerRegistry.GetProviderForDomain(domain)
		if err != nil {
			return fmt.Errorf("no provider found for domain %s: %w", domain, err)
		}

		fmt.Fprintf(cmd.OutOrStderr(), "Retrieving certificate for %s from %s provider...\n", 
			domain, provider.GetProviderName())

		certChain, privateKey, err := provider.RetrieveCertificate(domain)
		if err != nil {
			return fmt.Errorf("failed to retrieve certificate: %w", err)
		}

		if outputDir == "" {
			return outputToStdout(cmd, certChain, privateKey, separateFiles)
		} else {
			return outputToFiles(cmd, domain, outputDir, certChain, privateKey, 
				separateFiles, certFileName, keyFileName, bundleFileName)
		}
	},
}

func outputToStdout(cmd *cobra.Command, certChain, privateKey []byte, separateFiles bool) error {
	if separateFiles {
		fmt.Fprintln(cmd.OutOrStdout(), "=== Certificate Chain ===")
		fmt.Fprintln(cmd.OutOrStdout(), string(certChain))
		fmt.Fprintln(cmd.OutOrStdout(), "\n=== Private Key ===")
		fmt.Fprintln(cmd.OutOrStdout(), string(privateKey))
	} else {
		fmt.Fprint(cmd.OutOrStdout(), string(certChain))
		fmt.Fprint(cmd.OutOrStdout(), string(privateKey))
	}
	return nil
}

func outputToFiles(cmd *cobra.Command, domain, outputDir string, certChain, privateKey []byte,
	separateFiles bool, certFileName, keyFileName, bundleFileName string) error {
	
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if separateFiles {
		if certFileName == "" {
			certFileName = fmt.Sprintf("%s.crt", domain)
		}
		if keyFileName == "" {
			keyFileName = fmt.Sprintf("%s.key", domain)
		}

		certPath := filepath.Join(outputDir, certFileName)
		keyPath := filepath.Join(outputDir, keyFileName)

		if err := os.WriteFile(certPath, certChain, 0644); err != nil {
			return fmt.Errorf("failed to write certificate file: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStderr(), "Certificate saved to: %s\n", certPath)

		if err := os.WriteFile(keyPath, privateKey, 0600); err != nil {
			return fmt.Errorf("failed to write private key file: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStderr(), "Private key saved to: %s\n", keyPath)

	} else {
		if bundleFileName == "" {
			bundleFileName = fmt.Sprintf("%s-bundle.pem", domain)
		}

		bundlePath := filepath.Join(outputDir, bundleFileName)
		bundle := append(certChain, privateKey...)

		if err := os.WriteFile(bundlePath, bundle, 0600); err != nil {
			return fmt.Errorf("failed to write bundle file: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStderr(), "Certificate bundle saved to: %s\n", bundlePath)
	}

	return nil
}

func init() {
	retrieveCmd.Flags().String("output-dir", "", "Directory to save certificate files (default: output to stdout)")
	retrieveCmd.Flags().Bool("separate-files", false, "Save certificate and key as separate files")
	retrieveCmd.Flags().String("cert-file", "", "Certificate file name (default: <domain>.crt)")
	retrieveCmd.Flags().String("key-file", "", "Private key file name (default: <domain>.key)")
	retrieveCmd.Flags().String("bundle-file", "", "Bundle file name (default: <domain>-bundle.pem)")

	certsCmd.AddCommand(retrieveCmd)
}
