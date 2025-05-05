import typer
from typing import Optional
import questionary

from floww.cli.helpers import (
    check_initialized,
    print_error,
    logger,
    get_workflow_name,
)
from floww import (
    ConfigManager,
    WorkflowManager,
    WorkflowNotFoundError,
    WorkflowSchemaError,
    ConfigError,
    WorkspaceError,
)


def apply(
    name: Optional[str] = typer.Argument(
        None,
        autocompletion=lambda: ConfigManager().list_workflow_names(),
        help="Workflow name to apply",
    ),
    file_path: Optional[str] = typer.Option(
        None, "--file", "-f", help="Path to the workflow file to apply"
    ),
    append: bool = typer.Option(
        False,
        "--append",
        "-a",
        help="Append the workflow starting from the last workspace",
    ),
):
    """Apply the named workflow."""
    check_initialized()
    cfg = ConfigManager()
    workflow_mgr = WorkflowManager()
    workflow_name = None

    try:
        if file_path:
            workflow_name = file_path
            logger.info(f"Loading workflow from file: {file_path}")
            workflow_data = cfg.load_workflow(workflow_name, is_direct_load=True)
        else:
            workflow_name = get_workflow_name(name, "apply", cfg)
            logger.info(f"Loading workflow: {workflow_name}")
            workflow_data = cfg.load_workflow(workflow_name)

        logger.info(f"Applying workflow: {workflow_name}")
        workflow_mgr.apply(workflow_data, append)

    except (WorkflowNotFoundError, WorkflowSchemaError, ConfigError) as e:
        print_error(f"Failed to load workflow '{workflow_name or 'selected'}': {e}")
        raise typer.Exit(1)
    except WorkspaceError as e:
        print_error(f"Workflow '{workflow_name or 'unknown'}' failed: {e}")
        raise typer.Exit(1)
    except questionary.ValidationError as e:
        print_error(f"Interactive selection failed: {e}")
        raise typer.Exit(1)
    except typer.Exit as e:
        raise e
    except Exception as e:
        print_error(
            f"An unexpected error occurred applying workflow '{workflow_name or 'unknown'}' ."
        )
        logger.exception(f"Unexpected apply error: {e}")
        raise typer.Exit(1)
