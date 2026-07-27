package validate

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

func setupTest(t *testing.T, dir string) string {
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

func TestValidate_ValidWorkflow(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)

	content := `description: "Example workflow"
workspaces:
  - target: 1
    apps:
      - name: Terminal
        exec: gnome-terminal
`
	err := os.WriteFile(filepath.Join(wfDir, "example.yaml"), []byte(content), 0600)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"example"})

	err = Command.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Workflow is valid")
}

func TestValidate_NonexistentWorkflow(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupTest(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"nonexistent"})

	err := Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestValidate_InvalidWorkflow(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)

	content := `description: "Invalid workflow"
workspaces:
  - target: 1
`
	err := os.WriteFile(filepath.Join(wfDir, "invalid.yaml"), []byte(content), 0600)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"invalid"})

	err = Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}
