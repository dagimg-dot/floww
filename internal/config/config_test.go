package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dagimg-dot/floww/internal/utils"
	"github.com/dagimg-dot/floww/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigManager_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cm := NewConfigManager()
	assert.Equal(t, filepath.Join(dir, "floww"), cm.configDir)
	assert.Equal(t, filepath.Join(dir, "floww", "config.yaml"), cm.configPath)
	assert.Equal(t, filepath.Join(dir, "floww", "workflows"), cm.workflowsDir)
}

func TestNewConfigManager_CustomFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "myconfig.toml")
	cm := NewConfigManager(cfgPath)
	assert.Equal(t, cfgPath, cm.configPath)
	assert.Equal(t, dir, cm.configDir)
	assert.Equal(t, filepath.Join(dir, "workflows"), cm.workflowsDir)
}

func TestNewConfigManager_CustomDir(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	assert.Equal(t, dir, cm.configDir)
	assert.Equal(t, filepath.Join(dir, "config.yaml"), cm.configPath)
	assert.Equal(t, filepath.Join(dir, "workflows"), cm.workflowsDir)
}

func TestNewConfigManager_XDG_CONFIG_HOME(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cm := NewConfigManager()
	assert.Equal(t, filepath.Join(dir, "floww"), cm.configDir)
}

func TestNewConfigManager_DefaultsDeepCopy(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cm := NewConfigManager()
	cfg := cm.GetConfig()
	assert.Equal(t, float64(3), cfg.Timing.WorkspaceSwitchWait)
	assert.Equal(t, float64(1), cfg.Timing.AppLaunchWait)
	assert.True(t, cfg.Timing.RespectAppWait)
	assert.True(t, cfg.General.ShowNotifications)
	assert.Equal(t, "auto", cfg.General.WorkspaceBackend)
}

func TestInit_CreatesDirectoriesAndConfig(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	err := cm.Init(false, "")
	require.NoError(t, err)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	wfDir := filepath.Join(dir, "workflows")
	info, err = os.Stat(wfDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	cfgFile := filepath.Join(dir, "config.yaml")
	info, err = os.Stat(cfgFile)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
}

func TestInit_WithExample(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	err := cm.Init(true, "")
	require.NoError(t, err)

	examplePath := filepath.Join(dir, "workflows", "example.yaml")
	info, err := os.Stat(examplePath)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
}

func TestInit_WithCustomType(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	err := cm.Init(true, "toml")
	require.NoError(t, err)

	examplePath := filepath.Join(dir, "workflows", "example.toml")
	info, err := os.Stat(examplePath)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())

	raw, err := cm.loader.Load(examplePath)
	require.NoError(t, err)
	assert.Equal(t, "An example workflow.", raw["description"])
}

func TestInit_Idempotent(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)

	err := cm.Init(true, "")
	require.NoError(t, err)

	err = cm.Init(true, "")
	require.NoError(t, err)
}

func TestInit_WithInvalidFileTypeFallsBack(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	err := cm.Init(true, "unsupported")
	require.NoError(t, err)

	examplePath := filepath.Join(dir, "workflows", "example.yaml")
	_, err = os.Stat(examplePath)
	assert.NoError(t, err)
}

func TestIsInitialized_BeforeInit(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(filepath.Join(dir, "nonexistent"))
	assert.False(t, cm.IsInitialized())
}

func TestIsInitialized_AfterInit(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	err := cm.Init(false, "")
	require.NoError(t, err)
	assert.True(t, cm.IsInitialized())
}

func TestIsInitialized_MissingWorkflowsDir(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)

	err := os.MkdirAll(dir, 0750)
	require.NoError(t, err)

	assert.False(t, cm.IsInitialized())
}

func TestListWorkflowNames_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	_ = cm.Init(false, "")
	names := cm.ListWorkflowNames()
	assert.Empty(t, names)
}

