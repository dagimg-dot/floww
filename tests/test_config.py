from pathlib import Path
import yaml

from floww.config import ConfigManager


def test_init_creates_config_and_workflows(tmp_path):
    # Use a temporary config path
    if not tmp_path:
        # root directory
        base_dir = Path("/") / "floww"
    else:
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
    assert content.get("workspaces") == {}, "Default workspaces should be an empty dict"


def test_list_workflows(tmp_path):
    base_dir = tmp_path / "floww"
    config_mgr = ConfigManager(config_path=base_dir)
    config_mgr.init()

    # No workflow files initially
    assert config_mgr.list_workflows() == []

    # Create sample workflow files
    workflows_dir = base_dir / "workflows"
    (workflows_dir / "alpha.yaml").write_text("dummy: 1")
    (workflows_dir / "beta.yaml").write_text("dummy: 2")
    # Non-YAML files should be ignored
    (workflows_dir / "ignore.txt").write_text("text")

    names = config_mgr.list_workflows()
    assert sorted(names) == ["alpha", "beta"]


def test_load_workflow(tmp_path):
    base_dir = tmp_path / "floww"
    config_mgr = ConfigManager(config_path=base_dir)
    config_mgr.init()

    # Write a workflow
    workflows_dir = base_dir / "workflows"
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
