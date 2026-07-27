package backends

import (
	"encoding/json"
	"fmt"
)

// HyprlandBackend manages workspace switching for the Hyprland compositor.
type HyprlandBackend struct {
	hyprctlCmd string
}

// NewHyprlandBackend creates a HyprlandBackend with hyprctl as the default command.
func NewHyprlandBackend() *HyprlandBackend {
	return &HyprlandBackend{
		hyprctlCmd: "hyprctl",
	}
}

// Switch switches to the given workspace using hyprctl dispatch workspace.
func (b *HyprlandBackend) Switch(target int) bool {
	return runCmd(b.hyprctlCmd, "dispatch", "workspace", fmt.Sprintf("%d", target))
}

// hyprctlWorkspace represents a single workspace from hyprctl workspaces -j JSON.
type hyprctlWorkspace struct {
	ID int `json:"id"`
}

// GetTotalWorkspaces returns the highest workspace ID from hyprctl workspaces -j.
// Returns 1 if no workspaces are reported, 0 on error.
func (b *HyprlandBackend) GetTotalWorkspaces() int {
	cmd := execCmd(b.hyprctlCmd, "workspaces", "-j")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	var workspaces []hyprctlWorkspace
	if err := json.Unmarshal(out, &workspaces); err != nil {
		return 0
	}

	if len(workspaces) == 0 {
		return 1
	}

	maxID := 0
	for _, ws := range workspaces {
		if ws.ID > maxID {
			maxID = ws.ID
		}
	}
	return maxID
}

// GetAppendBaseOffset returns the append base offset based on total workspaces.
func (b *HyprlandBackend) GetAppendBaseOffset() int {
	return AppendBaseOffset(b.GetTotalWorkspaces())
}
