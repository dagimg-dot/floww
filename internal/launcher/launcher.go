// Package launcher provides application launching for binary, flatpak, and snap types.
package launcher

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/dagimg-dot/floww/internal/config"
	"github.com/dagimg-dot/floww/internal/workflow"
)

// AppLauncher launches applications of various types (binary, flatpak, snap)
// with subprocess.Popen-like semantics (detached, stdout/stderr to /dev/null).
//
// RunCommand can be overridden for testing; when nil exec.Command is used.
type AppLauncher struct {
	RunCommand func(name string, arg ...string) *exec.Cmd
}

// New creates a new AppLauncher with default settings.
func New() *AppLauncher {
	return &AppLauncher{}
}

func (l *AppLauncher) command(name string, arg ...string) *exec.Cmd {
	if l.RunCommand != nil {
		return l.RunCommand(name, arg...)
	}
	return exec.Command(name, arg...) //nolint:gosec // Intentional app launcher
}

// LaunchApp dispatches an app launch by type.
//   - binary: launches app.Exec directly with app.Args
//   - flatpak: runs "flatpak run <app.Exec> [app.Args...]"
//   - snap: launches app.Exec (snap name) directly with app.Args
//
// Returns (true, nil) on success.
// Returns (false, *config.AppLaunchError) when the executable is not found.
// Returns (false, nil) for all other launch errors.
func (l *AppLauncher) LaunchApp(app workflow.App) (bool, error) {
	appType := app.Type
	if appType == "" {
		appType = "binary"
	}

	var cmdArgs []string
	switch appType {
	case "flatpak":
		cmdArgs = append([]string{"flatpak", "run", app.Exec}, app.Args...)
	case "snap":
		cmdArgs = append([]string{app.Exec}, app.Args...)
	default: // "binary"
		cmdArgs = append([]string{app.Exec}, app.Args...)
	}

	return l.LaunchProcess(cmdArgs)
}

// LaunchProcess launches a command detached from the parent process,
// with stdout and stderr redirected to /dev/null.
//
// Tilde (~) is expanded in all arguments using the current user's home directory.
// The process is started with Start() (not Run()) so the call returns immediately
// without waiting for the child to exit.
//
// Returns (true, nil) on success.
// Returns (false, *config.AppLaunchError) when the executable is not found.
// Returns (false, nil) for all other launch errors.
func (l *AppLauncher) LaunchProcess(cmd []string) (bool, error) {
	if len(cmd) == 0 {
		return false, nil
	}

	// Expand ~ in all elements.
	expanded := make([]string, len(cmd))
	for i, arg := range cmd {
		expanded[i] = expandTilde(arg)
	}

	c := l.command(expanded[0], expanded[1:]...)

	// Redirect stdout/stderr to /dev/null (subprocess.Popen semantics).
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return false, nil
	}
	defer devNull.Close() //nolint:errcheck
	c.Stdout = devNull
	c.Stderr = devNull

	if err := c.Start(); err != nil {
		// Distinguish file-not-found from other launch errors.
		if errors.Is(err, exec.ErrNotFound) {
			return false, &config.AppLaunchError{
				FlowwError: config.FlowwError{
					Msg:   "application not found: " + expanded[0],
					Cause: err,
				},
			}
		}
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			return false, &config.AppLaunchError{
				FlowwError: config.FlowwError{
					Msg:   "application not found: " + expanded[0],
					Cause: err,
				},
			}
		}
		return false, nil
	}

	return true, nil
}

// expandTilde replaces a leading "~" or "~/" prefix with the current user's
// home directory. If the path does not start with "~" or the user lookup
// fails, the original path is returned unchanged.
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	usr, err := user.Current()
	if err != nil {
		return path
	}
	home := usr.HomeDir
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[1:])
}
