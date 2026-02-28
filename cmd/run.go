package cmd

import (
	"fmt"

	"github.com/chibuka/95-cli/ui/banner"
	"github.com/spf13/cobra"
)

func doRun(uuid string) error {
	fmt.Println(banner.Render())
	fmt.Println()
	fmt.Println(fmt.Sprintf("  %s                              %s\n",
		mutedStyle.Render("95™"),
		mutedStyle.Render("v"+banner.Version)))
	fmt.Println()
	return submitToServer(uuid)
}

var runCmd = &cobra.Command{
	Use:   "run <stage-uuid>",
	Short: "Submit your code to the server for testing and validation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return doRun(args[0])
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
