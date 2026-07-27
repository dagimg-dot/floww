package backends

import (
	"log/slog"
	"strconv"
	"strings"
)

// WmctrlBackend manages workspace switching via wmctrl.
type WmctrlBackend struct {
	wmctrlCmd string
}

// NewWmctrlBackend creates a new WmctrlBackend with the default wmctrl command.
func NewWmctrlBackend() *WmctrlBackend {
	return &WmctrlBackend{wmctrlCmd: "wmctrl"}
}

// Switch switches to the given workspace index.
func (b *WmctrlBackend) Switch(target int) bool {
	return runCmd(b.wmctrlCmd, "-s", strconv.Itoa(target))
}

// GetTotalWorkspaces returns the total number of workspaces reported by wmctrl -d.
func (b *WmctrlBackend) GetTotalWorkspaces() int {
	cmd := execCmd(b.wmctrlCmd, "-d")
	out, err := cmd.Output()
	if err != nil {
		slog.Warn("wmctrl -d failed",
			"cmd", b.wmctrlCmd,
			"error", err,
		)
		return 0
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// GetAppendBaseOffset returns the append base offset (max(0, total-1)).
func (b *WmctrlBackend) GetAppendBaseOffset() int {
	return AppendBaseOffset(b.GetTotalWorkspaces())
}
