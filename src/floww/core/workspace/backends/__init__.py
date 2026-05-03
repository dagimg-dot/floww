from .base import WorkspaceBackend
from .ewmh import EwmhBackend
from .hyprland import HyprlandBackend
from .wmctrl import WmctrlBackend

__all__ = [
    "WorkspaceBackend",
    "EwmhBackend",
    "HyprlandBackend",
    "WmctrlBackend",
]
