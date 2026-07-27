package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionDisplay_Plain(t *testing.T) {
	// Clear all env vars that influence version display
	t.Setenv("FLOWW_VERSION_SUFFIX", "")
	t.Setenv("FLOWW_DEV", "")
	t.Setenv("ENV", "")

	// Chdir to temp dir to avoid picking up the project's own .env file
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()

	dir := t.TempDir()
	err = os.Chdir(dir)
	require.NoError(t, err)

	v := VersionDisplay()
	assert.Equal(t, Version, v)
}

func TestVersionDisplay_FLOWW_VERSION_SUFFIX(t *testing.T) {
	t.Setenv("FLOWW_VERSION_SUFFIX", "@custom")
	v := VersionDisplay()
	assert.Equal(t, "0.4.0@custom", v)
}

func TestVersionDisplay_ENVEnvVar(t *testing.T) {
	t.Setenv("FLOWW_VERSION_SUFFIX", "")
	t.Setenv("FLOWW_DEV", "")
	t.Setenv("ENV", "dev")
	v := VersionDisplay()
	assert.Equal(t, "0.4.0@dev", v)
}

func TestVersionDisplay_FLOWW_VERSION_SUFFIX_TakesPrecedence(t *testing.T) {
	// FLOWW_VERSION_SUFFIX has highest priority, even when ENV is also set
	t.Setenv("FLOWW_VERSION_SUFFIX", "@rc1")
	t.Setenv("ENV", "dev")
	v := VersionDisplay()
	assert.Equal(t, "0.4.0@rc1", v)
}

func TestVersionDisplay_DotenvFile(t *testing.T) {
	t.Setenv("FLOWW_VERSION_SUFFIX", "")
	t.Setenv("FLOWW_DEV", "")
	t.Setenv("ENV", "")

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origDir)
	}()

	dir := t.TempDir()
	err = os.WriteFile(filepath.Join(dir, ".env"), []byte("ENV=dev\n"), 0600)
	require.NoError(t, err)

	err = os.Chdir(dir)
	require.NoError(t, err)

	v := VersionDisplay()
	assert.Equal(t, "0.4.0@dev", v)
}

func TestVersionDisplay_DotenvFileStripsLeadingAt(t *testing.T) {
	t.Setenv("FLOWW_VERSION_SUFFIX", "")
	t.Setenv("FLOWW_DEV", "")
	t.Setenv("ENV", "")

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origDir)
	}()

	dir := t.TempDir()
	err = os.WriteFile(filepath.Join(dir, ".env"), []byte("ENV=@beta\n"), 0600)
	require.NoError(t, err)

	err = os.Chdir(dir)
	require.NoError(t, err)

	v := VersionDisplay()
	assert.Equal(t, "0.4.0@beta", v)
}

func TestVersionDisplay_NoDotenvFileReturnsPlain(t *testing.T) {
	t.Setenv("FLOWW_VERSION_SUFFIX", "")
	t.Setenv("FLOWW_DEV", "")
	t.Setenv("ENV", "")

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origDir)
	}()

	dir := t.TempDir()
	err = os.Chdir(dir)
	require.NoError(t, err)

	v := VersionDisplay()
	assert.Equal(t, Version, v)
}

func TestIterDotenvSearchDirs_IncludesCWD(t *testing.T) {
	dirs := IterDotenvSearchDirs("/a/b/c")
	assert.Contains(t, dirs, "/a/b/c")
}

func TestIterDotenvSearchDirs_WalksUp(t *testing.T) {
	dirs := IterDotenvSearchDirs("/a/b/c")
	assert.Contains(t, dirs, "/a/b")
	assert.Contains(t, dirs, "/a")
}

func TestIterDotenvSearchDirs_Deduplicates(t *testing.T) {
	dirs := IterDotenvSearchDirs("/")
	// / resolved to abs / — only one entry for root
	assert.Len(t, dirs, 1)
	assert.Equal(t, "/", dirs[0])
}

func TestIterDotenvSearchDirs_Limit(t *testing.T) {
	dirs := IterDotenvSearchDirs("/a/b/c/d/e/f/g/h/i/j/k")
	// Up to 8 levels + the starting dir = max 9
	assert.LessOrEqual(t, len(dirs), 9)
}

