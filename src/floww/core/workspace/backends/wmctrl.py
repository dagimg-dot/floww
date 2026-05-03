import logging
import subprocess

from floww.utils import run_command

from .base import WorkspaceBackend

logger = logging.getLogger(__name__)


class WmctrlBackend(WorkspaceBackend):
    def __init__(self, wmctrl_cmd: str = "wmctrl"):
        self._wmctrl_cmd = wmctrl_cmd

    def switch(self, target: int) -> bool:
        try:
            cmd = [self._wmctrl_cmd, "-s", str(target)]
            success = run_command(cmd)
            if success:
                logger.info("Switched to desktop %s via wmctrl.", target)
            else:
                logger.warning("wmctrl command failed for desktop %s.", target)
            return success
        except FileNotFoundError:
            logger.error(
                "wmctrl command '%s' not found. Consider installing wmctrl.",
                self._wmctrl_cmd,
            )
            return False
        except Exception as e:
            logger.error("Error running wmctrl command: %s", e)
            return False

    def get_total_workspaces(self) -> int:
        try:
            result = subprocess.run(
                [self._wmctrl_cmd, "-d"],
                capture_output=True,
                text=True,
                check=False,
            )
            if result.returncode != 0:
                logger.debug(
                    "wmctrl -d failed: %s",
                    (result.stderr or "").strip(),
                )
                return 0
            lines = [ln for ln in result.stdout.splitlines() if ln.strip()]
            return len(lines)
        except FileNotFoundError:
            logger.error("wmctrl command '%s' not found.", self._wmctrl_cmd)
            return 0
        except Exception as e:
            logger.error("Error getting total workspaces with wmctrl: %s", e)
            return 0
