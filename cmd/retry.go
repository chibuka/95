package cmd

import (
	"fmt"
	"strings"

	"github.com/chibuka/95-cli/internal/config"
	"github.com/spf13/cobra"
)

func doRetry() error {
	projectCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	action := strings.TrimSpace(projectCfg.LastAction)
	uuid := strings.TrimSpace(projectCfg.LastUUID)
	if action == "" || uuid == "" {
		return fmt.Errorf("no previous test/run command found in this project")
	}

	switch action {
	case "test":
		return doTest(uuid)
	case "run":
		return doRun(uuid)
	default:
		return fmt.Errorf("unsupported retry action '%s' in project config", action)
	}
}

var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Retry the last test/run command in this project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doRetry()
	},
}

func init() {
	rootCmd.AddCommand(retryCmd)
}
