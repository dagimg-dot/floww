PYTHONPATH=src
PYINSTALLER_OPTS=\
    --onefile \
	--hidden-import shellingham.posix \
    --name floww \
    --clean \
    --noconfirm \
    --strip \
    --log-level WARN \
    --upx-dir /usr/bin --upx-exclude '*.so' \
    --exclude-module pytest \
    --exclude-module tests \

install-deps:
	uv pip install -e .
	uv pip install -e .[test]

clean:
	rm -rf dist build *.spec __pycache__ .pytest_cache test_config

test:
	PYTHONPATH=$(PYTHONPATH) XDG_CONFIG_HOME=tests/test_config pytest

build:
	# Include hidden imports for shell detection (used by Typer for --show-completion)
	pyinstaller $(PYINSTALLER_OPTS) src/floww/__main__.py