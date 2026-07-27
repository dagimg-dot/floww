package list

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetCmd() {
	Command.SetArgs([]string{})
	Command.SetOut(nil)
	Command.SetErr(nil)
	Command.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

func setupInit(t *testing.T, dir string) string {
	t.Helper()
	flowwDir := filepath.Join(dir, "floww")
	workflowsDir := filepath.Join(flowwDir, "workflows")
	configFile := filepath.Join(flowwDir, "config.yaml")

	err := os.MkdirAll(workflowsDir, 0750)
	require.NoError(t, err)

	err = os.WriteFile(configFile, []byte("general: {}\ntiming: {}\n"), 0600)
	require.NoError(t, err)

	return workflowsDir
}

func createWorkflowFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	content := `description: "test"
workspaces:
  - target: 0
    apps:
      - name: Terminal
        exec: gnome-terminal
`
	err := os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)
}

func TestList_NoWorkflows(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupInit(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{})

	err := Command.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No workflows found")
}

func TestList_WithWorkflows(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupInit(t, dir)

	createWorkflowFile(t, wfDir, "alpha.yaml")
	createWorkflowFile(t, wfDir, "beta.yaml")

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{})

	err := Command.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Available workflows:")
	assert.Contains(t, output, "- alpha")
	assert.Contains(t, output, "- beta")
}
