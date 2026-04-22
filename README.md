# Forge 🔨

Forge is a developer CLI that scaffolds repositories and manages environment files — so you spend less time on setup and more time building.

> Forge is designed for Linux and macOS. Windows is not currently supported.
> ⚠️ Forge is in early development. APIs and commands are subject to change.

## Commands

### `forge config`
Groups all `.forge.toml` management subcommands.
```bash
forge config   # prints help and available subcommands
```

#### `forge config new`
Generates a `.forge.toml` configuration scaffold in the current or specified directory. Prompts before overwriting an existing file.
```bash
forge config new              # generate .forge.toml in current directory
forge config new path/to/dir  # generate .forge.toml in specified directory
```

---

### `forge env`
Groups all `.env` management subcommands — generate, validate, and populate your env files.
```bash
forge env   # prints help and available subcommands
```

#### `forge env init`
Initializes a .env file from .env.example. If no example file exists, it creates an empty one. It also automatically adds the target file to .gitignore to prevent secret leaks.
```bash
forge env init                 # create .env and update .gitignore
forge env init .env.dev        # initialize a custom path
forge env init --no-gitignore  # initialize without modifying .gitignore
```

#### `forge env check`
Validates a `.env` file against key naming rules, reporting errors and warnings with line numbers.
```bash
forge env check            # validate .env, show errors and warnings
forge env check -e         # show errors only
forge env check .env.prod  # validate a specific file
```

#### `forge env add`
Appends predefined variable sets (presets) to your .env file. It skips existing keys and adds section headers for organization.
```bash
forge env add --db --redis     # add Database and Redis boilerplate
forge env add --ai --web       # add AI and Web framework presets
```

#### `forge env example`
Generates a `.env.example` from the current `.env` file, stripping values and preserving comments.
```bash
forge env example      # generate .env.example, prompt if it already exists
forge env example -y   # overwrite existing .env.example without prompt
```

> `forge env example` is currently in beta — review your `.env.example` before committing.
---

### `forge repo`
Groups all repository lifecycle subcommands — scaffold, generate, and manage your repo structure.
```bash
forge repo   # prints help and available subcommands
```

#### `forge repo gitignore`
Generates a `.gitignore` from an embedded template for the declared language, or a generic one if no language is provided. Prompts before overwriting an existing `.gitignore`.
```bash
forge repo gitignore                    # generate a generic .gitignore
forge repo gitignore python             # generate a Python .gitignore
forge repo gitignore go                 # generate a Go .gitignore
forge repo gitignore go path/to/dir     # generate a Go .gitignore in specified directory
```

#### `forge repo readme`
Generates a `README.md` scaffold in the current or specified directory. Project name is inferred from the directory name. Author is read from `git config user.name`, falls back to a prompt if not set.
```bash
forge repo readme              # generate README.md in current directory
forge repo readme path/to/dir  # generate README.md in specified directory
```

#### `forge repo license`
Generates a `LICENSE` file in the current or specified directory. Author is read from `git config user.name`, falls back to a prompt if not set. Year is inferred from the system clock. Defaults to MIT if no license type is provided.
```bash
forge repo license             # generate MIT LICENSE in current directory
forge repo license apache      # generate Apache LICENSE
forge repo license gpl path/   # generate GPL LICENSE in specified directory
```

#### `forge repo changelog`
Generates a `CHANGELOG.md` scaffold in the current or specified directory. Handles any casing variant of existing changelog files before overwriting.
```bash
forge repo changelog              # generate CHANGELOG.md in current directory
forge repo changelog path/to/dir  # generate CHANGELOG.md in specified directory
```

#### `forge repo init`
Initializes a new git repository with forge scaffolding — generates a `.gitignore`, `README.md`, `LICENSE`, and `CHANGELOG.md` in one shot. Defaults to generic gitignore and MIT license if no options provided.
```bash
forge repo init                                         # initialize current directory
forge repo init my-project                              # create and initialize a new directory
forge repo init my-project --lang go                    # initialize with Go gitignore
forge repo init my-project --license apache             # initialize with Apache license
forge repo init my-project --lang go --license apache   # Go gitignore, Apache license
```

---

### `forge git`
Opinionated git helpers — not a git replacement. Forge handles commit structure and convention enforcement; for everything else, use git directly.
```bash
forge git   # prints help and available subcommands
```

#### `forge git commit`
Validates a commit message against the domains and format defined in `.forge.toml`. Falls back to defaults if no config file exists.
```bash
forge git commit "[FEAT] add config loader"  # validate
```

#### `forge git clean`
Scans local branches and flags ones that are stale by age or commits behind the base branch. Dry-run by default — `--remove` shows what will be deleted and prompts for confirmation, `--force` skips the prompt. `main`, `master`, and the default branch are always protected.
```bash
forge git clean                             # dry-run, show stale branches
forge git clean --days 30                   # flag branches older than 30 days
forge git clean --behind 10                 # flag branches 10+ commits behind
forge git clean --days 30 --remove          # show deletions and prompt
forge git clean --days 30 --remove --force  # delete without prompt
```

---

#### `forge git undo`
Reverts the last commit and buffers the commit message to `.git/forge/undo_msg.txt` for reuse. Soft reset by default — staged files are preserved. `--hard` wipes uncommitted changes and requires confirmation if the worktree is dirty. Prompts before overwriting an existing buffered message on consecutive undos.
```bash
forge git undo          # soft reset, buffer last commit message
forge git undo --hard   # destructive reset, prompt if dirty worktree
```

---

## Installation

**Using Go:**
```bash
go install github.com/GuechtouliAnis/forge@v1.3.0
```
> Ensure `$(go env GOPATH)/bin` is in your `PATH`. On most systems this is `~/go/bin`.

**Build from source:**
```bash
git clone git@github.com:GuechtouliAnis/forge.git
cd forge
go mod tidy
go build -o forge .
sudo mv forge /usr/local/bin/
```

## Author

[Guechtouli Anis](https://github.com/GuechtouliAnis)

*In the heat of creation, Forge shapes raw repositories into living projects — one command, one strike at a time.*
<p align="center"><sub>Where data sparks the light of revelation. ✨</sub></p>