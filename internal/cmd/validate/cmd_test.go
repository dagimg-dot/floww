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
	useColorFunc = func() bool { return false }
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

func runCommand(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	resetCmd()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	Command.SetOut(out)
	Command.SetErr(errOut)
	Command.SetArgs(args)
	err = Command.Execute()
	return out.String(), errOut.String(), err
}

func TestValidate_ValidWorkflow(t *testing.T) {
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

	stdout, stderr, err := runCommand(t, "example")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Workflow is valid")
	assert.Empty(t, stderr)
}

func TestValidate_NonexistentWorkflow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupTest(t, dir)

	_, _, err := runCommand(t, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestValidate_InvalidWorkflow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)

	content := `description: "Invalid workflow"
workspaces:
  - target: 1
`
	err := os.WriteFile(filepath.Join(wfDir, "invalid.yaml"), []byte(content), 0600)
	require.NoError(t, err)

	stdout, stderr, err := runCommand(t, "invalid")
	require.Error(t, err)
	assert.Equal(t, "validation failed", err.Error())
	assert.Contains(t, stdout, "Validating workflow: invalid")
	assert.NotContains(t, stdout, "Workflow is valid")
	assert.Contains(t, stderr, ":3:5: error: Workspace definition")
	assert.Contains(t, stderr, "missing the required 'apps' key")
}

func TestValidate_SyntaxErrorShowsPosition(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)

	err := os.WriteFile(filepath.Join(wfDir, "bad.yaml"), []byte("workspaces:\n  - target: [1, 2\n"), 0600)
	require.NoError(t, err)

	_, stderr, err := runCommand(t, "bad")
	require.Error(t, err)
	assert.Contains(t, stderr, ":1: error: yaml: line 1:")
}

func TestValidate_MultipleErrorsReported(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)

	content := `workspaces:
  - target: 1
    apps:
      - name: ""
        exec: xterm
        type: invalid
`
	err := os.WriteFile(filepath.Join(wfDir, "multi.yaml"), []byte(content), 0600)
	require.NoError(t, err)

	_, stderr, err := runCommand(t, "multi")
	require.Error(t, err)
	assert.Contains(t, stderr, "missing the required 'name' key")
	assert.Contains(t, stderr, "must be one of 'binary', 'flatpak', 'snap'")
	assert.Contains(t, stderr, ":6:15: error:")
}

func TestValidate_FileFlagSkipsInit(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	// no floww config dir at all — --file must still work

	workflowPath := filepath.Join(dir, "outside.yaml")
	content := `workspaces:
  - target: 1
    apps:
      - name: term
        exec: xterm
`
	err := os.WriteFile(workflowPath, []byte(content), 0600)
	require.NoError(t, err)

	stdout, stderr, err := runCommand(t, "--file", workflowPath)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Workflow is valid")
	assert.Empty(t, stderr)
}

func TestValidate_FileFlagInvalid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupTest(t, dir)

	workflowPath := filepath.Join(dir, "bad.json")
	err := os.WriteFile(workflowPath, []byte(`{"workspaces": [{"target": "abc"}]}`), 0600)
	require.NoError(t, err)

	_, stderr, err := runCommand(t, "--file", workflowPath)
	require.Error(t, err)
	assert.Contains(t, stderr, "bad.json:1:28: error: cannot unmarshal string into int")
}

func TestValidate_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupTest(t, dir)

	_, _, err := runCommand(t, "--file", filepath.Join(dir, "ghost.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
