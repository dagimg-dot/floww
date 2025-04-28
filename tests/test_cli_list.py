import pytest
from typer.testing import CliRunner

from floww.cli import app
from floww.config import ConfigManager


@pytest.fixture(autouse=True)
def set_xdg_config_home(tmp_path, monkeypatch):
    # Redirect config home to a temporary directory
    config_home = tmp_path / "config"
    monkeypatch.setenv("XDG_CONFIG_HOME", str(config_home))
    return config_home


def test_list_no_workflows(set_xdg_config_home):
    runner = CliRunner()
    result = runner.invoke(app, ["init"])
    assert result.exit_code == 0
    assert "Initialized config at" in result.stdout

    # Listing before any workflows
    result = runner.invoke(app, ["list"])
    assert result.exit_code == 0
    assert "No workflows found" in result.stdout


def test_list_with_workflows(set_xdg_config_home):
    cfg = ConfigManager()
    cfg.init()
    workflows_dir = cfg.workflows_dir
    # Create sample workflows
    (workflows_dir / "alpha.yaml").write_text("dummy: 1")
    (workflows_dir / "beta.yaml").write_text("dummy: 2")

    runner = CliRunner()
    result = runner.invoke(app, ["list"])
    assert result.exit_code == 0
    lines = result.stdout.strip().splitlines()
    assert "alpha" in lines
    assert "beta" in lines
