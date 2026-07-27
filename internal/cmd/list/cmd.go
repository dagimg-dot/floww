package list

import (
	"github.com/dagimg-dot/floww/internal/clihelper"
	"github.com/dagimg-dot/floww/internal/config"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "list",
	Short: "List available workflows",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg := config.NewConfigManager()
		if err := clihelper.CheckInitialized(cfg); err != nil {
			return err
		}

		names := cfg.ListWorkflowNames()
		if len(names) == 0 {
			cmd.Println("No workflows found")
			return nil
		}

		cmd.Println("Available workflows:")
		for _, name := range names {
			cmd.Printf("  - %s\n", name)
		}
		return nil
	},
}
