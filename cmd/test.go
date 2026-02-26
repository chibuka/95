package cmd

import (
	"fmt"

	"github.com/chibuka/95-cli/ui/banner"
	"github.com/spf13/cobra"
)

func doTest(uuid string) error {
	fmt.Println()
	fmt.Println(banner.Render())
	fmt.Println()
	return testOnServer(uuid)
}

var testCmd = &cobra.Command{
	Use:   "test <stage-uuid>",
	Short: "Run tests locally without submitting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return doTest(args[0])
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
