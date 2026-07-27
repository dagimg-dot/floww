package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Version is the current version of floww, matching Python __version__.
var Version = "0.5.0"

// IterDotenvSearchDirs walks up from cwd up to 8 parent directories,
// returning a deduplicated list of directories to search for a .env file.
// This intentionally only scans CWD parents (not package dir parents like Python does).
func IterDotenvSearchDirs(cwd string) []string {
	seen := make(map[string]bool)
	var ordered []string

	dir := cwd
	for i := 0; i <= 8; i++ {
		resolved, err := filepath.Abs(dir)
		if err != nil {
			resolved = dir
		}
		if !seen[resolved] {
			seen[resolved] = true
			ordered = append(ordered, resolved)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return ordered
}

// FirstDotenvPath returns a pointer to the first .env file path found
// by searching IterDotenvSearchDirs, or nil if none is found.
func FirstDotenvPath(cwd string) *string {
	for _, d := range IterDotenvSearchDirs(cwd) {
		candidate := filepath.Join(d, ".env")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return &candidate
		}
	}
	return nil
}

// ReadEnvValueFromDotenv reads a .env file and returns the value for the given key.
// It parses .env manually (no external library) matching Python's behavior:
//   - Skips # comments
//   - Strips whitespace around key and value
//   - Handles ' and " quoted values
//   - Returns empty string if key not found or .env doesn't exist
func ReadEnvValueFromDotenv(key string, cwd string) string {
	pathPtr := FirstDotenvPath(cwd)
	if pathPtr == nil {
		return ""
	}

	f, err := os.Open(*pathPtr)
	if err != nil {
		return ""
	}
	defer f.Close() //nolint:errcheck

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		name = strings.TrimSpace(name)
		if name != key {
			continue
		}

		value = strings.TrimSpace(value)
		if len(value) >= 2 && value[0] == value[len(value)-1] && (value[0] == '\'' || value[0] == '"') {
			value = value[1 : len(value)-1]
		}
		return value
	}

	_ = scanner.Err()
	return ""
}

// VersionDisplay returns the version string with optional suffix.
// Priority:
//  1. FLOWW_VERSION_SUFFIX env var (highest priority)
//  2. ENV env var (from environment or .env file)
//  3. FLOWW_DEV=1 flag
//  4. Plain version
func VersionDisplay() string {
	// 1. FLOWW_VERSION_SUFFIX has highest priority
	if suffix := strings.TrimSpace(os.Getenv("FLOWW_VERSION_SUFFIX")); suffix != "" {
		return Version + suffix
	}

	// 2. ENV env var (from environment or .env file)
	envLabel := strings.TrimSpace(os.Getenv("ENV"))
	if envLabel == "" {
		if cwd, err := os.Getwd(); err == nil {
			envLabel = strings.TrimSpace(ReadEnvValueFromDotenv("ENV", cwd))
		}
	}
	if envLabel != "" {
		suffix := strings.TrimLeft(envLabel, "@")
		return Version + "@" + suffix
	}

	// 3. FLOWW_DEV=1 flag
	dev := strings.TrimSpace(os.Getenv("FLOWW_DEV"))
	if dev == "1" || strings.EqualFold(dev, "true") || strings.EqualFold(dev, "yes") || strings.EqualFold(dev, "on") {
		return Version + "@dev"
	}

	// 4. Plain version
	return Version
}
