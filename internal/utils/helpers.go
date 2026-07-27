package utils

import (
	"errors"
	"log/slog"
	"os/exec"
)

// RunCommand executes a command synchronously and returns true on success.
// On failure it logs the error and stderr (if available), then returns false.
func RunCommand(name string, args ...string) bool {
	cmd := exec.Command(name, args...) //nolint:gosec // Intentional run_command utility
	out, err := cmd.CombinedOutput()
	if err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			slog.Error("Command not found", "cmd", name)
		} else {
			slog.Error("Error running command",
				"cmd", name,
				"stderr", string(out),
				"error", err,
			)
		}
		return false
	}
	slog.Info("Successfully ran", "cmd", name)
	return true
}

// Notify sends a desktop notification via notify-send.
// If notify-send is not found on the system, it logs a warning and returns.
func Notify(message string) {
	_, err := exec.LookPath("notify-send")
	if err != nil {
		slog.Warn("notify-send not found, cannot notify user")
		return
	}
	//nolint:gosec // user-provided message is intentional input
	cmd := exec.Command("notify-send", "--app-name", "Floww", message)
	if err := cmd.Run(); err != nil {
		slog.Error("Failed to send notification", "error", err)
	}
}
