package clihelper

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/dagimg-dot/floww/internal/config"
)

// PrintError writes a styled error message to stderr.
func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "\033[1;31mError:\033[0m %s\n", msg)
}

// CheckInitialized returns an error when floww has not been initialised yet.
// The caller should handle the error appropriately (e.g. return it from RunE).
func CheckInitialized(cfg *config.ConfigManager) error {
	if !cfg.IsInitialized() {
		return fmt.Errorf("floww is not initialized. Please run 'floww init' first")
	}
	return nil
}

// SelectWorkflow presents an interactive huh select list of available workflow
// names.  It returns the selected name, or exits when cancelled / empty.
func SelectWorkflow(cfg *config.ConfigManager, action string) string {
	available := cfg.ListWorkflowNames()
	if len(available) == 0 {
		PrintError(fmt.Sprintf("No workflows found to %s", action))
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
		PrintError(fmt.Sprintf("Selection failed: %v", err))
		os.Exit(1)
	}

	return selected
}

// OpenInEditor opens the given file path in $EDITOR, falling back to vim → vi
// → nano.  It exits with code 1 when no editor can be found or the editor
// returns a non-zero exit code.
func OpenInEditor(filePath string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		for _, ed := range []string{"vim", "vi", "nano"} {
			if _, err := exec.LookPath(ed); err == nil {
				editor = ed
				break
			}
		}
	}

	if editor == "" {
		PrintError("No suitable editor found. Please set the EDITOR environment variable.")
		os.Exit(1)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	edCmd := exec.Command(editor, absPath) //nolint:gosec // Intentional editor open
	edCmd.Stdin = os.Stdin
	edCmd.Stdout = os.Stdout
	edCmd.Stderr = os.Stderr

	if err := edCmd.Run(); err != nil {
		PrintError(fmt.Sprintf("Editor exited with error: %v", err))
		os.Exit(1)
	}
}
