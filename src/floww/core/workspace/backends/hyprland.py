import json
import logging
import subprocess

from floww.utils import run_command

from .base import WorkspaceBackend

logger = logging.getLogger(__name__)


class HyprlandBackend(WorkspaceBackend):
    def __init__(self, hyprctl_cmd: str = "hyprctl"):
        self._hyprctl_cmd = hyprctl_cmd

    def switch(self, target: int) -> bool:
        try:
            cmd = [self._hyprctl_cmd, "dispatch", "workspace", str(target)]
            return run_command(cmd)
        except Exception as e:
            logger.error("hyprctl workspace switch failed: %s", e)
            return False

    def get_total_workspaces(self) -> int:
        try:
            cmd = [self._hyprctl_cmd, "workspaces", "-j"]
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            workspaces = json.loads(result.stdout)
            if not workspaces:
                return 1
            return max(ws["id"] for ws in workspaces)
        except Exception as e:
            logger.debug("hyprctl workspaces query failed: %s", e)
            return 0
