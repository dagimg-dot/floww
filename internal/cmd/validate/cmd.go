package validate

import (
	"errors"
	"os"

	"github.com/dagimg-dot/floww/internal/clihelper"
	"github.com/dagimg-dot/floww/internal/config"
	"github.com/dagimg-dot/floww/internal/diagnostic"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// useColorFunc decides whether diagnostics render with colors and source
// excerpts. Overridable in tests.
var useColorFunc = func() bool {
	return isatty.IsTerminal(os.Stderr.Fd()) && os.Getenv("NO_COLOR") == ""
}

var Command = &cobra.Command{
	Use:           "validate [name]",
	Short:         "Validate a workflow file",
	Long:          "Validate a workflow's schema without applying it.",
	Args:          cobra.MaximumNArgs(1),
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		filePath, _ := cobraCmd.Flags().GetString("file")

		cfg := config.NewConfigManager()

		// --file must NOT require init (matching apply --file behaviour).
		if filePath == "" {
			if err := clihelper.CheckInitialized(cfg); err != nil {
				return err
			}
		}

		var name, path string
		if filePath != "" {
			name, path = filePath, filePath
		} else {
			if len(args) > 0 {
				name = args[0]
			} else {
				name = clihelper.SelectWorkflow(cfg, "validate")
			}
			var err error
			path, err = cfg.ResolveWorkflowPath(name)
			if err != nil {
				return err
			}
		}

		cobraCmd.Printf("Validating workflow: %s\n", name)

		result, err := cfg.ValidateWorkflowFile(path)
		if err != nil {
			return err
		}

		if len(result.Diagnostics) == 0 {
			cobraCmd.Println("✓ Workflow is valid")
			return nil
		}

		diagnostic.Render(cobraCmd.ErrOrStderr(), path, result.Source, result.Diagnostics, useColorFunc())
		return errors.New("validation failed")
	},
}

func init() {
	Command.Flags().StringP("file", "f", "", "Path to the workflow file to validate")
}
