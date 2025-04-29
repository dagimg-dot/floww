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
	uv pip install -e .[build]

clean:
	rm -rf dist build *.spec __pycache__ .pytest_cache

lint:
	ruff check . --fix
	ruff format .

test:
	PYTHONPATH=$(PYTHONPATH) XDG_CONFIG_HOME=$(shell pwd)/tests/test_config pytest -v

build:
	pyinstaller $(PYINSTALLER_OPTS) src/floww/__main__.py

bump:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is not set. Usage: make bump VERSION=x.y.z"; \
		exit 1; \
	fi
	@echo "Bumping version to $(VERSION)..."
	@sed -i 's/^version = ".*"/version = "$(VERSION)"/' pyproject.toml
	@sed -i 's/^__version__ = ".*"/__version__ = "$(VERSION)"/' src/floww/__init__.py
	@echo "floww version updated to $(VERSION)"

release: bump
	@echo "Committing version update..."
	git add pyproject.toml src/floww/__init__.py
	git commit -m "chore: bump version to v$(VERSION)"
	git push
	@echo "Creating and pushing tag v$(VERSION)..."
	git tag v$(VERSION)
	git push origin v$(VERSION)
	@echo "Release v$(VERSION) created and pushed. GitHub Actions should start automatically"
