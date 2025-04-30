import typer
import questionary
import logging
from typing import Optional
import sys
import os
import subprocess
from pathlib import Path

from floww import ConfigManager
from floww.utils import run_command

FLOWW_ART = r"""

  /$$$$$$  /$$                                      
 /$$__  $$| $$                                      
| $$  \__/| $$  /$$$$$$  /$$  /$$  /$$ /$$  /$$  /$$
| $$$$    | $$ /$$__  $$| $$ | $$ | $$| $$ | $$ | $$
| $$_/    | $$| $$  \ $$| $$ | $$ | $$| $$ | $$ | $$
| $$      | $$| $$  | $$| $$ | $$ | $$| $$ | $$ | $$
| $$      | $$|  $$$$$$/|  $$$$$/$$$$/|  $$$$$/$$$$/
|__/      |__/ \______/  \_____/\___/  \_____/\___/ 
"""

logger = logging.getLogger("floww")

LogLevel = typer.Option(
    "WARNING",
    "--log-level",
    "-l",
    help="Set the logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL).",
    case_sensitive=False,
    show_choices=True,
)


def setup_logging(level_name: str):
    level = getattr(logging, level_name.upper(), logging.WARNING)
    formatter = logging.Formatter("%(levelname)s: %(name)s: %(message)s")
    handler = logging.StreamHandler(sys.stderr)
    handler.setFormatter(formatter)

    logging.basicConfig(level=level, handlers=[handler])
    logger.setLevel(level)
    if level <= logging.DEBUG:
        logger.debug(f"Logging level set to {level_name.upper()}")


def print_error(msg: str):
    """Print an error message to the console."""
    typer.echo(
        typer.style("Error:", fg=typer.colors.RED, bold=True) + f" {msg}", err=True
    )


def check_initialized():
    """Check if floww is initialized and raise error if not."""
    cfg = ConfigManager()
    if not cfg.is_initialized():
        print_error("floww is not initialized. Please run 'floww init' first.")
        raise typer.Exit(1)


def get_workflow_name(
    name: Optional[str], action: str, cfg: ConfigManager
) -> Optional[str]:
    """Get workflow name from argument or interactive selection.

    Args:
        name: Optional workflow name from command argument
        action: Action being performed (for messages)
        cfg: ConfigManager instance

    Returns:
        Selected workflow name or None if selection was cancelled
    """
    workflow_name = name
    if not workflow_name:
        available = cfg.list_workflow_names()
        if not available:
            print_error(f"No workflows found to {action}")
            raise typer.Exit(1)

        workflow_name = questionary.select(
            f"Select a workflow to {action}:",
            choices=available,
            use_arrow_keys=True,
            use_shortcuts=True,
            qmark="❯",
        ).ask()

        if not workflow_name:
            typer.echo("No workflow selected")
            raise typer.Exit(0)

    return workflow_name


def open_in_editor(file_path: Path):
    """Open a file in the default editor."""
    editor = os.environ.get("EDITOR", "")
    if not editor:
        # Try common editors
        for ed in ["vim", "vi", "nano"]:
            if run_command(["which", ed]):
                editor = ed
                break

    if not editor:
        print_error(
            "No suitable editor found. Please set the EDITOR environment variable."
        )
        raise typer.Exit(1)

    try:
        subprocess.run([editor, str(file_path)], check=True)
    except subprocess.CalledProcessError as e:
        print_error(f"Editor exited with error: {e}")
        raise typer.Exit(1)
    except FileNotFoundError:
        print_error(f"Editor '{editor}' not found")
        raise typer.Exit(1)
