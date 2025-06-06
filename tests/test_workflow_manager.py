import pytest
from unittest.mock import patch, MagicMock
import time

from floww import WorkflowManager


class DummyWorkspaceManager:
    def __init__(self):
        self.calls = []

    def get_total_workspaces(self):
        return 5

    def switch(self, desktop_num):
        self.calls.append(desktop_num)
        return True


class DummyAppLauncher:
    def __init__(self):
        self.calls = []

    def launch_app(self, app_config):
        self.calls.append(app_config)
        return True


class DummyConfigManager:
    def get_timing_config(self):
        return {
            "workspace_switch_wait": 0,
            "app_launch_wait": 0,
            "respect_app_wait": False,
        }


@pytest.fixture
def workflow_manager():
    return WorkflowManager(show_notifications=False)


def test_apply_empty_workflow(workflow_manager):
    workflow_data = {"workspaces": []}

    success = workflow_manager.apply(workflow_data)
    assert success is True


def test_apply_workflow_with_description(workflow_manager):
    workflow_data = {"description": "Test workflow", "workspaces": []}

    with patch("floww.core.workflow.typer.echo") as mock_echo:
        success = workflow_manager.apply(workflow_data)
        assert success is True
        mock_echo.assert_any_call("Workflow: Test workflow")


def test_apply_workflow_switch_workspace_failure(workflow_manager):
    workflow_data = {"workspaces": [{"target": 1, "apps": []}]}

    # Mock the workspace manager to fail switching
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=False)

    success = workflow_manager.apply(workflow_data)
    assert success is False
    workflow_manager.workspace_mgr.switch.assert_called_once_with(1)


def test_apply_workflow_launch_apps(workflow_manager):
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1"},
                    {"name": "App2", "exec": "app2", "args": ["--flag"]},
                ],
            }
        ]
    }

    # Mock successful workspace switch
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    # Mock successful app launches
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    with patch("floww.core.workflow.typer.secho") as mock_secho:
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify workspace was switched
        workflow_manager.workspace_mgr.switch.assert_called_once_with(1)

        # Verify apps were launched
        assert workflow_manager.app_launcher.launch_app.call_count == 2
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App1", "exec": "app1"}
        )
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App2", "exec": "app2", "args": ["--flag"]}
        )

        # Verify success message with color
        mock_secho.assert_any_call("✓ Workflow applied successfully", fg="green")


def test_apply_workflow_app_launch_failure(workflow_manager):
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1"},
                    {"name": "App2", "exec": "app2"},
                ],
            }
        ]
    }

    # Mock successful workspace switch
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    # Mock app launch failure
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=False)

    with patch("floww.core.workflow.typer.secho") as mock_secho:
        # Patch time.sleep to avoid actual waits
        with patch("floww.core.workflow.time.sleep") as mock_sleep:
            success = workflow_manager.apply(workflow_data)
            assert success is False  # Should fail due to app launch failure

            # Verify workspace switch happened
            workflow_manager.workspace_mgr.switch.assert_called_once_with(1)

            # Verify app launch attempts
            assert workflow_manager.app_launcher.launch_app.call_count == 2

            # Verify error messages were displayed
            assert mock_secho.call_count >= 2
            args_list = [
                call[0][0]
                for call in mock_secho.call_args_list
                if isinstance(call[0][0], str)
            ]
            assert any("Failed to launch App1" in arg for arg in args_list)
            assert any("Failed to launch App2" in arg for arg in args_list)

            # Verify sleep was not called since app launches failed
            assert mock_sleep.call_count == 0


def test_apply_workflow_multiple_workspaces(workflow_manager):
    workflow_data = {
        "workspaces": [
            {"target": 1, "apps": [{"name": "App1", "exec": "app1"}]},
            {"target": 2, "apps": [{"name": "App2", "exec": "app2"}]},
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    success = workflow_manager.apply(workflow_data)
    assert success is True

    # Verify workspace switches
    assert workflow_manager.workspace_mgr.switch.call_count == 2
    workflow_manager.workspace_mgr.switch.assert_any_call(1)
    workflow_manager.workspace_mgr.switch.assert_any_call(2)

    # Verify app launches
    assert workflow_manager.app_launcher.launch_app.call_count == 2
    workflow_manager.app_launcher.launch_app.assert_any_call(
        {"name": "App1", "exec": "app1"}
    )
    workflow_manager.app_launcher.launch_app.assert_any_call(
        {"name": "App2", "exec": "app2"}
    )


def test_apply_workflow_with_wait(workflow_manager):
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1", "wait": 0.5},
                    {
                        "name": "App2",
                        "exec": "app2",
                        "wait": "invalid",
                    },  # Test invalid wait
                    {"name": "App3", "exec": "app3"},  # No wait after last app
                ],
            }
        ]
    }

    # Mock successful workspace switch and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Patch time.sleep and typer.echo/secho
    with (
        patch("floww.core.workflow.time.sleep") as mock_sleep,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True  # Should succeed even with wait issues

        # Verify workspace switch
        workflow_manager.workspace_mgr.switch.assert_called_once_with(1)

        # Verify app launches
        assert workflow_manager.app_launcher.launch_app.call_count == 3
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App1", "exec": "app1", "wait": 0.5}
        )
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App2", "exec": "app2", "wait": "invalid"}
        )
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App3", "exec": "app3"}
        )

        # Verify time.sleep was called correctly:
        # 1. After App1 (0.5s from app-specific wait)
        # Only App1's wait should be applied, no final wait since it's the only workspace
        assert mock_sleep.call_count == 1


