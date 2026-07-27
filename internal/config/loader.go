package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// ConfigLoader provides format-agnostic loading and saving of configuration
// files in YAML, JSON, and TOML formats. Dispatch happens by file extension.
type ConfigLoader struct {
	loaders    map[string]func(string) (map[string]any, error)
	savers     map[string]func(map[string]any, string) error
	extensions []string // ordered list of supported extensions
}

// NewConfigLoader creates a ConfigLoader with dispatch maps for
// .toml, .yaml, .yml, and .json extensions.
func NewConfigLoader() *ConfigLoader {
	cl := &ConfigLoader{
		loaders:    make(map[string]func(string) (map[string]any, error)),
		savers:     make(map[string]func(map[string]any, string) error),
		extensions: []string{".toml", ".yaml", ".yml", ".json"},
	}
	cl.loaders[".toml"] = cl.loadTOML
	cl.loaders[".yaml"] = cl.loadYAML
	cl.loaders[".yml"] = cl.loadYAML
	cl.loaders[".json"] = cl.loadJSON

	cl.savers[".toml"] = cl.saveTOML
	cl.savers[".yaml"] = cl.saveYAML
	cl.savers[".yml"] = cl.saveYAML
	cl.savers[".json"] = cl.saveJSON

	return cl
}

// Load reads a configuration file and returns its contents as a map.
// The file format is determined by its extension (case-insensitive).
func (cl *ConfigLoader) Load(path string) (map[string]any, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return nil, fmt.Errorf("unsupported configuration format: %s", ext)
	}

	loader, ok := cl.loaders[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported configuration format: %s", ext)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	return loader(path)
}

// Save writes the configuration data to a file in the format determined
// by the file's extension (case-insensitive).
func (cl *ConfigLoader) Save(data map[string]any, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return fmt.Errorf("unsupported configuration format: %s", ext)
	}

	saver, ok := cl.savers[ext]
	if !ok {
		return fmt.Errorf("unsupported configuration format: %s", ext)
	}

	return saver(data, path)
}

// GetSupportedFormats returns the list of supported file extensions in order:
// .toml, .yaml, .yml, .json.
func (cl *ConfigLoader) GetSupportedFormats() []string {
	result := make([]string, len(cl.extensions))
	copy(result, cl.extensions)
	return result
}

// IsSupportedFormat returns true if the given file path has a supported
// configuration file extension (case-insensitive).
func (cl *ConfigLoader) IsSupportedFormat(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := cl.loaders[ext]
	return ok
}

func (cl *ConfigLoader) loadYAML(path string) (map[string]any, error) {
	data, err := os.ReadFile(path) //nolint:gosec // Intentional file open
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if result == nil {
		return map[string]any{}, nil
	}
	return result, nil
}

func (cl *ConfigLoader) saveYAML(data map[string]any, path string) error {
	f, err := os.Create(path) //nolint:gosec // Intentional file create
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	encoder := yaml.NewEncoder(f)
	defer encoder.Close() //nolint:errcheck

	// default_flow_style=False equivalent: yaml.v3 uses block style by default
	// sort_keys=False equivalent: yaml.v3 does not sort keys by default
	return encoder.Encode(data)
}

func (cl *ConfigLoader) loadJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path) //nolint:gosec // Intentional file open
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (cl *ConfigLoader) saveJSON(data map[string]any, path string) error {
	f, err := os.Create(path) //nolint:gosec // Intentional file create
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (cl *ConfigLoader) loadTOML(path string) (map[string]any, error) {
	data, err := os.ReadFile(path) //nolint:gosec // Intentional file open
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if result == nil {
		return map[string]any{}, nil
	}
	return result, nil
}

func (cl *ConfigLoader) saveTOML(data map[string]any, path string) error {
	f, err := os.Create(path) //nolint:gosec // path comes from ConfigManager's WorkflowsDir
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	encoder := toml.NewEncoder(f)
	return encoder.Encode(data)
}
