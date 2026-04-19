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

	if err := config.SaveProjectConfig(chosen.Build, chosen.Run, m.Language()); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Show the build step too when there is one, so the user sees exactly what
	// the server will execute — easier to self-diagnose "why did my submission fail".
	if chosen.Build != "" {
		fmt.Printf("\n  ✓ initialized — %s (%s)\n       build: %s\n", chosen.Run, m.Language(), chosen.Build)
	} else {
		fmt.Printf("\n  ✓ initialized — %s (%s)\n", chosen.Run, m.Language())
	}
	// `make` requires a specific Makefile shape (./app output, tabs, no network
	// at build time). Point users at the docs the moment they commit so they
	// don't learn the rules via a failed submission.
	if chosen.IsMakePreset() {
		fmt.Printf("       Makefile setup guide: %s\n", picker.MakefileDocsURL)
	}
	fmt.Println()
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
