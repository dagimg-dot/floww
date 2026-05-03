from .backends import WorkspaceBackend
from .factory import create_backend


class WorkspaceManager:
    """Delegates workspace operations to a compositor-specific backend."""

    def __init__(self, backend: WorkspaceBackend | None = None):
        self._backend = backend if backend is not None else create_backend()

    def switch(self, target: int) -> bool:
        """Switch to workspace ``target``. Returns True on success, False otherwise."""
        return self._backend.switch(int(target))

    def get_total_workspaces(self) -> int:
        return self._backend.get_total_workspaces()

    def get_append_base_offset(self) -> int:
        return self._backend.get_append_base_offset()
