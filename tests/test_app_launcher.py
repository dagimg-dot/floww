import pytest
from unittest.mock import patch, MagicMock
from pathlib import Path

from floww.app_launcher import AppLauncher


@pytest.fixture
def app_launcher():
    return AppLauncher()


def test_launch_binary_app(app_launcher):
    app_config = {
        "name": "Test App",
        "exec": "test-app",
        "args": ["--flag", "value"],
    }

    with patch("floww.app_launcher.subprocess.Popen") as mock_popen:
        mock_process = MagicMock()
        mock_process.pid = 12345
        mock_popen.return_value = mock_process

        success = app_launcher.launch_app(app_config)
        assert success is True

        # Verify the command was constructed correctly
        mock_popen.assert_called_once()
        args, kwargs = mock_popen.call_args
        assert args[0] == ["test-app", "--flag", "value"]
        assert kwargs["start_new_session"] is True


def test_launch_flatpak_app(app_launcher):
    app_config = {
        "name": "Flatpak App",
        "exec": "org.example.app",
        "type": "flatpak",
        "args": ["--arg1", "value1"],
    }

    with patch("floww.app_launcher.subprocess.Popen") as mock_popen:
        mock_process = MagicMock()
        mock_process.pid = 12345
        mock_popen.return_value = mock_process

        success = app_launcher.launch_app(app_config)
        assert success is True

        # Verify flatpak command was constructed correctly
        mock_popen.assert_called_once()
        args, kwargs = mock_popen.call_args
        assert args[0] == ["flatpak", "run", "org.example.app", "--arg1", "value1"]
        assert kwargs["start_new_session"] is True


def test_launch_snap_app(app_launcher):
    app_config = {
        "name": "Snap App",
        "exec": "snap-app",
        "type": "snap",
        "args": ["--config", "~/config.yaml"],
    }

    with patch("floww.app_launcher.subprocess.Popen") as mock_popen:
        mock_process = MagicMock()
        mock_process.pid = 12345
        mock_popen.return_value = mock_process

        success = app_launcher.launch_app(app_config)
        assert success is True

        # Verify snap command was constructed correctly
        mock_popen.assert_called_once()
        args, kwargs = mock_popen.call_args
        # Check that ~ was expanded in the args
        home = str(Path.home())
        assert args[0] == ["snap-app", "--config", f"{home}/config.yaml"]
        assert kwargs["start_new_session"] is True


def test_launch_app_not_found(app_launcher):
    app_config = {
        "name": "Missing App",
        "exec": "non-existent-app",
    }

    with patch("floww.app_launcher.subprocess.Popen") as mock_popen:
        mock_popen.side_effect = FileNotFoundError()

        success = app_launcher.launch_app(app_config)
        assert success is False


def test_launch_app_permission_error(app_launcher):
    app_config = {
        "name": "No Permission App",
        "exec": "/root/app",
    }

    with patch("floww.app_launcher.subprocess.Popen") as mock_popen:
        mock_popen.side_effect = PermissionError()

        success = app_launcher.launch_app(app_config)
        assert success is False


def test_launch_app_invalid_type(app_launcher):
    app_config = {
        "name": "Invalid Type App",
        "exec": "app",
        "type": "invalid",
    }

    success = app_launcher.launch_app(app_config)
    assert success is False


def test_launch_app_with_tilde_in_exec(app_launcher):
    app_config = {
        "name": "Home App",
        "exec": "~/bin/app",
    }

    with patch("floww.app_launcher.subprocess.Popen") as mock_popen:
        mock_process = MagicMock()
        mock_process.pid = 12345
        mock_popen.return_value = mock_process

        success = app_launcher.launch_app(app_config)
        assert success is True

        # Verify ~ was expanded in the executable path
        mock_popen.assert_called_once()
        args, kwargs = mock_popen.call_args
        home = str(Path.home())
        assert args[0][0] == f"{home}/bin/app"
