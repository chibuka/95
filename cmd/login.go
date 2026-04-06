package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/client"
	"github.com/spf13/cobra"
)

func doLogin() error {
	var (
		orange = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
		green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		muted  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	)

	fmt.Println(muted.Render("  ◇ Authenticating"))
	fmt.Println()
	fmt.Println("  " + dim.Render("● Opening browser for GitHub OAuth..."))
	fmt.Println()

	if err := client.Login(); err != nil {
		fmt.Println("  " + orange.Render("✗ "+err.Error()))
		fmt.Println()
		return err
	}

	fmt.Println("  " + green.Render("✓ Logged in successfully!"))
	fmt.Println(muted.Render("    Build your coding skills, one challenge at a time."))
	fmt.Println()
	return nil
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub OAuth",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doLogin()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
