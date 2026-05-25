package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/client"
	"github.com/spf13/cobra"
)

var loginHeadless bool

func doLogin(headless bool) error {
	var (
		orange = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
		green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		muted  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	)

	fmt.Println(muted.Render("  ◇ Authenticating"))
	fmt.Println()
	if headless {
		fmt.Println("  " + dim.Render("● Starting manual GitHub OAuth..."))
	} else {
		fmt.Println("  " + dim.Render("● Opening browser for GitHub OAuth..."))
	}
	fmt.Println()

	if err := client.Login(client.LoginOptions{Headless: headless}); err != nil {
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
	Use:     "login",
	Short:   "Authenticate with GitHub OAuth",
	Example: "  95 login --headless   Print the GitHub login URL instead of opening a browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doLogin(loginHeadless)
	},
}

func init() {
	loginCmd.Flags().BoolVar(&loginHeadless, "headless", false, "print the GitHub login URL instead of opening a browser")
	rootCmd.AddCommand(loginCmd)
}
