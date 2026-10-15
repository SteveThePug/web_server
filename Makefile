# =============================================================================
# Makefile — one place for the commands that verify and run this repo.
#
# WHY THIS FILE EXISTS
#   The stack is four toolchains (Go, Python, Node, Docker) and the command
#   that checks each one used to live only in a gitignored notes file. `make
#   check` is the single thing to run before pushing; .gitea/workflows/
#   deploy.yaml runs the same backend and Python checks (in containers, since
#   the Pi runner has no host toolchains) before it deploys.
#
# WHAT IT NEEDS ON THE HOST
#   check-backend / generate   Go (the version in backend/go.mod)
#   check-python               python3 with the venv module
#   check-vue                  Node + `npm install` done in vue/
#   check-compose / dev        Docker with the compose plugin
#   Nothing here installs a toolchain; a missing one fails that target only.
#
# Recipes are plain `cd dir && cmd` one-liners on purpose: each line runs in
# its own shell, so the directory change cannot leak into the next line.
# =============================================================================

# `make` with no arguments prints the target list rather than running checks.
.DEFAULT_GOAL := help

# Overridable: `make check-python PYTHON=python3.12` to match the image.
PYTHON ?= python3

# The virtualenv for the Python checks. At the repo root rather than inside
# python/, because python/ is a docker build context and a venv there would be
# sent to the daemon on every build. Gitignored (/.venv/).
VENV := .venv

# Written after a successful install. Being a file target that depends on both
# requirements files, it is what makes `check-python` re-install when either
# changes and skip the install otherwise.
VENV_STAMP := $(VENV)/.installed

.PHONY: help check check-backend check-python check-vue check-compose generate dev

help:
	@echo "Targets:"
	@echo "  make check          run every check below"
	@echo "  make check-backend  go vet + go test in backend/"
	@echo "  make check-python   pytest in python/ (venv created on demand in $(VENV)/)"
	@echo "  make check-vue      vite build of vue/ into a throwaway directory"
	@echo "  make check-compose  validate docker-compose.yml, alone and with the dev override"
	@echo "  make generate       regenerate the gqlgen GraphQL code after a schema edit"
	@echo "  make dev            build and run the dev stack (http://localhost)"

# Run with `make -k check` to keep going past the first failing area.
check: check-backend check-python check-vue check-compose

check-backend:
	cd backend && go vet ./... && go test ./...

# requirements-dev.txt `-r`-includes requirements.txt, so one install covers
# the app's dependencies and pytest. The venv is recreated from scratch
# whenever a requirements file changes: upgrading in place would leave behind
# packages that were removed from the files, and the check would keep passing
# on a dependency the image no longer has.
$(VENV_STAMP): python/requirements.txt python/requirements-dev.txt
	rm -rf $(VENV)
	$(PYTHON) -m venv $(VENV)
	$(VENV)/bin/python -m pip install --quiet --upgrade pip
	$(VENV)/bin/python -m pip install --quiet -r python/requirements-dev.txt
	touch $@

check-python: $(VENV_STAMP)
	cd python && ../$(VENV)/bin/python -m pytest tests -q

# A Vite build is enough to catch a broken import or template; there is no
# test runner or linter for the frontend.
#
# The build goes to a temp directory that is removed afterwards (success or
# failure), so the check never leaves a vue/dist behind or overwrites one.
check-vue:
	@test -d vue/node_modules || { \
		echo "check-vue: vue/node_modules/ is missing. Run: cd vue && npm install"; \
		exit 1; }
	@out=$$(mktemp -d) && trap 'rm -rf "$$out"' EXIT && \
		cd vue && npx vite build --outDir "$$out" --emptyOutDir

# Validates the compose files: YAML, schema, and that every ${VAR} they
# interpolate resolves to something usable.
#
# With a real .env present this is simply `docker compose config -q`.
#
# Without one (a CI checkout, a fresh clone) two things break, and
# `--env-file .env.example` fixes neither:
#   - several services declare `env_file: ./.env`, and compose refuses to load
#     the project if that file does not exist;
#   - .env.example has blank values, and OBSIDIAN_DIR is used as a bind-mount
#     source, where an empty string is a hard "invalid spec" error.
# So in that case the files are validated against a temporary PROJECT
# DIRECTORY holding a copy of .env.example as its .env, with a placeholder for
# OBSIDIAN_DIR. `config` only resolves paths, it does not need them to exist,
# so the build contexts not being under the temp directory is fine. Nothing is
# written to the repo, and a real .env is never created by accident.
check-compose:
	@set -e; \
	if [ -f .env ]; then \
		pd="."; \
	else \
		echo "check-compose: no .env, validating against .env.example"; \
		pd=$$(mktemp -d); trap 'rm -rf "$$pd"' EXIT; \
		cp .env.example "$$pd/.env"; \
		echo "OBSIDIAN_DIR=/placeholder/obsidian" >> "$$pd/.env"; \
	fi; \
	echo "docker compose -f docker-compose.yml config -q"; \
	docker compose --project-directory "$$pd" -f docker-compose.yml config -q; \
	echo "docker compose -f docker-compose.yml -f docker-compose.dev.yml config -q"; \
	docker compose --project-directory "$$pd" -f docker-compose.yml -f docker-compose.dev.yml config -q

# Rewrites backend/graph/generated.go and graph/model/models_gen.go from
# graph/schema/*.graphql. Resolver bodies in *.resolvers.go are preserved.
generate:
	cd backend && go run github.com/99designs/gqlgen generate

# Dev stack: HTTP only, seeded database, Vite hot reload behind nginx.
dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
