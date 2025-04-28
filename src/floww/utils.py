import subprocess


def run_command(command: list[str]) -> bool:
    """
    Run a command list via subprocess.run and return True if it succeeds, otherwise False.
    """
    try:
        result = subprocess.run(command, check=True, capture_output=True, text=True)
        print(f"Successfully ran: {' '.join(command)}")
        return True
    except subprocess.CalledProcessError as e:
        print(f"Error running command: {' '.join(command)}")
        print(f"Stderr: {e.stderr}")
        return False
    except FileNotFoundError:
        print(f"Error: Command not found - {command[0]}")
        return False 