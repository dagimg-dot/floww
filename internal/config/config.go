package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/dagimg-dot/floww/internal/utils"
	"github.com/dagimg-dot/floww/internal/workflow"
)

// ConfigManager manages floww configuration — loading, merging, initializing,
// and listing workflows. NOT a singleton; must be passed explicitly via
// dependency injection.
type ConfigManager struct {
	configPath   string
	configDir    string
	workflowsDir string
	loader       *ConfigLoader
	config       *utils.DefaultConfig
	mu           sync.RWMutex
}

// NewConfigManager creates a new ConfigManager.
//
// If configPath is provided it is used as the config file path (or as a
// directory when the path has no recognised extension).  Otherwise the
// XDG_CONFIG_HOME convention is followed, defaulting to
// ~/.config/floww/config.yaml.
func NewConfigManager(configPath ...string) *ConfigManager {
	cm := &ConfigManager{
		loader: NewConfigLoader(),
	}

	if len(configPath) > 0 && configPath[0] != "" {
		cp := configPath[0]
		if cm.loader.IsSupportedFormat(cp) {
			cm.configPath = cp
			cm.configDir = filepath.Dir(cp)
		} else {
			cm.configDir = cp
			cm.configPath = filepath.Join(cp, "config.yaml")
		}
	} else {
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				home = "~"
			}
			xdg = filepath.Join(home, ".config")
		}
		cm.configDir = filepath.Join(xdg, "floww")
		cm.configPath = filepath.Join(cm.configDir, "config.yaml")
	}

	cm.workflowsDir = filepath.Join(cm.configDir, "workflows")

	defaults := utils.DefaultConfigValues
	cm.config = &defaults

	// Load and merge user config from disk; falls back to defaults silently.
	_ = cm.loadAndMergeConfig()

	return cm
}

// Init creates the config directory, the workflows sub-directory, and the
// default config file (if none exists).  When createExample is true an
// example workflow file is written into the workflows directory using the
// given fileType ("yaml", "json", "toml", or "" → "yaml").
func (cm *ConfigManager) Init(createExample bool, fileType string) error {
	// Config directory
	if err := os.MkdirAll(cm.configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.MkdirAll(cm.workflowsDir, 0750); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		if err := cm.writeDefaultConfig(); err != nil {
			return err
		}
	}

	if createExample {
		if err := cm.writeExampleWorkflow(fileType); err != nil {
			return err
		}
	}

	return nil
}

func (cm *ConfigManager) writeDefaultConfig() error {
	data := map[string]any{
		"general": map[string]any{
			"show_notifications": true,
			"workspace_backend":  "auto",
		},
		"timing": map[string]any{
			"workspace_switch_wait": 3,
			"app_launch_wait":       1,
			"respect_app_wait":      true,
		},
	}
	if err := cm.loader.Save(data, cm.configPath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	return nil
}

func (cm *ConfigManager) writeExampleWorkflow(fileType string) error {
	if fileType == "" {
		fileType = "yaml"
	}
	ext := "." + fileType
	if !cm.loader.IsSupportedFormat("x" + ext) {
		ext = ".yaml"
	}

	examplePath := filepath.Join(cm.workflowsDir, "example"+ext)
	if _, err := os.Stat(examplePath); err == nil {
		return nil
	}

	if err := cm.loader.Save(sampleWorkflowToMap(), examplePath); err != nil {
		return fmt.Errorf("failed to create example workflow: %w", err)
	}
	return nil
}

// IsInitialized returns true when both the config directory and the workflows
// sub-directory exist.
func (cm *ConfigManager) IsInitialized() bool {
	info, err := os.Stat(cm.configDir)
	if err != nil || !info.IsDir() {
		return false
	}
	info, err = os.Stat(cm.workflowsDir)
	if err != nil || !info.IsDir() {
		return false
	}
	return true
}

// ListWorkflowNames scans the workflows directory, collects all recognised
// workflow files, deduplicates by stem (same stem in multiple formats counts
// once), and returns the sorted list of names.
func (cm *ConfigManager) ListWorkflowNames() []string {
	entries, err := os.ReadDir(cm.workflowsDir)
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	var names []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !cm.loader.IsSupportedFormat(name) {
			continue
		}
		stem := strings.TrimSuffix(name, filepath.Ext(name))
		if !seen[stem] {
			seen[stem] = true
			names = append(names, stem)
		}
	}

	sort.Strings(names)
	return names
}

