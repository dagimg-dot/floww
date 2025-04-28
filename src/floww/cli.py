import typer
from . import __version__

app = typer.Typer(help="floww CLI tool")

@app.command()
def init():
    """Initialize configuration directory and workflows folder."""
    typer.echo("[stub] init command called")

@app.command(name="list")
def list_workflows():
    """List available workflows."""
    typer.echo("[stub] list command called")

@app.command()
def apply(name: str = typer.Argument(None, help="Workflow name to apply")):
    """Apply the named workflow (or interactive chooser)."""
    typer.echo(f"[stub] apply command called with name={name}")

@app.callback(invoke_without_command=True)
def version(ctx: typer.Context, version: bool = typer.Option(False, "--version", "-v", help="Show version")):
    """Show version and exit."""
    if version:
        typer.echo(f"floww version {__version__}")
        raise typer.Exit() 