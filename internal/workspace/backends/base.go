package backends

import (
	"os/exec"

	"github.com/dagimg-dot/floww/internal/utils"
)

// Overridable in tests — default to real implementations.
var runCmd = utils.RunCommand
var execCmd = exec.Command

// WorkspaceBackend defines compositor-specific workspace switching and counts.
type WorkspaceBackend interface {
	// Switch to workspace index/id used in workflow YAML for this backend.
	Switch(target int) bool

	// GetTotalWorkspaces returns a backend-specific notion of workspace count
	// (used for display/limits).
	GetTotalWorkspaces() int

	// GetAppendBaseOffset returns the offset added to workflow targets when
	// append is true. The default implementation returns max(0, total - 1).
	GetAppendBaseOffset() int
}

// AppendBaseOffset returns max(0, total-1), the default logic for
// GetAppendBaseOffset. Backends that embed the default behaviour can call
// this function rather than reimplementing it.
func AppendBaseOffset(total int) int {
	if total <= 1 {
		return 0
	}
	return total - 1
}