func TestFirstDotenvPath_Found(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, ".env"), []byte("KEY=val\n"), 0600)
	require.NoError(t, err)

	pathPtr := FirstDotenvPath(dir)
	require.NotNil(t, pathPtr)
	assert.Equal(t, filepath.Join(dir, ".env"), *pathPtr)
}

func TestFirstDotenvPath_NotFound(t *testing.T) {
	dir := t.TempDir()
	pathPtr := FirstDotenvPath(dir)
	assert.Nil(t, pathPtr)
}

func TestFirstDotenvPath_FindsNearest(t *testing.T) {
	// Create a .env in parent that should be shadowed by one in child
	parentDir := t.TempDir()
	childDir := filepath.Join(parentDir, "sub")
	err := os.Mkdir(childDir, 0750)
	require.NoError(t, err)

	// Write .env only in parent
	err = os.WriteFile(filepath.Join(parentDir, ".env"), []byte("KEY=from-parent\n"), 0600)
	require.NoError(t, err)

	pathPtr := FirstDotenvPath(childDir)
	require.NotNil(t, pathPtr)
	assert.Equal(t, filepath.Join(parentDir, ".env"), *pathPtr)
}

func TestReadEnvValueFromDotenv_Found(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, ".env"), []byte("MY_KEY=my_value\n"), 0600)
	require.NoError(t, err)

	val := ReadEnvValueFromDotenv("MY_KEY", dir)
	assert.Equal(t, "my_value", val)
}

func TestReadEnvValueFromDotenv_KeyNotFound(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, ".env"), []byte("OTHER=val\n"), 0600)
	require.NoError(t, err)

	val := ReadEnvValueFromDotenv("MISSING", dir)
	assert.Equal(t, "", val)
}

func TestReadEnvValueFromDotenv_NoFile(t *testing.T) {
	dir := t.TempDir()
	val := ReadEnvValueFromDotenv("ANY", dir)
	assert.Equal(t, "", val)
}

func TestReadEnvValueFromDotenv_SkipsComments(t *testing.T) {
	dir := t.TempDir()
	content := []byte("# this is a comment\nKEY=value\n")
	err := os.WriteFile(filepath.Join(dir, ".env"), content, 0600)
	require.NoError(t, err)

	val := ReadEnvValueFromDotenv("KEY", dir)
	assert.Equal(t, "value", val)
}

func TestReadEnvValueFromDotenv_StripWhitespace(t *testing.T) {
	dir := t.TempDir()
	content := []byte("  KEY  =  spaced-value  \n")
	err := os.WriteFile(filepath.Join(dir, ".env"), content, 0600)
	require.NoError(t, err)

	val := ReadEnvValueFromDotenv("KEY", dir)
	assert.Equal(t, "spaced-value", val)
}

func TestReadEnvValueFromDotenv_QuotedValues(t *testing.T) {
	dir := t.TempDir()
	content := []byte("KEY='quoted-value'\nOTHER=\"double-quoted\"\n")
	err := os.WriteFile(filepath.Join(dir, ".env"), content, 0600)
	require.NoError(t, err)

	assert.Equal(t, "quoted-value", ReadEnvValueFromDotenv("KEY", dir))
	assert.Equal(t, "double-quoted", ReadEnvValueFromDotenv("OTHER", dir))
}

func TestReadEnvValueFromDotenv_EmptyLine(t *testing.T) {
	dir := t.TempDir()
	content := []byte("\n\nKEY=val\n\n")
	err := os.WriteFile(filepath.Join(dir, ".env"), content, 0600)
	require.NoError(t, err)

	val := ReadEnvValueFromDotenv("KEY", dir)
	assert.Equal(t, "val", val)
}

func TestReadEnvValueFromDotenv_NoEqualsSign(t *testing.T) {
	dir := t.TempDir()
	// Lines without = should be skipped
	content := []byte("JUST_WORDS\nKEY=val\n")
	err := os.WriteFile(filepath.Join(dir, ".env"), content, 0600)
	require.NoError(t, err)

	val := ReadEnvValueFromDotenv("KEY", dir)
	assert.Equal(t, "val", val)
}
