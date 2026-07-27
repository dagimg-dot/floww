package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// FLOWW_ART is the ASCII art logo displayed on --help and init.
const FLOWW_ART = `
  /$$$$$$  /$$                                      
 /$$__  $$| $$                                      
| $$  \__/| $$  /$$$$$$  /$$  /$$  /$$ /$$  /$$  /$$
| $$$$    | $$ /$$__  $$| $$ | $$ | $$| $$ | $$ | $$
| $$_/    | $$| $$  \ $$| $$ | $$ | $$| $$ | $$ | $$
| $$      | $$| $$  | $$| $$ | $$ | $$| $$ | $$ | $$
| $$      | $$|  $$$$$$/|  $$$$$/$$$$/|  $$$$$/$$$$/
|__/      |__/ \______/  \_____/\___/  \_____/\___/ 
`

// SetupLogging configures slog with the given level name (DEBUG, INFO, WARNING,
// ERROR).  The handler writes to stderr.  Unknown level names default to WARNING.
func SetupLogging(levelName string) {
	var level slog.Level
	switch strings.ToUpper(levelName) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARNING":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelWarn
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))

	if level == slog.LevelDebug {
		slog.Debug(fmt.Sprintf("Logging level set to %s", strings.ToUpper(levelName)))
	}
}
