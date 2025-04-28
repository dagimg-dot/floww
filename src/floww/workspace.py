from .utils import run_command


class WorkspaceManager:
    """Handles workspace management via wmctrl."""

    def __init__(self):
        self.wmctrl = "wmctrl"

    def switch(self, desktop_num: int) -> bool:
        """
        Switch to a given workspace number.
        Returns True on success, False otherwise.
        """
        cmd = [self.wmctrl, "-s", str(desktop_num)]
        return run_command(cmd)
