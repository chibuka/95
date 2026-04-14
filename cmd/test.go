package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/ui/banner"
	"github.com/spf13/cobra"
)

var mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

func doTest(uuid string) error {
	fmt.Println(banner.Render())
	fmt.Println()
	fmt.Printf("  %s\n",mutedStyle.Render("95™"))
	fmt.Println()
	return testOnServer(uuid)
}

var testCmd = &cobra.Command{
	Use:   "test <stage-uuid>",
	Short: "Run tests locally without submitting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return doTest(args[0])
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
