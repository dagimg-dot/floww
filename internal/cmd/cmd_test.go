package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dagimg-dot/floww/internal/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const subTestEnv = "_FLOWW_SUB_TEST"

func resetRootCmd() {
	rootCmd.SetArgs([]string{})
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	visitAllFlags(rootCmd.Flags())
	visitAllFlags(rootCmd.PersistentFlags())
	for _, sub := range rootCmd.Commands() {
		visitAllFlags(sub.Flags())
	}
}

func visitAllFlags(flagSet *pflag.FlagSet) {
	if flagSet == nil {
		return
	}
	flagSet.VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

func runSubTest(t *testing.T) (string, error) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$") //nolint:gosec // Intentional test helper
	cmd.Env = append(os.Environ(), subTestEnv+"=1")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func validWorkflowData() map[string]any {
	return map[string]any{
		"description": "test-workflow",
		"workspaces": []any{
			map[string]any{
				"target": 0,
				"apps": []any{
					map[string]any{
						"name": "Test App",
						"exec": "true",
					},
				},
			},
		},
	}
}

func TestInit_CreatesConfigDir(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		resetRootCmd()
		rootCmd.SetArgs([]string{"init"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Initialized config at"))
}

func TestInit_WithExampleCreatesExampleWorkflow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfg := config.NewConfigManager()
	require.NoError(t, cfg.Init(true, ""))
	examplePath := filepath.Join(cfg.WorkflowsDir(), "example.yaml")
	_, err := os.Stat(examplePath)
	require.NoError(t, err)
}

func TestInit_WithExampleAndJSONType(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfg := config.NewConfigManager()
	require.NoError(t, cfg.Init(true, "json"))
	examplePath := filepath.Join(cfg.WorkflowsDir(), "example.json")
	info, err := os.Stat(examplePath)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
}

func TestList_NoWorkflows(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		resetRootCmd()
		rootCmd.SetArgs([]string{"list"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "No workflows found"))
}

func TestList_WithWorkflows(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		for _, name := range []string{"dev.yaml", "prod.yaml", "test.toml"} {
			require.NoError(t, loader.Save(validWorkflowData(), filepath.Join(cfg.WorkflowsDir(), name)))
		}
		resetRootCmd()
		rootCmd.SetArgs([]string{"list"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Available workflows:"))
	assert.True(t, strings.Contains(out, "  - dev"))
	assert.True(t, strings.Contains(out, "  - prod"))
	assert.True(t, strings.Contains(out, "  - test"))
}

func TestAdd_NewWorkflow(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		resetRootCmd()
		rootCmd.SetArgs([]string{"add", "myworkflow"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "myworkflow.yaml"))
}

func TestAdd_WithJSONType(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		resetRootCmd()
		rootCmd.SetArgs([]string{"add", "myworkflow", "--type", "json"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "myworkflow.json"))
}

func TestAdd_ExistingSameExtension(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		require.NoError(t, loader.Save(validWorkflowData(), filepath.Join(cfg.WorkflowsDir(), "existing.yaml")))
		rootCmd.SetArgs([]string{"add", "existing"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "already exists with extension"))
}

func TestAdd_ExistingDifferentExtension(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		require.NoError(t, loader.Save(validWorkflowData(), filepath.Join(cfg.WorkflowsDir(), "existing.toml")))
		rootCmd.SetArgs([]string{"add", "existing", "--type", "yaml"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "already exists with extension"))
}

func TestAdd_PathSeparatorError(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		rootCmd.SetArgs([]string{"add", "test/name"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "name cannot contain path separators"))
}

func TestAdd_LeadingDotError(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		rootCmd.SetArgs([]string{"add", ".hidden"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "name cannot start with a dot"))
}

func TestAdd_ExtensionInNameError(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		rootCmd.SetArgs([]string{"add", "test.yaml"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "name without file extension"))
}

func TestEdit_ExistingWorkflow(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		t.Setenv("EDITOR", "true")
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		require.NoError(t, loader.Save(validWorkflowData(), filepath.Join(cfg.WorkflowsDir(), "test.yaml")))
		resetRootCmd()
		rootCmd.SetArgs([]string{"edit", "test"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Opening workflow"))
	assert.True(t, strings.Contains(out, "test"))
}

func TestEdit_NonexistentWorkflow(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		rootCmd.SetArgs([]string{"edit", "nonexistent"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "not found"))
}

func TestRemove_SingleWithForce(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		wfPath := filepath.Join(cfg.WorkflowsDir(), "test.yaml")
		require.NoError(t, loader.Save(validWorkflowData(), wfPath))
		resetRootCmd()
		rootCmd.SetArgs([]string{"remove", "test", "--force"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Removed workflow:"))
	assert.True(t, strings.Contains(out, "test.yaml"))
}

func TestRemove_MultipleWithForce(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		wf1 := filepath.Join(cfg.WorkflowsDir(), "wf1.yaml")
		wf2 := filepath.Join(cfg.WorkflowsDir(), "wf2.toml")
		require.NoError(t, loader.Save(validWorkflowData(), wf1))
		require.NoError(t, loader.Save(validWorkflowData(), wf2))
		resetRootCmd()
		rootCmd.SetArgs([]string{"remove", "wf1", "wf2", "--force"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Removed workflow: wf1.yaml"))
	assert.True(t, strings.Contains(out, "Removed workflow: wf2.toml"))
}

func TestRemove_NonexistentWithForce(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		rootCmd.SetArgs([]string{"remove", "nonexistent", "--force"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "not found"))
}

func TestValidate_ValidWorkflow(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		require.NoError(t, loader.Save(validWorkflowData(), filepath.Join(cfg.WorkflowsDir(), "test.yaml")))
		resetRootCmd()
		rootCmd.SetArgs([]string{"validate", "test"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Validating workflow: test"))
	assert.True(t, strings.Contains(out, "Workflow is valid"))
}

func TestValidate_InvalidWorkflow(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		invalidData := map[string]any{
			"description": "invalid",
			"workspaces": []any{
				map[string]any{
					"target": 0,
					"apps": []any{
						map[string]any{
							"exec": "true",
						},
					},
				},
			},
		}
		require.NoError(t, loader.Save(invalidData, filepath.Join(cfg.WorkflowsDir(), "invalid.yaml")))
		rootCmd.SetArgs([]string{"validate", "invalid"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "validation failed"))
}

func TestValidate_NonexistentWorkflow(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		rootCmd.SetArgs([]string{"validate", "nonexistent"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "not found"))
}

func TestApply_ByName(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		require.NoError(t, loader.Save(validWorkflowData(), filepath.Join(cfg.WorkflowsDir(), "test.yaml")))
		resetRootCmd()
		rootCmd.SetArgs([]string{"apply", "test"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Workflow: test-workflow"))
}

func TestApply_WithFile(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		loader := config.NewConfigLoader()
		wfPath := filepath.Join(dir, "external.yaml")
		require.NoError(t, loader.Save(validWorkflowData(), wfPath))
		resetRootCmd()
		rootCmd.SetArgs([]string{"apply", "--file", wfPath})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Workflow: test-workflow"))
}

func TestApply_WithFileWithoutInit(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		loader := config.NewConfigLoader()
		wfPath := filepath.Join(dir, "external.yaml")
		require.NoError(t, loader.Save(validWorkflowData(), wfPath))
		resetRootCmd()
		rootCmd.SetArgs([]string{"apply", "--file", wfPath})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "--> Switching to workspace 0"))
}

func TestApply_WithFileNonexistent(t *testing.T) {
	resetRootCmd()
	rootCmd.SetArgs([]string{"apply", "--file", "/nonexistent/path/workflow.yaml"})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestApply_WithAppend(t *testing.T) {
	if os.Getenv(subTestEnv) == "1" {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)
		cfg := config.NewConfigManager()
		require.NoError(t, cfg.Init(false, ""))
		loader := config.NewConfigLoader()
		require.NoError(t, loader.Save(validWorkflowData(), filepath.Join(cfg.WorkflowsDir(), "test.yaml")))
		resetRootCmd()
		rootCmd.SetArgs([]string{"apply", "test", "--append"})
		assert.NoError(t, rootCmd.Execute())
		return
	}
	out, _ := runSubTest(t)
	assert.True(t, strings.Contains(out, "Workflow: test-workflow"))
}
