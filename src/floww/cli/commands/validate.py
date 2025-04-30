import typer
from typing import Optional

from floww.cli.helpers import (
    check_initialized,
    print_error,
    get_workflow_name,
)
from floww import ConfigManager, WorkflowNotFoundError, WorkflowSchemaError, ConfigError

cfg = ConfigManager()


def validate(
    name: Optional[str] = typer.Argument(
        None,
        autocompletion=lambda: cfg.list_workflow_names(),
        help="Workflow name to validate",
    ),
):
    """Validate a workflow's schema without applying it."""
    check_initialized()

    workflow_name = get_workflow_name(name, "validate", cfg)
    typer.echo(f"Validating workflow: {workflow_name}")

    try:
        cfg.load_workflow(workflow_name)
        typer.echo("✓ Workflow is valid")
    except (WorkflowNotFoundError, WorkflowSchemaError, ConfigError) as e:
        print_error(f"Validation failed: {e}")
        raise typer.Exit(1)
