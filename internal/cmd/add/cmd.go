package add

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dagimg-dot/floww/internal/clihelper"
	"github.com/dagimg-dot/floww/internal/config"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		cfg := config.NewConfigManager()
		if err := clihelper.CheckInitialized(cfg); err != nil {
			return err
		}

		name := args[0]

		if strings.Contains(name, "/") || strings.Contains(name, "\\") {
			return fmt.Errorf("workflow name cannot contain path separators")
		}

		if strings.HasPrefix(name, ".") {
			return fmt.Errorf("workflow name cannot start with a dot")
		}

		if strings.Contains(name, ".") {
			return fmt.Errorf("please provide the name without file extension")
		}

		fileType, _ := cobraCmd.Flags().GetString("type")
		ext := "." + fileType

		loader := config.NewConfigLoader()
		if !loader.IsSupportedFormat("x" + ext) {
			return fmt.Errorf("unsupported format: %s", fileType)
		}

		var existingExts []string
		for _, supportedExt := range cfg.GetSupportedFormats() {
			candidate := filepath.Join(cfg.WorkflowsDir(), name+supportedExt)
			if _, err := os.Stat(candidate); err == nil {
				existingExts = append(existingExts, supportedExt)
			}
		}
		if len(existingExts) > 0 {
			extList := strings.Join(existingExts, ", ")
			return fmt.Errorf("workflow '%s' already exists with extension: %s", name, extList)
		}

		workflowData := map[string]any{
			"description": "",
			"workspaces": []any{
				map[string]any{
					"target": 0,
					"apps": []any{
						map[string]any{
							"name": "App Name",
							"exec": "command",
							"type": "binary",
						},
					},
				},
			},
		}

		filePath := filepath.Join(cfg.WorkflowsDir(), name+ext)
		if err := loader.Save(workflowData, filePath); err != nil {
			return fmt.Errorf("failed to create workflow: %w", err)
		}

		cobraCmd.Printf("Created workflow: %s\n", filePath)

		edit, _ := cobraCmd.Flags().GetBool("edit")
		if edit {
			clihelper.OpenInEditor(filePath)
		}
		return nil
	},
}

func init() {
	Command.Flags().BoolP("edit", "e", false, "Open the new workflow in your editor after creation")
	Command.Flags().StringP("type", "t", "yaml", "Workflow file type (yaml, json, toml)")
}
