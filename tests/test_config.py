import pytest
import yaml

from floww import ConfigManager


@pytest.fixture(autouse=True)
def reset_singleton():
    """Reset the ConfigManager singleton between tests."""
    # Store the original singleton instance
    original_instance = ConfigManager.instance
    ConfigManager.instance = None
    yield
    # Restore the original singleton instance
    ConfigManager.instance = original_instance


def test_init_creates_config_and_workflows(tmp_path):
    # Use a temporary config path
    base_dir = tmp_path / "floww"
    config_mgr = ConfigManager(config_path=base_dir)

    # Ensure directories do not exist initially
    assert not base_dir.exists()

    # Initialize configuration
    config_mgr.init()

    # Check that base directory, config file, and workflows subdir exist
    assert base_dir.exists() and base_dir.is_dir()
    config_file = base_dir / "config.yaml"
    workflows_dir = base_dir / "workflows"
    assert config_file.exists(), "config.yaml should be created"
    assert workflows_dir.exists() and workflows_dir.is_dir(), (
        "workflows directory should be created"
    )

    # Verify default YAML content
    content = yaml.safe_load(config_file.read_text())
    assert isinstance(content, dict)


def test_list_workflows(tmp_path):
    base_dir = tmp_path / "floww"
    config_mgr = ConfigManager(config_path=base_dir)
    config_mgr.init()

    # Create workflows directory if it doesn't exist
    workflows_dir = base_dir / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)

    # No workflow files initially
    assert config_mgr.list_workflow_names() == []

    # Create sample workflow files
    (workflows_dir / "alpha.yaml").write_text("dummy: 1")
    (workflows_dir / "beta.yaml").write_text("dummy: 2")
    # Non-YAML files should be ignored
    (workflows_dir / "ignore.txt").write_text("text")

    names = config_mgr.list_workflow_names()
    assert sorted(names) == ["alpha", "beta"]


def test_load_workflow(tmp_path):
    base_dir = tmp_path / "floww"
    config_mgr = ConfigManager(config_path=base_dir)
    config_mgr.init()

    # Create workflows directory if it doesn't exist
    workflows_dir = base_dir / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)

    # Write a workflow
    wf_path = workflows_dir / "demo.yaml"
    data = {
        "workspaces": [{"target": 0, "apps": [{"name": "Terminal", "exec": "xterm"}]}]
    }
    wf_path.write_text(yaml.dump(data))

    loaded = config_mgr.load_workflow("demo")
    # The loaded data will have default type field added
    expected = {
        "workspaces": [
            {
                "target": 0,
                "apps": [{"name": "Terminal", "exec": "xterm", "type": "binary"}],
            }
        ]
    }
    assert loaded == expected


def test_config_default_timing_values(tmp_path):
    """Test that default timing values are set correctly when not in config file."""
    config_dir = tmp_path / "config" / "floww"
    config_dir.mkdir(parents=True)
    config_file = config_dir / "config.yaml"

    # Create an empty config file
    config_file.write_text("# Empty config\n")

    # Initialize with the test config path (use config_path instead of config_root)
    config_mgr = ConfigManager(config_path=config_dir)

    # Check default timing values
    assert config_mgr.get_timing_config() == {
        "workspace_switch_wait": 3,
        "app_launch_wait": 1,
        "respect_app_wait": True,
    }


def test_config_custom_timing_values(tmp_path):
    """Test that custom timing values from config file are loaded correctly."""
    base_dir = tmp_path / "floww"
    base_dir.mkdir(parents=True, exist_ok=True)
    config_file = base_dir / "config.yaml"

    # Create a config with custom timing values
    config_file.write_text("""
timing:
  workspace_switch_wait: 0.5
  app_launch_wait: 0.2
  respect_app_wait: false
""")

    # Initialize with the test config path
    config_mgr = ConfigManager(config_path=base_dir)
    # Force reload of config
    config_mgr.config = config_mgr._load_and_merge_config()

    # Check custom timing values
    assert config_mgr.get_timing_config() == {
        "workspace_switch_wait": 0.5,
        "app_launch_wait": 0.2,
        "respect_app_wait": False,
    }


def test_config_partial_timing_override(tmp_path):
    """Test that partial timing config overrides work correctly."""
    config_dir = tmp_path / "config" / "floww"
    config_dir.mkdir(parents=True)
    config_file = config_dir / "config.yaml"

    # Create a config with partial timing values
    config_file.write_text("""
timing:
  workspace_switch_wait: 3
""")

    # Initialize with the test config path (use config_path instead of config_root)
    config_mgr = ConfigManager(config_path=config_dir)

    # Check that specified value is overridden but others remain default
    timing_config = config_mgr.get_timing_config()
    assert timing_config["workspace_switch_wait"] == 3
    assert timing_config["app_launch_wait"] == 1  # Default
    assert timing_config["respect_app_wait"] is True  # Default


def test_config_invalid_timing_values(tmp_path):
    """Test that invalid timing values are handled correctly."""
    config_dir = tmp_path / "config" / "floww"
    config_dir.mkdir(parents=True)
    config_file = config_dir / "config.yaml"

    # Create a config with invalid timing values
    config_file.write_text("""
timing:
  workspace_switch_wait: -1  # Negative should be replaced with default
  app_launch_wait: "invalid"  # Non-numeric should be replaced with default
""")

    # Initialize with the test config path (use config_path instead of config_root)
    config_mgr = ConfigManager(config_path=config_dir)

    # Check that invalid values are replaced with defaults
    timing_config = config_mgr.get_timing_config()
    assert timing_config["workspace_switch_wait"] == 3
    assert timing_config["app_launch_wait"] == 1
