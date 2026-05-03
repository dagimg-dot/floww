import json
import os
import unittest
from unittest.mock import MagicMock, patch

from floww.core.workspace import WorkspaceManager, create_backend
from floww.core.workspace.backends import HyprlandBackend, NiriBackend, WmctrlBackend


class TestWorkspaceManager(unittest.TestCase):
    def setUp(self):
        if "HYPRLAND_INSTANCE_SIGNATURE" in os.environ:
            del os.environ["HYPRLAND_INSTANCE_SIGNATURE"]
        self._prev_niri_socket = os.environ.pop("NIRI_SOCKET", None)
        self._prev_xdg_desktop = os.environ.get("XDG_CURRENT_DESKTOP")
        os.environ["XDG_CURRENT_DESKTOP"] = "FlowwTest"
        self._cm_patcher = patch("floww.core.workspace.factory.ConfigManager")
        mock_cm_cls = self._cm_patcher.start()
        self.mock_config = MagicMock()
        self.mock_config.get_general_config.return_value = {
            "workspace_backend": "auto",
            "show_notifications": True,
        }
        mock_cm_cls.return_value = self.mock_config

    def tearDown(self):
        self._cm_patcher.stop()
        if self._prev_niri_socket is not None:
            os.environ["NIRI_SOCKET"] = self._prev_niri_socket
        if self._prev_xdg_desktop is None:
            os.environ.pop("XDG_CURRENT_DESKTOP", None)
        else:
            os.environ["XDG_CURRENT_DESKTOP"] = self._prev_xdg_desktop

    def test_auto_selects_wmctrl_when_no_hyprland_and_no_ewmh(self):
        with patch(
            "floww.core.workspace.factory.EwmhBackend.try_create", return_value=None
        ):
            wm = WorkspaceManager()
            self.assertIsInstance(wm._backend, WmctrlBackend)

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    def test_auto_selects_hyprland(self):
        wm = WorkspaceManager()
        self.assertIsInstance(wm._backend, HyprlandBackend)

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.backends.hyprland.run_command")
    def test_switch_hyprland_success(self, mock_run):
        mock_run.return_value = True
        wm = WorkspaceManager()
        result = wm.switch(2)
        self.assertTrue(result)
        mock_run.assert_called_once_with(["hyprctl", "dispatch", "workspace", "2"])

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.backends.hyprland.run_command")
    def test_switch_hyprland_failure(self, mock_run):
        mock_run.return_value = False
        wm = WorkspaceManager()
        self.assertFalse(wm.switch(2))

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.backends.hyprland.subprocess.run")
    def test_get_total_workspaces_hyprland(self, mock_sub_run):
        mock_response = MagicMock()
        mock_response.stdout = json.dumps(
            [{"id": 1, "name": "1"}, {"id": 3, "name": "3"}, {"id": 2, "name": "2"}]
        )
        mock_sub_run.return_value = mock_response

        wm = WorkspaceManager()
        total = wm.get_total_workspaces()
        self.assertEqual(total, 3)
        mock_sub_run.assert_called()

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.backends.hyprland.subprocess.run")
    def test_get_total_workspaces_hyprland_empty(self, mock_sub_run):
        mock_response = MagicMock()
        mock_response.stdout = json.dumps([])
        mock_sub_run.return_value = mock_response

        wm = WorkspaceManager()
        total = wm.get_total_workspaces()
        self.assertEqual(total, 1)

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    def test_explicit_wmctrl_under_hyprland_env(self):
        wm = WorkspaceManager(backend=create_backend("wmctrl"))
        self.assertIsInstance(wm._backend, WmctrlBackend)

    @patch.dict(os.environ, {"XDG_CURRENT_DESKTOP": "niri"}, clear=False)
    def test_auto_selects_niri_xdg(self):
        with patch(
            "floww.core.workspace.factory.EwmhBackend.try_create", return_value=None
        ):
            wm = WorkspaceManager()
            self.assertIsInstance(wm._backend, NiriBackend)

    @patch.dict(os.environ, {"NIRI_SOCKET": "/fake/socket"}, clear=False)
    def test_auto_selects_niri_socket(self):
        with patch(
            "floww.core.workspace.factory.EwmhBackend.try_create", return_value=None
        ):
            wm = WorkspaceManager()
            self.assertIsInstance(wm._backend, NiriBackend)

    @patch.dict(os.environ, {"XDG_CURRENT_DESKTOP": "niri"}, clear=False)
    @patch("floww.core.workspace.backends.niri.run_command")
    @patch("floww.core.workspace.backends.niri.subprocess.run")
    def test_switch_niri_success(self, mock_sub_run, mock_run):
        mock_sub_run.return_value = MagicMock()
        mock_sub_run.return_value.stdout = json.dumps(
            [
                {"id": 1, "idx": 1, "is_focused": True, "is_active": True},
                {"id": 2, "idx": 2},
            ]
        )
        mock_run.return_value = True
        wm = WorkspaceManager()
        self.assertTrue(wm.switch(2))
        mock_run.assert_called_once_with(
            ["niri", "msg", "action", "focus-workspace", "2"]
        )
        self.assertEqual(mock_sub_run.call_count, 1)

    @patch.dict(os.environ, {"XDG_CURRENT_DESKTOP": "niri"}, clear=False)
    @patch("floww.core.workspace.backends.niri.run_command")
    @patch("floww.core.workspace.backends.niri.subprocess.run")
    def test_switch_niri_target_one_sends_focus_one(self, mock_sub_run, mock_run):
        mock_sub_run.return_value = MagicMock()
        mock_sub_run.return_value.stdout = json.dumps(
            [
                {"id": 2, "idx": 2, "is_focused": True, "is_active": True},
                {"id": 1, "idx": 1},
            ]
        )
        mock_run.return_value = True
        wm = WorkspaceManager()
        self.assertTrue(wm.switch(1))
        mock_run.assert_called_once_with(
            ["niri", "msg", "action", "focus-workspace", "1"]
        )

    @patch.dict(os.environ, {"XDG_CURRENT_DESKTOP": "niri"}, clear=False)
    @patch("floww.core.workspace.backends.niri.run_command")
    @patch("floww.core.workspace.backends.niri.subprocess.run")
    def test_switch_niri_skips_focus_when_already_on_workspace(
        self, mock_sub_run, mock_run
    ):
        """Avoid focus-workspace when already there (niri workspace-auto-back-and-forth)."""
        mock_sub_run.return_value = MagicMock()
        mock_sub_run.return_value.stdout = json.dumps(
            [
                {"id": 1, "idx": 1, "is_focused": True, "is_active": True},
            ]
        )
        wm = WorkspaceManager()
        self.assertTrue(wm.switch(1))
        mock_run.assert_not_called()
        mock_sub_run.assert_called_once()

    @patch.dict(os.environ, {"XDG_CURRENT_DESKTOP": "niri"}, clear=False)
    @patch("floww.core.workspace.backends.niri.run_command")
    @patch("floww.core.workspace.backends.niri.subprocess.run")
    def test_switch_niri_focus_when_workspaces_unavailable(
        self, mock_sub_run, mock_run
    ):
        mock_sub_run.side_effect = OSError("no niri")
        mock_run.return_value = True
        wm = WorkspaceManager()
        self.assertTrue(wm.switch(1))
        mock_run.assert_called_once_with(
            ["niri", "msg", "action", "focus-workspace", "1"]
        )

    @patch.dict(os.environ, {"XDG_CURRENT_DESKTOP": "niri"}, clear=False)
    @patch("floww.core.workspace.backends.niri.subprocess.run")
    def test_get_total_workspaces_niri(self, mock_sub_run):
        mock_sub_run.return_value = MagicMock()
        mock_sub_run.return_value.stdout = json.dumps([{"id": 1}, {"id": 2}])

        wm = WorkspaceManager()
        self.assertEqual(wm.get_total_workspaces(), 2)
        mock_sub_run.assert_called_once_with(
            ["niri", "msg", "-j", "workspaces"],
            capture_output=True,
            text=True,
            check=True,
        )

    @patch.dict(os.environ, {"XDG_CURRENT_DESKTOP": "niri"}, clear=False)
    @patch("floww.core.workspace.backends.niri.subprocess.run")
    def test_get_total_workspaces_niri_empty(self, mock_sub_run):
        mock_sub_run.return_value = MagicMock()
        mock_sub_run.return_value.stdout = json.dumps([])

        wm = WorkspaceManager()
        self.assertEqual(wm.get_total_workspaces(), 1)

    def test_explicit_niri_backend(self):
        wm = WorkspaceManager(backend=create_backend("niri"))
        self.assertIsInstance(wm._backend, NiriBackend)


if __name__ == "__main__":
    unittest.main()
