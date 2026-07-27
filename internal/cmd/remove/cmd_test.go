package remove

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

func createWorkflow(t *testing.T, wfDir, name string) string {
	t.Helper()
	path := filepath.Join(wfDir, name)
	content := `description: "test"
workspaces:
  - target: 0
    apps:
      - name: Terminal
        exec: gnome-terminal
`
	err := os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)
	return path
}

func TestRemove_SingleWorkflow(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)
	filePath := createWorkflow(t, wfDir, "demo.yaml")

	assert.FileExists(t, filePath)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"demo", "--force"})

	err := Command.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Removed workflow: demo.yaml")
	assert.NoFileExists(t, filePath)
}

func TestRemove_MultipleWorkflows(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)

	file1 := createWorkflow(t, wfDir, "foo.yaml")
	file2 := createWorkflow(t, wfDir, "bar.yaml")

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"foo", "bar", "--force"})

	err := Command.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Removed workflow: foo.yaml")
	assert.Contains(t, output, "Removed workflow: bar.yaml")
	assert.NoFileExists(t, file1)
	assert.NoFileExists(t, file2)
}

func TestRemove_NonexistentWorkflow(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupTest(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"nope", "--force"})

	err := Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
