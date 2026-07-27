package edit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dagimg-dot/floww/internal/clihelper"
	"github.com/dagimg-dot/floww/internal/config"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit an existing workflow",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		cfg := config.NewConfigManager()
		if err := clihelper.CheckInitialized(cfg); err != nil {
			return err
		}

		var workflowName string
		if len(args) > 0 {
			workflowName = args[0]
		} else {
			workflowName = clihelper.SelectWorkflow(cfg, "edit")
		}

		var workflowPath string
		for _, ext := range cfg.GetSupportedFormats() {
			candidate := filepath.Join(cfg.WorkflowsDir(), workflowName+ext)
			if _, err := os.Stat(candidate); err == nil {
				workflowPath = candidate
				break
			}
		}

		if workflowPath == "" {
			return fmt.Errorf("workflow '%s' not found", workflowName)
		}

		cobraCmd.Printf("Opening workflow '%s' in editor...\n", workflowName)
		clihelper.OpenInEditor(workflowPath)
		return nil
	},
}
