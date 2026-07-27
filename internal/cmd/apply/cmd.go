package apply

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/dagimg-dot/floww/internal/clihelper"
	"github.com/dagimg-dot/floww/internal/config"
	"github.com/dagimg-dot/floww/internal/launcher"
	"github.com/dagimg-dot/floww/internal/workflow"
	"github.com/dagimg-dot/floww/internal/workspace"
	"github.com/spf13/cobra"
)

// workflowApplier is the interface satisfied by *workflow.WorkflowManager.
type workflowApplier interface {
	Apply(data *workflow.Workflow, append bool) bool
}

// wfManagerFactory is overridable in tests to inject a mock WorkflowManager.
var wfManagerFactory func(*config.ConfigManager, workflow.WorkspaceManager, workflow.AppLauncher) workflowApplier = defaultWFManagerFactory

func defaultWFManagerFactory(cfg *config.ConfigManager, wsMgr workflow.WorkspaceManager, appLauncher workflow.AppLauncher) workflowApplier {
	return workflow.NewWorkflowManager(wsMgr, appLauncher, cfg)
}

var Command = &cobra.Command{
	Use:   "apply [name]",
	Short: "Apply a workflow",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		filePath, _ := cobraCmd.Flags().GetString("file")
		appendMode, _ := cobraCmd.Flags().GetBool("append")

		cfg := config.NewConfigManager()

		// --file must NOT require init (matching Python's apply.py behaviour).
		if filePath == "" {
			if err := clihelper.CheckInitialized(cfg); err != nil {
				return err
			}
		}

		var workflowName string
		var workflowData *workflow.Workflow
		var err error

		if filePath != "" {
			workflowName = filePath
			slog.Info("Loading workflow from file", "path", filePath)
			workflowData, err = cfg.LoadWorkflow(workflowName, true)
		} else {
			if len(args) > 0 {
				workflowName = args[0]
			} else {
				workflowName = clihelper.SelectWorkflow(cfg, "apply")
			}
			slog.Info("Loading workflow", "name", workflowName)
			workflowData, err = cfg.LoadWorkflow(workflowName, false)
		}

		if err != nil {
			displayName := workflowName
			if displayName == "" {
				displayName = "selected"
			}

			var wnfe *config.WorkflowNotFoundError
			var cle *config.ConfigLoadError
			var ce *config.ConfigError
			if errors.As(err, &wnfe) || errors.As(err, &cle) || errors.As(err, &ce) {
				return fmt.Errorf("failed to load workflow '%s': %w", displayName, err)
			}
			return fmt.Errorf("failed to load workflow '%s': %w", displayName, err)
		}

		slog.Info("Applying workflow", "name", workflowName)

		backend := workspace.CreateBackend("", cfg)
		wsMgr := workspace.NewWorkspaceManager(backend)
		appLauncher := launcher.New()
		wfMgr := wfManagerFactory(cfg, wsMgr, appLauncher)

		if !wfMgr.Apply(workflowData, appendMode) {
			return fmt.Errorf("workflow application failed")
		}
		return nil
	},
}

func init() {
	Command.Flags().StringP("file", "f", "", "Path to the workflow file to apply")
	Command.Flags().BoolP("append", "a", false, "Append the workflow starting from the last workspace")
}
