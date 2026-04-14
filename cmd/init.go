package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chibuka/95-cli/client"
	"github.com/chibuka/95-cli/internal/config"
	"github.com/chibuka/95-cli/ui/picker"
	"github.com/spf13/cobra"
)

func doInit() error {
	runtimes, _ := client.FetchRuntimes()
	p := tea.NewProgram(picker.New(runtimes))
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("picker error: %w", err)
	}

	m, ok := result.(picker.Model)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	chosen := m.Chosen()
	if chosen == nil {
		return nil // user quit picker — back to dashboard
	}

	if err := config.SaveProjectConfig(chosen.Command, m.Language()); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n  ✓ initialized — %s (%s)\n\n", chosen.Command, m.Language())
	return nil
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project with run command",
	Long:  `Initialize your project by selecting how to run your Go code.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