// ResolveWorkflowPath resolves a workflow name to a file path within the
// workflows directory, trying each supported extension in order
// (.toml → .yaml → .yml → .json). Returns *WorkflowNotFoundError when no
// matching file exists.
func (cm *ConfigManager) ResolveWorkflowPath(name string) (string, error) {
	for _, ext := range cm.loader.GetSupportedFormats() {
		candidate := filepath.Join(cm.workflowsDir, name+ext)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", &WorkflowNotFoundError{
		ConfigError: ConfigError{
			FlowwError: FlowwError{
				Msg: fmt.Sprintf("Workflow '%s' not found", name),
			},
		},
	}
}

// LoadWorkflow loads a workflow by name (or by direct file path when
// isDirectLoad is true).  For name lookups the method tries each supported
// extension in the order .toml → .yaml → .yml → .json within the workflows
// directory.  The loaded workflow is validated before being returned.
func (cm *ConfigManager) LoadWorkflow(name string, isDirectLoad bool) (*workflow.Workflow, error) {
	var path string

	if isDirectLoad {
		if _, err := os.Stat(name); err == nil {
			path = name
		} else {
			return nil, &WorkflowNotFoundError{
				ConfigError: ConfigError{
					FlowwError: FlowwError{
						Msg: fmt.Sprintf("Workflow file '%s' not found", name),
					},
				},
			}
		}
	} else {
		var err error
		path, err = cm.ResolveWorkflowPath(name)
		if err != nil {
			return nil, err
		}
	}

	return cm.loadWorkflowFile(path)
}

// loadWorkflowFile reads a workflow file at the given path, parses it into a
// Workflow struct, and validates it.
func (cm *ConfigManager) loadWorkflowFile(path string) (*workflow.Workflow, error) {
	raw, err := cm.loader.Load(path)
	if err != nil {
		return nil, &ConfigLoadError{
			ConfigError: ConfigError{
				FlowwError: FlowwError{
					Msg:   fmt.Sprintf("failed to load workflow file '%s'", path),
					Cause: err,
				},
			},
		}
	}

	// Round-trip through JSON to convert the generic map to a typed struct.
	// This works because every supported format can be represented as JSON,
	// and the Workflow struct has json struct tags.
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow data: %w", err)
	}

	var wf workflow.Workflow
	if err := json.Unmarshal(encoded, &wf); err != nil {
		return nil, &ConfigLoadError{
			ConfigError: ConfigError{
				FlowwError: FlowwError{
					Msg:   fmt.Sprintf("failed to parse workflow file '%s'", path),
					Cause: err,
				},
			},
		}
	}

	return &wf, nil
}

// ValidateWorkflow delegates to the workflow package's schema validator.
func (cm *ConfigManager) ValidateWorkflow(name string, data *workflow.Workflow) error {
	return workflow.ValidateWorkflow(name, data)
}

// GetConfig returns a read-only pointer to the current merged configuration.
func (cm *ConfigManager) GetConfig() *utils.DefaultConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// GetTimingConfig returns a read-only pointer to the timing sub-section.
func (cm *ConfigManager) GetTimingConfig() *utils.TimingConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return &cm.config.Timing
}

// ConfigPath returns the resolved config file path.
func (cm *ConfigManager) ConfigPath() string {
	return cm.configPath
}

// ConfigDir returns the resolved config directory path.
func (cm *ConfigManager) ConfigDir() string {
	return cm.configDir
}

// GetGeneralConfig returns a read-only pointer to the general sub-section.
func (cm *ConfigManager) GetGeneralConfig() *utils.GeneralConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return &cm.config.General
}

