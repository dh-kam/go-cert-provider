package cmd

import (
	"fmt"

	"github.com/dh-kam/go-cert-provider/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the current version of the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-cert-provider v%s\n", config.Version)
		fmt.Printf("  Build Time: %s\n", config.BuildTime)
		fmt.Printf("  Git Commit: %s\n", config.GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
} 