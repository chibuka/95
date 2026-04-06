package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/client"
	"github.com/chibuka/95-cli/internal/config"
	"github.com/spf13/cobra"
)

func doLogout() error {
	var (
		green = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		muted = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		dim   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	)

	fmt.Println(muted.Render("  ◇ Signing out"))
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	if cfg.RefreshToken == "" {
		fmt.Println("  " + dim.Render("● Already logged out"))
		fmt.Println()
		return nil
	}

	apiURL := cfg.GetAPIURL()
	if err := client.Logout(cfg.RefreshToken, apiURL); err != nil {
		fmt.Println("  " + dim.Render("● Could not notify server — clearing local credentials..."))
	}

	if err := config.Clear(); err != nil {
		return fmt.Errorf("failed to clear credentials: %w", err)
	}

	fmt.Println("  " + green.Render("✓ Signed out successfully"))
	fmt.Println()
	return nil
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear local credentials and logout",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doLogout()
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
