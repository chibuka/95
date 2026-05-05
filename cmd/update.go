package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update 95 CLI to the latest release",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func doUpdate() error {
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	success := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	fmt.Println(info.Render("  → checking for updates..."))

	relTag, err := updater.FetchLatestTag()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if relTag == Version {
		fmt.Println(success.Render(fmt.Sprintf("  ✓ already on latest version (%s)", Version)))
		return nil
	}

	if Version == "dev" {
		fmt.Println(warn.Render("  ! running a dev build, updating to " + relTag))
	} else {
		fmt.Println(info.Render(fmt.Sprintf("  → updating %s → %s", Version, relTag)))
	}

	asset := fmt.Sprintf("95-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		asset += ".exe"
	}
	url := fmt.Sprintf("https://github.com/chibuka/95-cli/releases/download/%s/%s", relTag, asset)

	dlResp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != 200 {
		return fmt.Errorf("failed to download binary: status %d (url: %s)", dlResp.StatusCode, url)
	}

	tmp, err := os.CreateTemp("", "95-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, dlResp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write binary: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to locate current binary: %w", err)
	}

	if runtime.GOOS == "windows" {
		old := exe + ".old"
		os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to move old binary: %w", err)
		}
	}

	if err := os.Rename(tmpPath, exe); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	updater.RecordAppliedVersion(relTag)
	fmt.Println(success.Render(fmt.Sprintf("  ✓ updated to %s", relTag)))
	return nil
}
