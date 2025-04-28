import typer
from . import __version__ as VERSION
from .config import ConfigManager

app = typer.Typer(help="floww CLI tool")


@app.command()
def init():
    """Initialize configuration directory and workflows folder."""
    cfg = ConfigManager()
    cfg.init()
    typer.echo(f"Initialized config at {cfg.config_path}")


@app.command(name="list")
def list_workflows():
    """List available workflows."""
    cfg = ConfigManager()
    names = cfg.list_workflows()
    if not names:
        typer.echo("No workflows found")
    else:
        for name in names:
            typer.echo(name)


@app.command()
def apply(name: str = typer.Argument(None, help="Workflow name to apply")):
    """Apply the named workflow (or interactive chooser)."""
    typer.echo(f"[stub] apply command called with name={name}")


@app.callback(invoke_without_command=True)
def version(
    ctx: typer.Context,
    version: bool = typer.Option(False, "--version", "-v", help="Show version"),
):
    """Show version and exit."""
    if version:
        typer.echo(f"floww version {VERSION}")
        raise typer.Exit()
