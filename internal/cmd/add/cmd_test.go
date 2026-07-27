package add

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

func setupCmdTest(t *testing.T, dir string) string {
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

func TestAdd_NewWorkflow(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupCmdTest(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"myflow"})

	err := Command.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Created workflow:")

	workflowFile := filepath.Join(dir, "floww", "workflows", "myflow.yaml")
	_, err = os.Stat(workflowFile)
	assert.NoError(t, err, "workflow file should exist")
}

func TestAdd_WorkflowWithType(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupCmdTest(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"jsonflow", "--type", "json"})

	err := Command.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Created workflow:")
	assert.Contains(t, output, "jsonflow.json")

	workflowFile := filepath.Join(dir, "floww", "workflows", "jsonflow.json")
	_, err = os.Stat(workflowFile)
	assert.NoError(t, err, "workflow file should exist")
}

func TestAdd_ExistingWorkflowSameExtension(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupCmdTest(t, dir)

	existing := filepath.Join(wfDir, "existing.yaml")
	err := os.WriteFile(existing, []byte("description: test\nworkspaces: []\n"), 0600)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"existing"})

	err = Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAdd_ExistingWorkflowDifferentExtension(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupCmdTest(t, dir)

	existing := filepath.Join(wfDir, "crosstype.json")
	err := os.WriteFile(existing, []byte(`{"description": "test", "workspaces": []}`), 0600)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"crosstype", "--type", "yaml"})

	err = Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Contains(t, err.Error(), ".json")
}

func TestAdd_ExistingWorkflowMultipleExtensions(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupCmdTest(t, dir)

	err := os.WriteFile(filepath.Join(wfDir, "multitype.json"), []byte(`{}`), 0600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(wfDir, "multitype.yaml"), []byte(""), 0600)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"multitype", "--type", "toml"})

	err = Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Contains(t, err.Error(), ".json")
	assert.Contains(t, err.Error(), ".yaml")
}

func TestAdd_WorkflowNameWithSlash(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupCmdTest(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"my/workflow"})

	err := Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot contain path separators")
}

func TestAdd_WorkflowNameWithDot(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupCmdTest(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"my.workflow"})

	err := Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "without file extension")
}

func TestAdd_WorkflowNameStartingWithDot(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupCmdTest(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{".hidden"})

	err := Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot start with a dot")
}