func TestListWorkflowNames_MultipleFormats(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	_ = cm.Init(false, "")

	wfDir := cm.workflowsDir

	createWorkflowFile(t, wfDir, "dev.yaml", "test-description")
	createWorkflowFile(t, wfDir, "prod.toml", "prod-description")
	createWorkflowFile(t, wfDir, "test.json", "test-json-desc")
	createWorkflowFile(t, wfDir, "docs.yml", "docs-desc")

	names := cm.ListWorkflowNames()
	expected := []string{"dev", "docs", "prod", "test"}
	assert.Equal(t, expected, names)
}

func TestListWorkflowNames_Dedup(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	_ = cm.Init(false, "")

	wfDir := cm.workflowsDir

	createWorkflowFile(t, wfDir, "dev.yaml", "yaml-version")
	createWorkflowFile(t, wfDir, "dev.toml", "toml-version")

	names := cm.ListWorkflowNames()
	assert.Equal(t, []string{"dev"}, names)
}

func TestListWorkflowNames_IgnoresNonWorkflowFiles(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	_ = cm.Init(false, "")

	wfDir := cm.workflowsDir
	createWorkflowFile(t, wfDir, "dev.yaml", "desc")
	// Create a non-workflow file directly (not through loader, which rejects .txt)
	err := os.WriteFile(filepath.Join(wfDir, "readme.txt"), []byte("hello"), 0600)
	require.NoError(t, err)

	names := cm.ListWorkflowNames()
	assert.Equal(t, []string{"dev"}, names)
}

func TestListWorkflowNames_Uninitialized(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(filepath.Join(dir, "does-not-exist"))
	names := cm.ListWorkflowNames()
	assert.Nil(t, names)
}

func TestLoadWorkflow_ByName(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	_ = cm.Init(false, "")

	wfDir := cm.workflowsDir
	createValidWorkflowFile(t, wfDir, "dev.yaml", 2)

	wf, err := cm.LoadWorkflow("dev", false)
	require.NoError(t, err)
	require.NotNil(t, wf)
	assert.Equal(t, "test-workflow", wf.Description)
	assert.Len(t, wf.Workspaces, 2)
}

func TestLoadWorkflow_ByNameNotFound(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	_ = cm.Init(false, "")

	wf, err := cm.LoadWorkflow("nonexistent", false)
	assert.Error(t, err)
	assert.Nil(t, wf)
	assert.IsType(t, &WorkflowNotFoundError{}, err)
}

func TestLoadWorkflow_DirectLoad(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)

	wfPath := filepath.Join(dir, "external.yaml")
	createValidWorkflowFile(t, dir, "external.yaml", 1)

	wf, err := cm.LoadWorkflow(wfPath, true)
	require.NoError(t, err)
	require.NotNil(t, wf)
	assert.Equal(t, "test-workflow", wf.Description)
}

func TestLoadWorkflow_DirectLoadNotFound(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)

	wf, err := cm.LoadWorkflow("/nonexistent/path/workflow.yaml", true)
	assert.Error(t, err)
	assert.Nil(t, wf)
	assert.IsType(t, &WorkflowNotFoundError{}, err)
}

func TestLoadWorkflow_FormatPriority(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)
	_ = cm.Init(false, "")

	wfDir := cm.workflowsDir

	createValidWorkflowFile(t, wfDir, "multi.yaml", 1)
	createValidWorkflowWithDesc(t, wfDir, "multi.toml", 2, "toml-version")

	wf, err := cm.LoadWorkflow("multi", false)
	require.NoError(t, err)
	assert.Equal(t, "toml-version", wf.Description)
	assert.Len(t, wf.Workspaces, 2)
}

func TestValidateWorkflow_Valid(t *testing.T) {
	cm := NewConfigManager()
	wf := &workflow.Workflow{
		Description: "test",
		Workspaces: []workflow.Workspace{
			{Target: 0, Apps: []workflow.App{
				{Name: "Terminal", Exec: "gnome-terminal"},
			}},
		},
	}
	err := cm.ValidateWorkflow("test", wf)
	assert.NoError(t, err)
}

