package cmd

import (
	"bufio"
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
		return runInteractive("")
	},
}

// runInteractive runs the dashboard loop. prevErrMsg, if non-empty, causes
// the dashboard to open in stateResult so the user sees the error first.
func runInteractive(prevErrMsg string) error {
	for {
		cfg, _ := config.Load()
		username := "not authenticated"
		if cfg != nil && cfg.Username != "" {
			username = cfg.Username
		}

		dir, _ := os.Getwd()
		if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(dir, home) {
			dir = "~" + dir[len(home):]
		}

		fmt.Print("\033[2J\033[H")

		p := tea.NewProgram(dashboard.New(username, dir, prevErrMsg))
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
			prevErrMsg = runErr.Error()
		} else {
			prevErrMsg = ""
			if !askContinue() {
				return nil
			}
		}
	}
}

var (
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	keyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
)

// askContinue prints a continue/quit prompt and returns true if the user
// wants to run another command.
func askContinue() bool {
	fmt.Printf("\n  %s  %s   %s  %s\n  > ",
		keyStyle.Render("enter"),
		promptStyle.Render("another command"),
		keyStyle.Render("q"),
		promptStyle.Render("quit"),
	)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(line) != "q"
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
