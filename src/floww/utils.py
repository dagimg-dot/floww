import subprocess
import logging


logger = logging.getLogger(__name__)


def run_command(command: list[str]) -> bool:
    """
    Run a command list via subprocess.run and return True if it succeeds, otherwise False.
    """
    try:
        subprocess.run(command, check=True, capture_output=True, text=True)
        logger.info(f"Successfully ran: {' '.join(command)}")
        return True
    except subprocess.CalledProcessError as e:
        logger.error(f"Error running command: {' '.join(command)}")
        logger.error(f"Stderr: {e.stderr}")
        return False
    except FileNotFoundError:
        logger.error(f"Error: Command not found - {command[0]}")
        return False
