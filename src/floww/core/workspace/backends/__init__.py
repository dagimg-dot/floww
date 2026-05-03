from .base import WorkspaceBackend
from .ewmh import EwmhBackend
from .hyprland import HyprlandBackend
from .niri import NiriBackend
from .wmctrl import WmctrlBackend

__all__ = [
    "WorkspaceBackend",
    "EwmhBackend",
    "HyprlandBackend",
    "NiriBackend",
    "WmctrlBackend",
]
