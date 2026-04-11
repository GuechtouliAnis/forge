# Forge 🔨

Forge is a developer CLI that scaffolds repositories and manages environment files — so you spend less time on setup and more time building.

> Forge is designed for Linux and macOS. Windows is not currently supported.

## Commands

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

#### `forge repo init`
Initializes a new git repository with forge scaffolding — generates a `.gitignore`, `README.md`, and `LICENSE` in one shot. Defaults to generic gitignore and MIT license if no options provided.
```bash
forge repo init                                         # initialize current directory
forge repo init my-project                              # create and initialize a new directory
forge repo init my-project --lang go                    # initialize with Go gitignore
forge repo init my-project --license apache             # initialize with Apache license
forge repo init my-project --lang go --license apache   # Go gitignore, Apache license
```

---

## Installation

**Using Go:**
```bash
go install github.com/GuechtouliAnis/forge@latest
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