def test_workflow_respects_workspace_switch_wait(workflow_manager):
    """Test that workflow manager respects workspace switch wait time."""
    workflow_data = {
        "workspaces": [
            {"target": 1, "apps": [{"name": "App1", "exec": "app1"}]},
            {"target": 2, "apps": [{"name": "App2", "exec": "app2"}]},
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 1.5,
        "app_launch_wait": 0,  # No wait after app launches
        "respect_app_wait": True,
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.core.workflow.time.sleep") as mock_sleep,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify workspace switches happened
        assert workflow_manager.workspace_mgr.switch.call_count == 2

        # Verify time.sleep was called after first workspace only
        # Only one sleep call for workspace_switch_wait after workspace 1
        assert mock_sleep.call_count == 1
        mock_sleep.assert_called_with(1.5)


def test_workflow_respects_app_launch_wait(workflow_manager):
    """Test that workflow manager respects app launch wait time."""
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1"},
                    {"name": "App2", "exec": "app2"},
                ],
            },
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 0,  # No wait after workspace switch
        "app_launch_wait": 0.7,  # Wait after each app launch
        "respect_app_wait": True,
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.core.workflow.time.sleep") as mock_sleep,
        patch("floww.core.workflow.typer.echo") as mock_echo,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called after first app launch, not after last app
        assert mock_sleep.call_count == 1
        mock_sleep.assert_called_once_with(0.7)

        # Verify waiting message was displayed
        mock_echo.assert_any_call("    ... Waiting 0.7s before next action...")


def test_workflow_respects_app_specific_wait(workflow_manager):
    """Test that workflow manager respects app-specific wait time."""
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1", "wait": 1.2},  # App-specific wait
                    {"name": "App2", "exec": "app2"},  # No app-specific wait
                ],
            },
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 0,  # No wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": True,  # Respect app-specific wait
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.core.workflow.time.sleep") as mock_sleep,
        patch("floww.core.workflow.typer.echo") as mock_echo,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called with app-specific wait for App1
        # and default wait for App2
        assert mock_sleep.call_count == 1  # Only for App1, App2 is last app
        mock_sleep.assert_called_once_with(1.2)  # App-specific wait time

        # Verify waiting message was displayed with app-specific wait time
        mock_echo.assert_any_call("    ... Waiting 1.2s before next action...")


def test_workflow_disables_app_specific_wait(workflow_manager):
    """Test that workflow manager can ignore app-specific wait time."""
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {
                        "name": "App1",
                        "exec": "app1",
                        "wait": 1.2,
                    },  # App-specific wait (should be ignored)
                    {"name": "App2", "exec": "app2"},  # No app-specific wait
                ],
            },
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 0,  # No wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": False,  # IGNORE app-specific wait
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.core.workflow.time.sleep") as mock_sleep,
        patch("floww.core.workflow.typer.echo") as mock_echo,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called with config wait for App1, not app-specific
        assert mock_sleep.call_count == 1  # Only for App1, App2 is last app
        mock_sleep.assert_called_once_with(0.5)  # Config wait time

        # Verify waiting message was displayed with config wait time
        mock_echo.assert_any_call("    ... Waiting 0.5s before next action...")


def test_workflow_respects_last_app_wait(workflow_manager):
    """Test that workflow manager respects wait time from last app in a workspace."""
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1"},
                    {"name": "App2", "exec": "app2", "wait": 2.5},  # Last app with wait
                ],
            },
            {"target": 2, "apps": [{"name": "App3", "exec": "app3"}]},
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 1.0,  # Wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": True,  # Respect app-specific wait
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.core.workflow.time.sleep") as mock_sleep,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called for:
        # 1. After App1 (0.5s from app_launch_wait)
        # 2. After App2 (2.5s from app-specific wait)
        # 3. After workspace 1 (2.5s from App2's wait again)
        # No sleep after workspace 2 as it's the last workspace
        assert mock_sleep.call_count == 3

        # Check that the calls were made with the right arguments
        mock_sleep.assert_any_call(0.5)  # App launch wait
        mock_sleep.assert_any_call(2.5)  # App2's wait (applied twice)


