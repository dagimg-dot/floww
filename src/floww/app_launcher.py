import subprocess
from typing import List, Dict, Any
import os
import logging

logger = logging.getLogger(__name__)


class AppLauncher:
    """Handles launching applications via subprocess.Popen."""

    def launch_app(self, app_config: Dict[str, Any]) -> bool:
        """
        Launch an application based on its configuration from a workflow.

        Args:
            app_config: Dictionary with app configuration (name, exec, args, type)

        Returns:
            bool: True if launch was successful, False otherwise
        """
        app_type = app_config.get("type", "binary")
        app_name = app_config.get("name", app_config["exec"])
        app_exec = app_config["exec"]
        app_args = app_config.get("args", [])

        # Convert any non-string args to strings
        app_args = [str(arg) for arg in app_args]

        app_args = [
            os.path.expanduser(arg) if isinstance(arg, str) and "~" in arg else arg
            for arg in app_args
        ]

        logger.info(f"Launching {app_name} ({app_type}) with {app_exec}")

        if app_type == "binary":
            return self._launch_binary(app_exec, app_args, app_name)
        elif app_type == "flatpak":
            return self._launch_flatpak(app_exec, app_args, app_name)
        elif app_type == "snap":
            return self._launch_snap(app_exec, app_args, app_name)
        else:
            logger.error(f"Unknown app type: {app_type}")
            return False

    def _launch_binary(self, executable: str, args: List[str], app_name: str) -> bool:
        """Launch a binary application with the given arguments."""
        executable = os.path.expanduser(executable)
        cmd = [executable] + args
        return self._launch_process(cmd, app_name)

    def _launch_flatpak(self, app_id: str, args: List[str], app_name: str) -> bool:
        """Launch a Flatpak application with the given arguments."""
        cmd = ["flatpak", "run"] + [app_id] + args
        return self._launch_process(cmd, app_name)

    def _launch_snap(self, snap_name: str, args: List[str], app_name: str) -> bool:
        """Launch a Snap application with the given arguments."""
        cmd = [snap_name] + args
        return self._launch_process(cmd, app_name)

    def _launch_process(self, cmd: List[str], app_name: str) -> bool:
        """Launch a process and handle errors."""
        try:
            # Detach the process from the parent process group
            if os.name == "posix":  # Linux/Unix systems
                # Create a new session and redirect output to /dev/null
                with open(os.devnull, "w") as devnull:
                    process = subprocess.Popen(
                        cmd, stdout=devnull, stderr=devnull, start_new_session=True
                    )

            logger.info(f"Launched: {app_name} (PID: {process.pid})")
            return True

        except FileNotFoundError:
            error_msg = f"Executable not found: {cmd[0]}"
            logger.error(error_msg)
            return False

        except PermissionError:
            error_msg = f"Permission denied when launching: {cmd[0]}"
            logger.error(error_msg)
            return False

        except Exception as e:
            error_msg = f"Error launching {app_name}: {e}"
            logger.error(error_msg)
            return False
