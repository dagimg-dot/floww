import subprocess

class AppLauncher:
    """Handles launching applications via subprocess.Popen."""

    def launch(self, executable: str, args: list[str] = None) -> bool:
        """Launch an application with optional arguments."""
        cmd = [executable] + (args or [])
        try:
            subprocess.Popen(cmd)
            print(f"Launched: {' '.join(cmd)}")
            return True
        except FileNotFoundError:
            print(f"Executable not found: {executable}")
            return False
        except Exception as e:
            print(f"Error launching {executable}: {e}")
            return False 