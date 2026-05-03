import logging
import os

from floww.core.config import ConfigManager
from floww.utils.constants import VALID_WORKSPACE_BACKENDS

from .backends import EwmhBackend, HyprlandBackend, WmctrlBackend, WorkspaceBackend

logger = logging.getLogger(__name__)


def _normalize_backend_name(raw: str) -> str:
    name = str(raw).lower().strip()
    if name not in VALID_WORKSPACE_BACKENDS:
        logger.warning(
            "Unknown workspace_backend %r; valid options are %s. Using 'auto'.",
            raw,
            ", ".join(sorted(VALID_WORKSPACE_BACKENDS)),
        )
        return "auto"
    return name


def _is_hyprland_session() -> bool:
    return os.environ.get("HYPRLAND_INSTANCE_SIGNATURE") is not None


def _detect_auto() -> WorkspaceBackend:
    if _is_hyprland_session():
        logger.info("Hyprland detected. Using hyprctl for workspace management.")
        return HyprlandBackend()

    ewmh = EwmhBackend.try_create()
    if ewmh is not None:
        return ewmh

    logger.warning(
        "ewmhlib unavailable or failed to initialize. "
        "Workspace switching may be limited under Wayland. Falling back to wmctrl if installed."
    )
    return WmctrlBackend()


def create_backend(workspace_backend: str | None = None) -> WorkspaceBackend:
    """
    Build a workspace backend from config (when ``workspace_backend`` is None)
    or from an explicit override used by tests.
    """
    if workspace_backend is None:
        cfg = ConfigManager().get_general_config()
        workspace_backend = cfg.get("workspace_backend", "auto")

    name = _normalize_backend_name(workspace_backend)

    if name == "auto":
        return _detect_auto()
    if name == "hyprland":
        return HyprlandBackend()
    if name == "ewmh":
        ewmh = EwmhBackend.try_create()
        if ewmh is not None:
            return ewmh
        logger.warning("Explicit ewmh backend requested but unavailable; using wmctrl.")
        return WmctrlBackend()
    if name == "wmctrl":
        return WmctrlBackend()

    return _detect_auto()
