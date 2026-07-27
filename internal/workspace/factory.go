package workspace

import (
	"log/slog"
	"os"
	"strings"

	"github.com/dagimg-dot/floww/internal/config"
	"github.com/dagimg-dot/floww/internal/utils"
	"github.com/dagimg-dot/floww/internal/workspace/backends"
)

// CreateBackend builds a WorkspaceBackend from an explicit backend name or
// from config when backendName is empty.  It mirrors the detection and
// fallback logic of the Python factory.
func CreateBackend(backendName string, cfg *config.ConfigManager) backends.WorkspaceBackend {
	if backendName == "" {
		general := cfg.GetGeneralConfig()
		backendName = general.WorkspaceBackend
	}

	name := normalizeBackendName(backendName)

	switch name {
	case "auto":
		return detectAuto()
	case "hyprland":
		return backends.NewHyprlandBackend()
	case "niri":
		return backends.NewNiriBackend()
	case "ewmh":
		be, err := backends.TryCreate()
		if err == nil && be != nil {
			return be
		}
		slog.Warn("Explicit ewmh backend requested but unavailable; falling back to wmctrl",
			"backend", backendName)
		return backends.NewWmctrlBackend()
	case "wmctrl":
		return backends.NewWmctrlBackend()
	default:
		return detectAuto()
	}
}

// normalizeBackendName lowercases, trims, and validates the backend name.
// Invalid names are replaced with "auto".
func normalizeBackendName(raw string) string {
	name := strings.ToLower(strings.TrimSpace(raw))
	if !utils.ValidWorkspaceBackends[name] {
		slog.Warn("Unknown workspace_backend; valid options are auto, hyprland, niri, ewmh, wmctrl. Using 'auto'.",
			"backend", raw)
		return "auto"
	}
	return name
}

// isHyprlandSession returns true when the HYPRLAND_INSTANCE_SIGNATURE
// environment variable is set, indicating a Hyprland session.
func isHyprlandSession() bool {
	return os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != ""
}

// isNiriSession returns true when NIRI_SOCKET is set or
// XDG_CURRENT_DESKTOP equals "niri".
func isNiriSession() bool {
	if os.Getenv("NIRI_SOCKET") != "" {
		return true
	}
	return strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP")) == "niri"
}

// detectAuto probes the running session in order: Hyprland → Niri → EWMH → wmctrl.
func detectAuto() backends.WorkspaceBackend {
	if isHyprlandSession() {
		slog.Info("Hyprland detected, using hyprctl for workspace management.")
		return backends.NewHyprlandBackend()
	}

	if isNiriSession() {
		slog.Info("Niri detected, using niri msg for workspace management.")
		return backends.NewNiriBackend()
	}

	be, err := backends.TryCreate()
	if err == nil && be != nil {
		return be
	}

	slog.Warn("EWMH unavailable or failed to initialize; falling back to wmctrl.")
	return backends.NewWmctrlBackend()
}
