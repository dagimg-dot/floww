import typer
import questionary
import logging
from typing import Optional
import yaml
import sys
import os
import subprocess
from pathlib import Path

from . import __version__ as VERSION
from .config import ConfigManager
from .errors import (
    ConfigError,
    WorkflowNotFoundError,
    WorkflowSchemaError,
    WorkspaceError,
)
from .workflow_manager import WorkflowManager
from .utils import run_command

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
)

app = typer.Typer(
    help="floww - your workflow automations in one place",
    context_settings={"help_option_names": ["-h", "--help"]},
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
    typer.echo(
        typer.style("Error:", fg=typer.colors.RED, bold=True) + f" {msg}", err=True
    )


def check_initialized():
    """Check if floww is initialized and raise error if not."""
    cfg = ConfigManager()
    if not cfg.is_initialized():
        print_error("floww is not initialized. Please run 'floww init' first.")
        raise typer.Exit(1)


@app.command()
def init(
    create_example: bool = typer.Option(
        False, "--example", "-e", help="Create an example workflow"
    ),
):
    """Initialize configuration directory and workflows folder."""
    try:
        cfg = ConfigManager()
        cfg.init(create_example=create_example)
        typer.echo(f"Initialized config at {cfg.config_path}")
    except ConfigError as e:
        print_error(f"Initialization failed: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred during initialization.")
        logger.exception(f"Unexpected init error: {e}")
        raise typer.Exit(1)


@app.command(name="list")
def list_workflows():
    """List available workflows."""
    check_initialized()
    try:
        cfg = ConfigManager()
        names = cfg.list_workflow_names()
        if not names:
            typer.echo("No workflows found")
        else:
            typer.echo("Available workflows:")
            for name in names:
                typer.echo(f"  - {name}")
    except ConfigError as e:
        print_error(f"Failed to list workflows: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred while listing workflows.")
        logger.exception(f"Unexpected list error: {e}")
        raise typer.Exit(1)


@app.command()
def validate(name: str = typer.Argument(..., help="Workflow name to validate")):
    """Validate a workflow's schema without applying it."""
    check_initialized()
    cfg = ConfigManager()

    try:
        typer.echo(f"Validating workflow: {name}")
        workflow_file = cfg.workflows_dir / f"{name}.yaml"

        if not workflow_file.is_file():
            typer.echo(f"Error: Workflow '{name}' not found at: {workflow_file}")
            raise typer.Exit(1)

        try:
            with open(workflow_file, "r") as f:
                workflow_data = yaml.safe_load(f)
        except yaml.YAMLError as e:
            typer.echo(f"Error: Invalid YAML format in {workflow_file}: {e}")
            raise typer.Exit(1)
        except Exception as e:
            typer.echo(f"Error: Could not read workflow file {workflow_file}: {e}")
            raise typer.Exit(1)

        cfg.validate_workflow(name, workflow_data)
        typer.echo("✓ Workflow is valid")

    except WorkflowNotFoundError as e:
        print_error(f"{e}")
        raise typer.Exit(1)
    except (WorkflowSchemaError, ConfigError) as e:
        print_error(f"Validation failed: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred during validation.")
        logger.exception(f"Unexpected validation error for '{name}': {e}")
        raise typer.Exit(1)


@app.command()
def apply(name: Optional[str] = typer.Argument(None, help="Workflow name to apply")):
    """Apply the named workflow (or interactive chooser if no name provided)."""
    check_initialized()
    cfg = ConfigManager()
    workflow_mgr = WorkflowManager()

    try:
        # Get the workflow name [argument or interactive chooser]
        workflow_name = name
        if not workflow_name:
            available = cfg.list_workflow_names()
            if not available:
                typer.echo("No workflows found to apply.")
                raise typer.Exit(0)

            workflow_name = questionary.select(
                "Select a workflow to apply:",
                choices=available,
                use_arrow_keys=True,
                use_shortcuts=True,
                qmark="❯",
            ).ask()

            if not workflow_name:
                typer.echo("No workflow selected.")
                raise typer.Exit(0)

        logger.info(f"Loading workflow: {workflow_name}")
        workflow_data = cfg.load_workflow(workflow_name)

        logger.info(f"Applying workflow: {workflow_name}")
        workflow_mgr.apply(workflow_data)

    except (WorkflowNotFoundError, WorkflowSchemaError, ConfigError) as e:
        print_error(f"Failed to load workflow '{workflow_name or 'selected'}': {e}")
        raise typer.Exit(1)
    except WorkspaceError as e:
        print_error(f"Workflow '{workflow_name}' failed: {e}")
        raise typer.Exit(1)
    except questionary.ValidationError as e:
        print_error(f"Interactive selection failed: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error(
            f"An unexpected error occurred applying workflow '{workflow_name}'."
        )
        logger.exception(f"Unexpected apply error: {e}")
        raise typer.Exit(1)


def open_in_editor(file_path: Path):
    """Open a file in the default editor."""
    editor = os.environ.get("EDITOR", "")
    if not editor:
        # Try common editors
        for ed in ["vim", "nano", "vi"]:
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


@app.command()
def add(
    name: str = typer.Argument(..., help="Name for the new workflow"),
    edit: bool = typer.Option(
        False, "--edit", "-e", help="Open the workflow in editor after creation"
    ),
):
    """Create a new workflow file with basic structure."""
    check_initialized()
    cfg = ConfigManager()

    if "/" in name or "\\" in name:
        print_error("Workflow name cannot contain path separators")
        raise typer.Exit(1)

    if name.startswith("."):
        print_error("Workflow name cannot start with a dot")
        raise typer.Exit(1)

    if name.endswith(".yaml"):
        print_error("Please provide the name without .yaml extension")
        raise typer.Exit(1)

    workflow_file = cfg.workflows_dir / f"{name}.yaml"
    if workflow_file.exists():
        print_error(f"Workflow '{name}' already exists")
        raise typer.Exit(1)

    workflow_content = {
        "description": "A new workflow.",
        "workspaces": [
            {
                "target": 0,
                "apps": [{"name": "Example App", "exec": "app-name", "type": "binary"}],
            }
        ],
    }

    try:
        with open(workflow_file, "w") as f:
            yaml.dump(workflow_content, f, default_flow_style=False, sort_keys=False)
        typer.echo(f"Created new workflow: {name}")

        if edit:
            typer.echo("Opening workflow in editor...")
            open_in_editor(workflow_file)

    except OSError as e:
        print_error(f"Failed to create workflow file: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred while creating the workflow.")
        logger.exception(f"Unexpected error creating workflow '{name}': {e}")
        raise typer.Exit(1)


@app.command()
def edit(
    name: Optional[str] = typer.Argument(None, help="Name of the workflow to edit"),
):
    """Open a workflow file in the default editor."""
    check_initialized()
    cfg = ConfigManager()

    # Get the workflow name [argument or interactive chooser]
    workflow_name = name
    if not workflow_name:
        available = cfg.list_workflow_names()
        if not available:
            print_error("No workflows found to edit")
            raise typer.Exit(1)

        workflow_name = questionary.select(
            "Select a workflow to edit:",
            choices=available,
            use_arrow_keys=True,
            use_shortcuts=True,
            qmark="❯",
        ).ask()

        if not workflow_name:
            typer.echo("No workflow selected")
            raise typer.Exit(0)

    workflow_file = cfg.workflows_dir / f"{workflow_name}.yaml"
    if not workflow_file.is_file():
        print_error(f"Workflow '{workflow_name}' not found")
        raise typer.Exit(1)

    try:
        typer.echo(f"Opening workflow '{workflow_name}' in editor...")
        open_in_editor(workflow_file)
    except Exception as e:
        print_error("An unexpected error occurred while opening the editor.")
        logger.exception(f"Unexpected error editing workflow '{workflow_name}': {e}")
        raise typer.Exit(1)


@app.command()
def remove(
    name: str = typer.Argument(..., help="Name of the workflow to remove"),
    force: bool = typer.Option(False, "--force", "-f", help="Skip confirmation prompt"),
):
    """Remove a workflow file."""
    check_initialized()
    cfg = ConfigManager()

    workflow_file = cfg.workflows_dir / f"{name}.yaml"
    if not workflow_file.is_file():
        print_error(f"Workflow '{name}' not found")
        raise typer.Exit(1)

    if not force:
        confirm = questionary.confirm(
            f"Are you sure you want to remove workflow '{name}'?", default=False
        ).ask()

        if not confirm:
            typer.echo("Operation cancelled")
            raise typer.Exit(0)

    try:
        workflow_file.unlink()
        typer.echo(f"Removed workflow: {name}")
    except OSError as e:
        print_error(f"Failed to remove workflow file: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred while removing the workflow.")
        logger.exception(f"Unexpected error removing workflow '{name}': {e}")
        raise typer.Exit(1)


@app.callback(invoke_without_command=True)
def main_callback(
    ctx: typer.Context,
    log_level: str = LogLevel,
    version: bool = typer.Option(
        False, "--version", "-v", help="Show version", is_eager=True
    ),
):
    """Main entry point, sets up logging and handles version/help."""
    setup_logging(log_level)

    if version:
        typer.echo(f"floww version {VERSION}")
        raise typer.Exit()
    if ctx.invoked_subcommand is None:
        typer.echo(FLOWW_ART)
        typer.echo(ctx.get_help())
        raise typer.Exit()
