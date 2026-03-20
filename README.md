# Forge 🔨

Personal dev CLI — clone, scaffold, and set up projects your way.

## Commands

### `forge clone`

Clones a repository and sets up the development environment automatically.
```bash
forge clone --py git@github.com:user/repo.git   # Python project
forge clone --go git@github.com:user/repo.git   # Go project
forge clone git@github.com:user/repo.git        # just clone, no setup
```

### `forge gitignore`

Generates a `.gitignore` for the current project.
```bash
forge gitignore        # generic
forge gitignore --py   # Python project
forge gitignore --go   # Go project
```

**Flags:**
- `--py` — creates a venv, upgrades pip, installs from `pyproject.toml` or `requirements.txt` if present
- `--go` — runs `go mod init` if no `go.mod` exists, then `go mod tidy`
- `-u / --username` — GitHub username for Go module path (falls back to git config if not provided)

**Python setup behavior:**
- `pyproject.toml` found → `pip install -e .`
- `requirements.txt` found → `pip install -r requirements.txt`
- neither found → venv created and ready, no deps installed

## Installation
```bash
git clone git@github.com:GuechtouliAnis/forge.git
cd forge
go build -o forge .
sudo mv forge /usr/local/bin/
```

## Roadmap

- `forge new --py/--go <name>` — scaffold a fresh project from scratch
- `forge gitignore --py/--go` — generate a `.gitignore` for the project type