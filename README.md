# Forge 🔨
*In the heat of creation, Forge shapes raw repositories into living projects — one command, one strike at a time.*

Personal dev CLI — clone, scaffold, and set up projects your way.

> Forge is designed for Linux and macOS. Windows is not currently supported.

## Commands
### `forge env`
Groups all `.env` management subcommands — generate, validate, populate, and sync your env files.
```bash
forge env   # prints help and available subcommands
```

#### `forge env example`
Generates a `.env.example` from the current `.env` file, stripping values and preserving comments.
```bash
forge env example      # generate .env.example, prompt if it already exists
forge env example -y   # overwrite existing .env.example without prompt
```

> `forge env example` is currently in beta — review your `.env.example` before committing.
---
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

### `forge new`

Scaffolds a fresh project locally without git.
```bash
forge new --py myproject    # Python project
forge new --go myproject    # Go project
forge new myproject         # generic, no language setup
```

### `forge init`

Same as `new` but initializes a git repository and creates an initial commit.
```bash
forge init --py myproject   # Python project with git
forge init --go myproject   # Go project with git
forge init myproject        # generic with git
```

## Flags

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
go mod tidy
go build -o forge .
sudo mv forge /usr/local/bin/
```

## Author

[Guechtouli Anis](https://github.com/GuechtouliAnis)

<p align="center"><sub>Where data sparks the light of revelation. ✨</sub></p>