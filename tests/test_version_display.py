import os

import pytest

import floww


@pytest.fixture
def restore_env():
    env_keys = ("FLOWW_VERSION_SUFFIX", "FLOWW_DEV", "ENV")
    saved = {k: os.environ.get(k) for k in env_keys}
    yield
    for k in env_keys:
        v = saved[k]
        if v is None:
            os.environ.pop(k, None)
        else:
            os.environ[k] = v


def test_version_display_plain(monkeypatch, restore_env):
    os.environ.pop("FLOWW_VERSION_SUFFIX", None)
    os.environ.pop("FLOWW_DEV", None)
    os.environ.pop("ENV", None)
    monkeypatch.setattr(floww, "_first_dotenv_path", lambda: None)
    assert floww.version_display() == floww.__version__


def test_version_display_floww_version_suffix(restore_env):
    os.environ["FLOWW_VERSION_SUFFIX"] = "@rc"
    assert floww.version_display() == floww.__version__ + "@rc"


def test_version_display_env_var(restore_env):
    os.environ.pop("FLOWW_VERSION_SUFFIX", None)
    os.environ.pop("FLOWW_DEV", None)
    os.environ["ENV"] = "staging"
    assert floww.version_display() == f"{floww.__version__}@staging"


def test_version_display_from_dotenv(monkeypatch, restore_env, tmp_path):
    os.environ.pop("FLOWW_VERSION_SUFFIX", None)
    os.environ.pop("FLOWW_DEV", None)
    os.environ.pop("ENV", None)
    dotenv = tmp_path / ".env"
    dotenv.write_text("ENV=dev\n", encoding="utf-8")
    monkeypatch.setattr(floww, "_first_dotenv_path", lambda: dotenv)
    assert floww.version_display() == f"{floww.__version__}@dev"


def test_version_display_dotenv_strips_at(monkeypatch, restore_env, tmp_path):
    os.environ.pop("FLOWW_VERSION_SUFFIX", None)
    os.environ.pop("FLOWW_DEV", None)
    os.environ.pop("ENV", None)
    dotenv = tmp_path / ".env"
    dotenv.write_text("ENV=@beta\n", encoding="utf-8")
    monkeypatch.setattr(floww, "_first_dotenv_path", lambda: dotenv)
    assert floww.version_display() == f"{floww.__version__}@beta"
