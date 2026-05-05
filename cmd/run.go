package cmd

import (
	"fmt"
	"os"

	"github.com/chibuka/95-cli/internal/config"
	"github.com/chibuka/95-cli/ui/banner"
	"github.com/spf13/cobra"
)

func doRun(uuid string) error {
	if err := config.SaveLastCommand("run", uuid); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save retry command: %v\n", err)
	}

	fmt.Println(banner.Render())
	fmt.Println()
	fmt.Printf("  %s\n", mutedStyle.Render("95™"))
	fmt.Println()
	return submitToServer(uuid)
}

var runCmd = &cobra.Command{
	Use:   "run <floor-uuid>",
	Short: "Submit your code to the server for testing and validation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return doRun(args[0])
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
