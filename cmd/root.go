package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/internal/config"
	"github.com/chibuka/95-cli/ui/dashboard"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "95",
	Short:         "Practice coding challenges and level up your skills",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))

		for {
			// Reload config each iteration so username reflects login/logout changes.
			cfg, _ := config.Load()
			username := "not authenticated"
			if cfg != nil && cfg.Username != "" {
				username = cfg.Username
			}

			dir, _ := os.Getwd()
			if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(dir, home) {
				dir = "~" + dir[len(home):]
			}

			p := tea.NewProgram(dashboard.New(username, dir))
			result, err := p.Run()
			if err != nil {
				return err
			}

			m := result.(dashboard.Model)
			action := m.Result()
			if action == nil || action.Quit {
				return nil
			}

			fmt.Println()

			var runErr error
			switch action.Command {
			case "login":
				runErr = doLogin()
			case "init":
				runErr = doInit()
			case "test":
				runErr = doTest(action.UUID)
			case "run":
				runErr = doRun(action.UUID)
			case "logout":
				runErr = doLogout()
			}

			if runErr != nil {
				fmt.Println(errStyle.Render("  ✗ " + runErr.Error()))
				fmt.Println()
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).Render("  ✗ "+err.Error()))
		os.Exit(1)
	}
}

func init() {
	config.Init()
}
