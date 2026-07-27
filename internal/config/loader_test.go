package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_YAML_RoundTrip(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	original := map[string]any{
		"name":    "test",
		"version": 1,
		"nested": map[string]any{
			"key": "value",
		},
	}

	err := cl.Save(original, path)
	require.NoError(t, err, "Save should not error")

	loaded, err := cl.Load(path)
	require.NoError(t, err, "Load should not error")
	assert.Equal(t, original, loaded)
}

func TestLoader_YAML_NilToEmptyMap(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")

	// Write an empty YAML file
	err := os.WriteFile(path, []byte{}, 0600)
	require.NoError(t, err)

	result, err := cl.Load(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{}, result)
}

func TestLoader_YAML_OnlyComments(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "comments.yaml")

	err := os.WriteFile(path, []byte("# just a comment\n"), 0600)
	require.NoError(t, err)

	result, err := cl.Load(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{}, result)
}

func TestLoader_JSON_RoundTrip(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	original := map[string]any{
		"name":    "test",
		"version": float64(1),
		"nested": map[string]any{
			"key": "value",
		},
		"list": []any{float64(1), float64(2), float64(3)},
	}

	err := cl.Save(original, path)
	require.NoError(t, err, "Save should not error")

	loaded, err := cl.Load(path)
	require.NoError(t, err, "Load should not error")
	assert.Equal(t, original, loaded)
}

func TestLoader_JSON_Indent(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	data := map[string]any{"a": 1}
	err := cl.Save(data, path)
	require.NoError(t, err)

	raw, err := os.ReadFile(path) //nolint:gosec // test reads its own temp dir
	require.NoError(t, err)
	assert.Contains(t, string(raw), "\n  ")
}

func TestLoader_TOML_RoundTrip(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")

	original := map[string]any{
		"name":    "test",
		"version": int64(1),
		"nested": map[string]any{
			"key": "value",
		},
	}

	err := cl.Save(original, path)
	require.NoError(t, err, "Save should not error")

	loaded, err := cl.Load(path)
	require.NoError(t, err, "Load should not error")
	// TOML flattens some types; compare key by key
	assert.Equal(t, "test", loaded["name"])
	assert.Equal(t, int64(1), loaded["version"])
}

func TestLoader_TOML_NilFile(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.toml")

	err := os.WriteFile(path, []byte{}, 0600)
	require.NoError(t, err)

	result, err := cl.Load(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{}, result)
}

func TestLoader_UnsupportedFormat(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.unsupported")

	_, err := cl.Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported configuration format: .unsupported")
}

func TestLoader_UnsupportedFormatSave(t *testing.T) {
	cl := NewConfigLoader()
	err := cl.Save(map[string]any{}, "/tmp/test.unsupported")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported configuration format: .unsupported")
}

func TestLoader_NonexistentFile(t *testing.T) {
	cl := NewConfigLoader()
	path := "/nonexistent/path/config.yaml"

	_, err := cl.Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found:")
	assert.Contains(t, err.Error(), "/nonexistent/path/config.yaml")
}

func TestLoader_CaseInsensitiveExtension(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()

	// Test .YAML (uppercase)
	path := filepath.Join(dir, "test.YAML")
	err := os.WriteFile(path, []byte("key: value\n"), 0600)
	require.NoError(t, err)

	result, err := cl.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "value", result["key"])

	// Test .JSON (uppercase)
	jsonPath := filepath.Join(dir, "test.JSON")
	err = os.WriteFile(jsonPath, []byte(`{"a": 1}`), 0600)
	require.NoError(t, err)

	result, err = cl.Load(jsonPath)
	require.NoError(t, err)
	assert.Equal(t, 1.0, result["a"])

	// Test .TOML (uppercase)
	tomlPath := filepath.Join(dir, "test.TOML")
	err = os.WriteFile(tomlPath, []byte("key = 'value'\n"), 0600)
	require.NoError(t, err)

	result, err = cl.Load(tomlPath)
	require.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

func TestLoader_GetSupportedFormats(t *testing.T) {
	cl := NewConfigLoader()
	formats := cl.GetSupportedFormats()

	expected := []string{".toml", ".yaml", ".yml", ".json"}
	assert.Equal(t, expected, formats)
}

func TestLoader_IsSupportedFormat(t *testing.T) {
	cl := NewConfigLoader()

	assert.True(t, cl.IsSupportedFormat("config.yaml"))
	assert.True(t, cl.IsSupportedFormat("config.YAML"))
	assert.True(t, cl.IsSupportedFormat("config.yml"))
	assert.True(t, cl.IsSupportedFormat("config.json"))
	assert.True(t, cl.IsSupportedFormat("config.toml"))
	assert.False(t, cl.IsSupportedFormat("config.txt"))
	assert.False(t, cl.IsSupportedFormat("config"))
	assert.False(t, cl.IsSupportedFormat(""))
}

func TestLoader_YML_Extension(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	original := map[string]any{
		"hello": "world",
	}

	err := cl.Save(original, path)
	require.NoError(t, err)

	loaded, err := cl.Load(path)
	require.NoError(t, err)
	assert.Equal(t, original, loaded)
}

func TestLoader_EmptyMapSaveLoad(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()

	for _, ext := range []string{".yaml", ".json", ".toml"} {
		path := filepath.Join(dir, "empty"+ext)

		err := cl.Save(map[string]any{}, path)
		require.NoError(t, err, "Save empty map to %s", ext)

		result, err := cl.Load(path)
		require.NoError(t, err, "Load empty %s", ext)
		assert.Equal(t, map[string]any{}, result, "Round-trip empty map for %s", ext)
	}
}

func TestLoader_SaveCreatesFile(t *testing.T) {
	cl := NewConfigLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "new.yaml")

	err := cl.Save(map[string]any{"a": "b"}, path)
	require.NoError(t, err)

	_, err = os.Stat(path)
	assert.NoError(t, err, "File should exist after Save")
}
