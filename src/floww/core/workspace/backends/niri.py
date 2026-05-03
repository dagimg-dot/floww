import json
import logging
import subprocess

from floww.utils import run_command

from .base import WorkspaceBackend

logger = logging.getLogger(__name__)


class NiriBackend(WorkspaceBackend):
    """Niri via ``niri msg`` (dynamic workspaces on the focused output)."""

    def __init__(self, niri_cmd: str = "niri"):
        self._niri_cmd = niri_cmd

    @staticmethod
    def _workflow_target_to_idx(target: int) -> int:
        """
        Floww YAML ``target`` and niri ``idx`` / ``focus-workspace`` numeric refs are **1-based**
        (first workspace is ``1``). Values below ``1`` are clamped to ``1``.
        """
        return target if target >= 1 else 1

    def _fetch_workspaces_json(self) -> list | None:
        try:
            cmd = [self._niri_cmd, "msg", "-j", "workspaces"]
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                check=True,
            )
            data = json.loads(result.stdout)
            return data if isinstance(data, list) else None
        except Exception as e:
            logger.debug("niri workspaces fetch failed: %s", e)
            return None

    @staticmethod
    def _focused_workspace_idx(workspaces: list) -> int | None:
        for ws in workspaces:
            if not isinstance(ws, dict):
                continue
            if ws.get("is_focused") or ws.get("is_active"):
                idx = ws.get("idx")
                if isinstance(idx, int):
                    return idx
        return None

    def _is_already_focused_on_target(self, target: int) -> bool:
        """
        True if the focused workspace's **1-based** ``idx`` equals this workflow ``target``.

        Used to avoid ``focus-workspace`` when already there (e.g. ``workspace-auto-back-and-forth``).
        False if state cannot be determined.
        """
        workspaces = self._fetch_workspaces_json()
        if workspaces is None:
            return False
        desired = self._workflow_target_to_idx(target)
        cur = self._focused_workspace_idx(workspaces)
        return cur is not None and cur == desired

    def switch(self, target: int) -> bool:
        """Focus workspace ``target`` (1-based, same convention as niri ``idx``)."""
        try:
            if self._is_already_focused_on_target(target):
                logger.debug(
                    "niri: already on workspace idx %s; skipping focus-workspace",
                    self._workflow_target_to_idx(target),
                )
                return True

            idx = self._workflow_target_to_idx(target)
            cmd = [
                self._niri_cmd,
                "msg",
                "action",
                "focus-workspace",
                str(idx),
            ]
            return run_command(cmd)
        except Exception as e:
            logger.error("niri focus-workspace failed: %s", e)
            return False

    def get_total_workspaces(self) -> int:
        workspaces = self._fetch_workspaces_json()
        if workspaces is None:
            return 0
        if not workspaces:
            return 1
        return len(workspaces)