func TestValidateWorkflow_Invalid(t *testing.T) {
	cm := NewConfigManager()
	wf := &workflow.Workflow{
		Description: "invalid",
		Workspaces: []workflow.Workspace{
			{Target: 0, Apps: nil},
		},
	}
	err := cm.ValidateWorkflow("invalid", wf)
	assert.Error(t, err)
}

func TestValidateWorkflow_Nil(t *testing.T) {
	cm := NewConfigManager()
	err := cm.ValidateWorkflow("nil", nil)
	assert.Error(t, err)
}

func TestGetConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cm := NewConfigManager()
	cfg := cm.GetConfig()
	assert.Equal(t, &utils.DefaultConfigValues, cfg)
}

func TestGetTimingConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cm := NewConfigManager()
	timing := cm.GetTimingConfig()
	assert.Equal(t, &utils.DefaultConfigValues.Timing, timing)
}

func TestGetGeneralConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cm := NewConfigManager()
	general := cm.GetGeneralConfig()
	assert.Equal(t, &utils.DefaultConfigValues.General, general)
}

func TestLoadAndMergeConfig_DefaultsWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	cm := NewConfigManager(dir)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.Equal(t, float64(3), cfg.Timing.WorkspaceSwitchWait)
	assert.Equal(t, "auto", cfg.General.WorkspaceBackend)
	assert.True(t, cfg.General.ShowNotifications)
}

func TestLoadAndMergeConfig_MergesConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cm := NewConfigManager(cfgPath)

	configData := map[string]any{
		"general": map[string]any{
			"show_notifications": false,
			"workspace_backend":  "hyprland",
		},
		"timing": map[string]any{
			"workspace_switch_wait": 5,
		},
	}
	err := cm.loader.Save(configData, cfgPath)
	require.NoError(t, err)

	err = cm.loadAndMergeConfig()
	require.NoError(t, err)

	cfg := cm.GetConfig()
	assert.False(t, cfg.General.ShowNotifications)
	assert.Equal(t, "hyprland", cfg.General.WorkspaceBackend)
	assert.Equal(t, float64(5), cfg.Timing.WorkspaceSwitchWait)
	assert.Equal(t, float64(1), cfg.Timing.AppLaunchWait)
	assert.True(t, cfg.Timing.RespectAppWait)
}

func TestLoadAndMergeConfig_TimingNegativeValue(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cm := NewConfigManager(cfgPath)

	configData := map[string]any{
		"timing": map[string]any{
			"workspace_switch_wait": -1,
			"app_launch_wait":       -5,
		},
	}
	err := cm.loader.Save(configData, cfgPath)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.Equal(t, float64(3), cfg.Timing.WorkspaceSwitchWait)
	assert.Equal(t, float64(1), cfg.Timing.AppLaunchWait)
}

func TestLoadAndMergeConfig_TimingNonNumeric(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cm := NewConfigManager(cfgPath)

	configData := map[string]any{
		"timing": map[string]any{
			"workspace_switch_wait": "not-a-number",
			"app_launch_wait":       true,
		},
	}
	err := cm.loader.Save(configData, cfgPath)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.Equal(t, float64(3), cfg.Timing.WorkspaceSwitchWait)
	assert.Equal(t, float64(1), cfg.Timing.AppLaunchWait)
}

func TestLoadAndMergeConfig_NonBoolRespectAppWait(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cm := NewConfigManager(cfgPath)

	configData := map[string]any{
		"timing": map[string]any{
			"respect_app_wait": "not-a-bool",
		},
	}
	err := cm.loader.Save(configData, cfgPath)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.True(t, cfg.Timing.RespectAppWait)
}

func TestLoadAndMergeConfig_InvalidBackend(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cm := NewConfigManager(cfgPath)

	configData := map[string]any{
		"general": map[string]any{
			"workspace_backend": "invalid-backend",
		},
	}
	err := cm.loader.Save(configData, cfgPath)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.Equal(t, "auto", cfg.General.WorkspaceBackend)
}

