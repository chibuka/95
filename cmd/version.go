package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at link time in release builds (see .github/workflows/release.yml).
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the installed CLI build version",
	Run: func(c *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}
