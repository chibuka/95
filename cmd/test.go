package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/internal/config"
	"github.com/chibuka/95-cli/ui/banner"
	"github.com/spf13/cobra"
)

var mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

func doTest(uuid string) error {
	if err := config.SaveLastCommand("test", uuid); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save retry command: %v\n", err)
	}

	fmt.Println(banner.Render())
	fmt.Println()
	fmt.Printf("  %s\n", mutedStyle.Render("95™"))
	fmt.Println()
	return testOnServer(uuid)
}

var testCmd = &cobra.Command{
	Use:   "test <floor-uuid>",
	Short: "Run tests locally without submitting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return doTest(args[0])
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
