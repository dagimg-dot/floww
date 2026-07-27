package remove

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/dagimg-dot/floww/internal/config"
)

var Command = &cobra.Command{
	Use:   "remove [names...]",
	Short: "Remove a workflow",
	Long:  "Remove one or more workflow files.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		force, _ := cobraCmd.Flags().GetBool("force")
		cfg := config.NewConfigManager()
		if err := checkInitialized(cfg); err != nil {
			return err
		}

		names := args
		if len(names) == 0 {
			selected := selectWorkflow(cfg, "remove")
			names = []string{selected}
		}

		type fileEntry struct {
			name string
			path string
		}
		var allFiles []fileEntry
		for _, workflowName := range names {
			found := findWorkflowFiles(cfg, workflowName)
			if len(found) == 0 {
				return fmt.Errorf("workflow '%s' not found", workflowName)
			}
			for _, f := range found {
				allFiles = append(allFiles, fileEntry{name: workflowName, path: f})
			}
		}

		if !force {
			var prompt string
			if len(names) > 1 {
				prompt = fmt.Sprintf("Are you sure you want to remove workflows %s?", joinQuoted(names))
			} else {
				prompt = fmt.Sprintf("Are you sure you want to remove workflow '%s'?", names[0])
			}

			confirmed := false
			err := huh.NewConfirm().
				Title(prompt).
				Value(&confirmed).
				Run()
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				cobraCmd.Println("Operation cancelled")
				return nil
			}
		}

		for _, entry := range allFiles {
			if err := os.Remove(entry.path); err != nil {
				return fmt.Errorf("failed to remove workflow file: %w", err)
			}
			cobraCmd.Printf("Removed workflow: %s\n", filepath.Base(entry.path))
		}
		return nil
	},
}

func init() {
	Command.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func checkInitialized(cfg *config.ConfigManager) error {
	if !cfg.IsInitialized() {
		return fmt.Errorf("floww is not initialized. Please run 'floww init' first")
	}
	return nil
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "\033[1;31mError:\033[0m %s\n", msg)
}

func selectWorkflow(cfg *config.ConfigManager, action string) string {
	available := cfg.ListWorkflowNames()
	if len(available) == 0 {
		printError(fmt.Sprintf("No workflows found to %s", action))
		os.Exit(1)
	}

	var selected string
	err := huh.NewSelect[string]().
		Title(fmt.Sprintf("Select a workflow to %s:", action)).
		Options(huh.NewOptions(available...)...).
		Value(&selected).
		Run()

	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Println("No workflow selected")
		os.Exit(0)
	}
	if err != nil {
		printError(fmt.Sprintf("Selection failed: %v", err))
		os.Exit(1)
	}

	return selected
}

func findWorkflowFiles(cfg *config.ConfigManager, name string) []string {
	var found []string
	for _, ext := range cfg.GetSupportedFormats() {
		path := filepath.Join(cfg.WorkflowsDir(), name+ext)
		if _, err := os.Stat(path); err == nil {
			found = append(found, path)
		}
	}
	return found
}

func joinQuoted(names []string) string {
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = "'" + n + "'"
	}
	return strings.Join(quoted, ", ")
}
