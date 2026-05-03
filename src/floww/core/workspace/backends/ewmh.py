import logging
from typing import ClassVar

from .base import WorkspaceBackend

logger = logging.getLogger(__name__)

_ewmh_root_cls = None
_ewmh_import_attempted: ClassVar[bool] = False


def _get_ewmh_root_class():
    global _ewmh_root_cls, _ewmh_import_attempted
    if _ewmh_import_attempted:
        return _ewmh_root_cls
    _ewmh_import_attempted = True
    try:
        from ewmhlib import EwmhRoot as _EwmhRoot

        _ewmh_root_cls = _EwmhRoot
    except Exception as e:
        logger.debug("ewmhlib not available: %s", e)
        _ewmh_root_cls = None
    return _ewmh_root_cls


class EwmhBackend(WorkspaceBackend):
    def __init__(self, ewmh):
        self._ewmh = ewmh

    @classmethod
    def try_create(cls) -> "EwmhBackend | None":
        EwmhRoot = _get_ewmh_root_class()
        if EwmhRoot is None:
            return None
        try:
            return cls(EwmhRoot())
        except Exception as e:
            logger.warning(
                "ewmhlib imported but failed to initialize (maybe Wayland?): %s",
                e,
            )
            return None

    def switch(self, target: int) -> bool:
        try:
            num_desktops = self._ewmh.getNumberOfDesktops()
            if 0 <= target < num_desktops:
                self._ewmh.setCurrentDesktop(target)
                logger.info("Switched to desktop %s via EWMH.", target)
                return True
            logger.warning(
                "Invalid desktop number: %s. Available: 0-%s",
                target,
                num_desktops - 1,
            )
            return False
        except Exception as e:
            logger.error("Failed to switch desktop using EWMH: %s", e)
            return False

    def get_total_workspaces(self) -> int:
        try:
            return self._ewmh.getNumberOfDesktops()
        except Exception as e:
            logger.error("EWMH getNumberOfDesktops failed: %s", e)
            return 0
