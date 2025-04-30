import pytest
from typer.testing import CliRunner

from floww.cli import app
from floww import ConfigManager


@pytest.fixture(autouse=True)
def reset_singleton():
    """Reset the ConfigManager singleton between tests."""
    original_instance = ConfigManager.instance
    ConfigManager.instance = None
    yield
    ConfigManager.instance = original_instance


@pytest.fixture(autouse=True)
def set_xdg_config_home(tmp_path, monkeypatch):
    # Redirect config home to a temporary directory
    config_home = tmp_path / "config"
    monkeypatch.setenv("XDG_CONFIG_HOME", str(config_home))
    return config_home


def test_list_no_workflows(set_xdg_config_home):
    print(f"\nTest files will be created in: {set_xdg_config_home}")
    runner = CliRunner()
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0
    assert "Initialized config at" in result.stdout

    # Show the contents of the config directory
    config_dir = set_xdg_config_home / "floww"
    print(f"Contents of {config_dir}:")
    for path in config_dir.rglob("*"):
        print(f"  {path.relative_to(config_dir)}")

    # Listing before any workflows
    result = runner.invoke(
        app, ["list"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0
    assert "No workflows found" in result.stdout


def test_list_with_workflows(set_xdg_config_home):
    print(f"\nTest files will be created in: {set_xdg_config_home}")
    cfg = ConfigManager()
    cfg.init()
    workflows_dir = cfg.workflows_dir
    # Create sample workflows with proper content
    (workflows_dir / "alpha.yaml").write_text("""
description: "Alpha workflow"
workspaces:
  - target: 1
    apps:
      - name: "Terminal"
        exec: "gnome-terminal"
""")
    (workflows_dir / "beta.yaml").write_text("""
description: "Beta workflow"
workspaces:
  - target: 2
    apps:
      - name: "Browser"
        exec: "firefox"
""")

    print(f"Contents of {workflows_dir}:")
    for path in workflows_dir.rglob("*"):
        print(f"  {path.relative_to(workflows_dir)}")

    runner = CliRunner()
    result = runner.invoke(
        app, ["list"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0
    lines = result.stdout.strip().splitlines()
    assert "  - alpha" in lines
    assert "  - beta" in lines
