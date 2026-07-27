package validate

import (
	"fmt"

	"github.com/dagimg-dot/floww/internal/clihelper"
	"github.com/dagimg-dot/floww/internal/config"
	"github.com/dagimg-dot/floww/internal/workflow"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "validate [name]",
	Short: "Validate a workflow file",
	Long:  "Validate a workflow's schema without applying it.",
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
			workflowName = clihelper.SelectWorkflow(cfg, "validate")
		}

		cobraCmd.Printf("Validating workflow: %s\n", workflowName)

		wf, err := cfg.LoadWorkflow(workflowName, false)
		if err != nil {
			return fmt.Errorf("validation failed: %s", err.Error())
		}

		if err := workflow.ValidateWorkflow(workflowName, wf); err != nil {
			return fmt.Errorf("validation failed: %s", err.Error())
		}

		cobraCmd.Println("✓ Workflow is valid")
		return nil
	},
}
