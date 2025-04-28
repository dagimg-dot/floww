import logging

try:
    from ewmhlib import EwmhRoot

    EWMHLIB_AVAILABLE = True
except ImportError:
    EWMHLIB_AVAILABLE = False
    # Fallback or error handling if ewmhlib is crucial but missing,
    # or potentially try wmctrl as a backup (though we aim to replace it)
    # For now, we'll just note it's unavailable.

from .utils import run_command
from .errors import WorkspaceError  # Import custom error

logger = logging.getLogger(__name__)


class WorkspaceManager:
    """Handles workspace management, preferring EWMH library if available."""

    def __init__(self):
        # Check if EWMH is usable (requires X11, not Wayland)
        # A more robust check might involve trying to instantiate EwmhRoot
        # but for now, we rely on the import success.
        self.use_ewmh = EWMHLIB_AVAILABLE
        if not self.use_ewmh:
            logger.warning(
                "ewmhlib not found or failed to import. "
                "Workspace switching might not work correctly, especially under Wayland."
                " Falling back to wmctrl if available."
            )
            # Keep wmctrl path as a potential fallback if needed
            self.wmctrl_cmd = "wmctrl"
        else:
            # Attempt to instantiate early to check for runtime issues (e.g., Wayland)
            try:
                EwmhRoot()
            except Exception as e:
                logger.warning(
                    f"ewmhlib imported but failed to initialize (maybe Wayland?): {e}. Falling back to wmctrl."
                )
                self.use_ewmh = False
                self.wmctrl_cmd = "wmctrl"

    def switch(self, desktop_num: int) -> bool:
        """
        Switch to a given workspace number.
        Uses EWMH if available (requires X11), otherwise attempts wmctrl.
        Returns True on success, False otherwise.
        Raises:
            WorkspaceError: If switching fails using the preferred method.
        """
        if self.use_ewmh:
            try:
                ewmh = EwmhRoot()
                num_desktops = ewmh.getNumberOfDesktops()
                if 0 <= desktop_num < num_desktops:
                    ewmh.setCurrentDesktop(desktop_num)
                    logger.info(f"Switched to desktop {desktop_num} via EWMH.")
                    return True
                else:
                    # Raise specific error for invalid desktop number
                    raise WorkspaceError(
                        f"Invalid desktop number: {desktop_num}. Available desktops: 0-{num_desktops - 1}"
                    )
            except Exception as e:
                # Raise specific error for EWMH failure
                err_msg = (
                    f"Failed to switch desktop using EWMH: {e}. "
                    "Ensure X11 session and ewmhlib is functional."
                )
                raise WorkspaceError(err_msg) from e
        else:
            # Fallback to wmctrl if ewmhlib wasn't available or failed init
            logger.debug("Attempting workspace switch using wmctrl fallback.")
            if not self._switch_with_wmctrl(desktop_num):
                # Raise specific error if wmctrl fallback also fails
                raise WorkspaceError(
                    f"Failed to switch to desktop {desktop_num} using wmctrl fallback."
                )
            return True  # wmctrl succeeded

    def _switch_with_wmctrl(self, desktop_num: int) -> bool:
        """Internal helper to switch using wmctrl command. Returns success bool."""
        try:
            cmd = [self.wmctrl_cmd, "-s", str(desktop_num)]
            success = run_command(cmd)
            if success:
                logger.info(f"Switched to desktop {desktop_num} via wmctrl.")
            else:
                logger.warning(f"wmctrl command failed for desktop {desktop_num}.")
            return success
        except FileNotFoundError:
            logger.error(
                f"wmctrl command '{self.wmctrl_cmd}' not found. Cannot switch workspace. "
                "Consider installing wmctrl."
            )
            return False
        except Exception as e:
            logger.error(f"Error running wmctrl command: {e}")
            return False