def test_workflow_respects_last_app_wait_with_final_workspace(workflow_manager):
    """Test that wait time for the last app in the last workspace is respected when there's a final_workspace."""
    workflow_data = {
        "workspaces": [
            {"target": 1, "apps": [{"name": "App1", "exec": "app1"}]},
            {
                "target": 2,
                "apps": [
                    {"name": "App2", "exec": "app2", "wait": 3.0}
                ],  # Last app in last workspace with wait
            },
        ],
        "final_workspace": 0,  # Add final workspace to return to
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 1.0,  # Wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": True,  # Respect app-specific wait
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with patch("floww.core.workflow.time.sleep") as mock_sleep:
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called for:
        # 1. After workspace 1 (1.0s from workspace_switch_wait)
        # 2. After App2 (3.0s from app-specific wait)
        # 3. Before final workspace (3.0s using last app's wait)
        assert mock_sleep.call_count == 3

        # Check that the calls were made with the right arguments
        mock_sleep.assert_any_call(1.0)  # Workspace switch wait
        mock_sleep.assert_any_call(
            3.0
        )  # App2's wait (applied twice: once after launching and once before final workspace)


def test_workspace_wait_before_final_workspace(workflow_manager):
    """Test that workspace wait is properly applied before switching to the final workspace."""
    workflow_data = {
        "workspaces": [
            {"target": 1, "apps": [{"name": "App1", "exec": "app1"}]},
            {
                "target": 2,
                "apps": [
                    {"name": "App2", "exec": "app2", "wait": 2.5}
                ],  # Last app in last workspace with wait
            },
        ],
        "final_workspace": 0,
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 1.5,  # Wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": True,
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with patch("floww.core.workflow.time.sleep") as mock_sleep:
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called for:
        # 1. After workspace 1 (1.5s from workspace_switch_wait)
        # 2. After App2 (2.5s from app-specific wait)
        # 3. Before final workspace (2.5s using last app's wait)
        assert mock_sleep.call_count == 3

        # Check the specific calls were made
        mock_sleep.assert_any_call(1.5)  # Workspace switch wait
        mock_sleep.assert_any_call(
            2.5
        )  # App2's wait before continuing and also for final workspace


def test_default_workspace_wait_before_final_workspace(workflow_manager):
    """Test that default workspace_switch_wait is applied before final workspace when last app has no wait."""
    workflow_data = {
        "workspaces": [
            {"target": 1, "apps": [{"name": "App1", "exec": "app1"}]},
            {
                "target": 2,
                "apps": [{"name": "App2", "exec": "app2"}],  # Last app with no wait
            },
        ],
        "final_workspace": 0,
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 2.0,  # Wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": True,
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with patch("floww.core.workflow.time.sleep") as mock_sleep:
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called for:
        # 1. After workspace 1 (2.0s from workspace_switch_wait)
        # 2. Before final workspace (2.0s using workspace_switch_wait since last app has no wait)
        assert mock_sleep.call_count == 2

        # Check both calls were made with workspace_switch_wait
        mock_sleep.assert_called_with(2.0)  # Workspace switch wait for both transitions


def test_apply_appends_workspaces(monkeypatch):
    """
    Test that append=True causes workspace targets to be offset by total_workspaces - 1.
    """
    manager = WorkflowManager(show_notifications=False)
    manager.workspace_mgr = DummyWorkspaceManager()
    manager.app_launcher = DummyAppLauncher()
    manager.config_mgr = DummyConfigManager()

    # Avoid real sleep delays
    monkeypatch.setattr(time, "sleep", lambda s: None)

    workflow_data = {
        "workspaces": [
            {"target": 1, "apps": [{"exec": "foo"}]},
            {"target": 2, "apps": [{"exec": "bar"}]},
        ],
    }

    result = manager.apply(workflow_data, append=True)

    assert result is True
    # total_workspaces = 5, so offset = 5 - 1 = 4
    assert manager.workspace_mgr.calls == [1 + 4, 2 + 4]


def test_apply_no_append_uses_original(monkeypatch):
    """
    Test that append=False leaves workspace targets unchanged.
    """
    manager = WorkflowManager(show_notifications=False)
    manager.workspace_mgr = DummyWorkspaceManager()
    manager.app_launcher = DummyAppLauncher()
    manager.config_mgr = DummyConfigManager()

    monkeypatch.setattr(time, "sleep", lambda s: None)

    workflow_data = {
        "workspaces": [
            {"target": 0, "apps": [{"exec": "foo"}]},
        ],
    }

    result = manager.apply(workflow_data, append=False)

    assert result is True
    assert manager.workspace_mgr.calls == [0]
