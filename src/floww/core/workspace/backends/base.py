from abc import ABC, abstractmethod


class WorkspaceBackend(ABC):
    """Compositor-specific workspace switching and counts."""

    @abstractmethod
    def switch(self, target: int) -> bool:
        """Switch to workspace index/id used in workflow YAML for this backend."""

    @abstractmethod
    def get_total_workspaces(self) -> int:
        """Return a backend-specific notion of workspace count (used for display/limits)."""

    def get_append_base_offset(self) -> int:
        """Offset added to workflow targets when ``append`` is true."""
        return max(0, self.get_total_workspaces() - 1)
