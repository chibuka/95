package cmd

import "github.com/spf13/cobra"

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Print help for 95 or a subcommand (usage, flags, and commands)",
	Args:  cobra.ArbitraryArgs,
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return rootCmd.Help()
		}
		cmd, _, err := rootCmd.Find(args)
		if err != nil {
			return err
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.SetHelpCommand(helpCmd)
}