// loadAndMergeConfig loads the user's config file from disk and deep-merges
// it with the hard-coded defaults.  Unknown or invalid values revert to the
// corresponding default.  This method is safe for concurrent callers.
func (cm *ConfigManager) loadAndMergeConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	merged := utils.DefaultConfigValues

	raw, err := cm.loadMainConfigFile()
	if err != nil {
		cm.config = &merged
		return nil
	}

	if generalRaw, ok := raw["general"]; ok {
		if general, ok := generalRaw.(map[string]any); ok {
			if v, ok := general["show_notifications"]; ok {
				if b, ok := v.(bool); ok {
					merged.General.ShowNotifications = b
				}
			}
			if v, ok := general["workspace_backend"]; ok {
				if s, ok := v.(string); ok {
					if utils.ValidWorkspaceBackends[s] {
						merged.General.WorkspaceBackend = s
					}
				}
			}
		}
	}

	if timingRaw, ok := raw["timing"]; ok {
		if timing, ok := timingRaw.(map[string]any); ok {
			if v, ok := timing["workspace_switch_wait"]; ok {
				if f, convOk := toFloat64(v); convOk && f >= 0 {
					merged.Timing.WorkspaceSwitchWait = f
				}
			}
			if v, ok := timing["app_launch_wait"]; ok {
				if f, convOk := toFloat64(v); convOk && f >= 0 {
					merged.Timing.AppLaunchWait = f
				}
			}
			if v, ok := timing["respect_app_wait"]; ok {
				if b, ok := v.(bool); ok {
					merged.Timing.RespectAppWait = b
				}
			}
		}
	}

	cm.config = &merged
	return nil
}

// loadMainConfigFile finds and loads the user's config file.
//
// The search order is:
//  1. config.yaml (preferred — checked first regardless of extension order)
//  2. config.toml, config.yaml (skipped), config.yml, config.json
//
// The config base is derived by stripping the extension from configPath (or
// using it verbatim when there is no extension).
func (cm *ConfigManager) loadMainConfigFile() (map[string]any, error) {
	base := cm.configPath
	ext := filepath.Ext(base)
	if ext != "" {
		base = strings.TrimSuffix(base, ext)
	}

	yamlPath := base + ".yaml"
	if _, err := os.Stat(yamlPath); err == nil {
		return cm.loader.Load(yamlPath)
	}

	for _, ext := range cm.loader.GetSupportedFormats() {
		if ext == ".yaml" {
			continue
		}
		path := base + ext
		if _, err := os.Stat(path); err == nil {
			return cm.loader.Load(path)
		}
	}

	return nil, fmt.Errorf("config file not found at %s.{yaml,toml,yml,json}", base)
}

// WorkflowsDir returns the absolute path to the workflows directory.
func (cm *ConfigManager) WorkflowsDir() string {
	return cm.workflowsDir
}

// ConfigLoader returns the underlying ConfigLoader, which provides
// format-agnostic Load and Save methods.
func (cm *ConfigManager) ConfigLoader() *ConfigLoader {
	return cm.loader
}

// GetSupportedFormats returns the ordered list of supported file extensions
// (.toml, .yaml, .yml, .json).
func (cm *ConfigManager) GetSupportedFormats() []string {
	return cm.loader.GetSupportedFormats()
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func sampleWorkflowToMap() map[string]any {
	sample := utils.SampleWorkflowContent
	workspaces := make([]any, len(sample.Workspaces))
	for i, ws := range sample.Workspaces {
		apps := make([]any, len(ws.Apps))
		for j, app := range ws.Apps {
			m := map[string]any{
				"name": app.Name,
				"exec": app.Exec,
			}
			if len(app.Args) > 0 {
				m["args"] = app.Args
			}
			apps[j] = m
		}
		workspaces[i] = map[string]any{
			"target": ws.Target,
			"apps":   apps,
		}
	}
	return map[string]any{
		"description": sample.Description,
		"workspaces":  workspaces,
	}
}
