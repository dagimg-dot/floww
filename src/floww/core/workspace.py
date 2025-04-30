# Workspace manager uses ewmhlib as the main tool for switching and getting info about workspaces
# The documentation says it only works on X11 but its working for me on wayland too.
# But as a backup I am using wmctrl which works on wayland

import logging

try:
    from ewmhlib import EwmhRoot

    EWMHLIB_AVAILABLE = True
except ImportError:
    EWMHLIB_AVAILABLE = False


from floww.utils import run_command
from .errors import WorkspaceError

logger = logging.getLogger(__name__)


class WorkspaceManager:
    """Handles workspace management, preferring EWMH library if available or use wmctrl as a fallback"""

    def __init__(self):
        self.use_ewmh = EWMHLIB_AVAILABLE

        if not self.use_ewmh:
            logger.warning(
                "ewmhlib not found or failed to import. "
                "Workspace switching might not work correctly, especially under Wayland."
                " Falling back to wmctrl if available."
            )

            self.wmctrl_cmd = "wmctrl"
        else:
            try:
                ewmh = EwmhRoot()
                self.ewmh = ewmh
            except Exception as e:
                logger.warning(
                    f"ewmhlib imported but failed to initialize (maybe Wayland?): {e}. Falling back to wmctrl."
                )

                self.use_ewmh = False
                self.wmctrl_cmd = "wmctrl"

    def switch(self, desktop_num: int) -> bool:
        """
        Switch to a given workspace number.
        Uses EWMH if available, otherwise attempts wmctrl.
        Returns True on success, False otherwise.\n
        Raises:
            WorkspaceError: If switching fails using the preferred method.
        """
        if self.use_ewmh:
            try:
                num_desktops = self.ewmh.getNumberOfDesktops()

                if 0 <= desktop_num < num_desktops:
                    self.ewmh.setCurrentDesktop(desktop_num)
                    logger.info(f"Switched to desktop {desktop_num} via EWMH.")
                    return True
                else:
                    raise WorkspaceError(
                        f"Invalid desktop number: {desktop_num}. Available desktops: 0-{num_desktops - 1}"
                    )
            except Exception as e:
                err_msg = f"Failed to switch desktop using EWMH: {e}. "
                raise WorkspaceError(err_msg) from e
        else:
            logger.debug("Attempting workspace switch using wmctrl fallback.")
            if not self._switch_with_wmctrl(desktop_num):
                raise WorkspaceError(
                    f"Failed to switch to desktop {desktop_num} using wmctrl fallback."
                )
            return True

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