func TestLoadAndMergeConfig_ConfigYamlPreferred(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cm := NewConfigManager(cfgPath)

	yamlData := map[string]any{
		"general": map[string]any{
			"workspace_backend": "hyprland",
		},
	}
	err := cm.loader.Save(yamlData, cfgPath)
	require.NoError(t, err)

	tomlData := map[string]any{
		"general": map[string]any{
			"workspace_backend": "niri",
		},
	}
	tomlPath := filepath.Join(dir, "config.toml")
	err = cm.loader.Save(tomlData, tomlPath)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.Equal(t, "hyprland", cfg.General.WorkspaceBackend)
}

func TestLoadAndMergeConfig_ConfigTomlFallback(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cm := NewConfigManager(cfgPath)

	tomlData := map[string]any{
		"general": map[string]any{
			"workspace_backend": "niri",
		},
	}
	tomlPath := filepath.Join(dir, "config.toml")
	err := cm.loader.Save(tomlData, tomlPath)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.Equal(t, "niri", cfg.General.WorkspaceBackend)
}

func TestLoadAndMergeConfig_TimingInt64FromTOML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	cm := NewConfigManager(cfgPath)

	tomlPath := filepath.Join(dir, "config.toml")
	tomlContent := []byte(`[timing]
workspace_switch_wait = 7
app_launch_wait = 2
respect_app_wait = false

[general]
show_notifications = true
workspace_backend = "wmctrl"
`)
	err := os.WriteFile(tomlPath, tomlContent, 0600)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.Equal(t, float64(7), cfg.Timing.WorkspaceSwitchWait)
	assert.Equal(t, float64(2), cfg.Timing.AppLaunchWait)
	assert.False(t, cfg.Timing.RespectAppWait)
	assert.Equal(t, "wmctrl", cfg.General.WorkspaceBackend)
}

func TestLoadAndMergeConfig_JSONConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	cm := NewConfigManager(cfgPath)

	jsonData := map[string]any{
		"general": map[string]any{
			"show_notifications": false,
		},
	}
	err := cm.loader.Save(jsonData, cfgPath)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.False(t, cfg.General.ShowNotifications)
}

func TestLoadAndMergeConfig_EmptyConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cm := NewConfigManager(cfgPath)

	err := os.WriteFile(cfgPath, []byte{}, 0600)
	require.NoError(t, err)

	_ = cm.loadAndMergeConfig()

	cfg := cm.GetConfig()
	assert.Equal(t, utils.DefaultConfigValues, *cfg)
}

// createWorkflowFile writes a minimal workflow file at the given path.
func createWorkflowFile(t *testing.T, dir, name, description string) {
	t.Helper()
	loader := NewConfigLoader()
	data := map[string]any{
		"description": description,
		"workspaces": []any{
			map[string]any{
				"target": 0,
				"apps": []any{
					map[string]any{
						"name": "Terminal",
						"exec": "gnome-terminal",
					},
				},
			},
		},
	}
	err := loader.Save(data, filepath.Join(dir, name))
	require.NoError(t, err)
}

// createValidWorkflowFile writes a valid workflow with the given workspace count.
func createValidWorkflowFile(t *testing.T, dir, name string, wsCount int) {
	t.Helper()
	createValidWorkflowWithDesc(t, dir, name, wsCount, "test-workflow")
}

// createValidWorkflowWithDesc writes a valid workflow with a custom description.
func createValidWorkflowWithDesc(t *testing.T, dir, name string, wsCount int, desc string) {
	t.Helper()
	loader := NewConfigLoader()

	workspaces := make([]any, wsCount)
	for i := 0; i < wsCount; i++ {
		workspaces[i] = map[string]any{
			"target": i,
			"apps": []any{
				map[string]any{
					"name": "Terminal",
					"exec": "gnome-terminal",
				},
			},
		}
	}

	data := map[string]any{
		"description": desc,
		"workspaces":  workspaces,
	}
	err := loader.Save(data, filepath.Join(dir, name))
	require.NoError(t, err)
}
