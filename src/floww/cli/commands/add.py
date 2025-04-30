import typer

from floww.cli.helpers import (
    check_initialized,
    print_error,
    logger,
    open_in_editor,
)
from floww import ConfigManager, ConfigLoadError
from floww.utils import FileType


def add(
    name: str = typer.Argument(..., help="Name for the new workflow"),
    edit: bool = typer.Option(
        False, "--edit", "-e", help="Open the workflow in editor after creation"
    ),
    file_type: FileType = typer.Option(
        FileType.YAML, "--type", "-t", help="File format for the workflow file"
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

    if "." in name:
        print_error("Please provide the name without file extension")
        raise typer.Exit(1)

    existing_files = []
    for ext in cfg.config_loader.get_supported_formats():
        path = cfg.workflows_dir / f"{name}{ext}"
        if path.is_file():
            existing_files.append(path)

    if existing_files:
        extensions = ", ".join(f.suffix for f in existing_files)
        print_error(f"Workflow '{name}' already exists with extension: {extensions}")
        raise typer.Exit(1)

    ext = f".{file_type.value}"
    workflow_file = cfg.workflows_dir / f"{name}{ext}"

    workflow_content = {
        "description": "A new workflow.",
        "workspaces": [
            {
                "target": 0,
                "apps": [{"name": "Example App", "exec": "app-name", "type": "binary"}],
            }
        ],
        "final_workspace": 0,
    }

    try:
        cfg.config_loader.save(workflow_content, workflow_file)
        typer.echo(f"Created new workflow: {name} ({file_type.value})")

        if edit:
            typer.echo("Opening workflow in editor...")
            open_in_editor(workflow_file)

    except ConfigLoadError as e:
        print_error(f"Failed to create workflow file: {e}")
        raise typer.Exit(1)
    except OSError as e:
        print_error(f"Failed to create workflow file: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred while creating the workflow.")
        logger.exception(f"Unexpected error creating workflow '{name}': {e}")
        raise typer.Exit(1)
