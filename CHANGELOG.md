# Changelog

## [Unreleased]

### Added
- `forge repo` — parent command grouping all repo subcommands, prints help when called alone
- `forge repo gitignore [language]` — generate a .gitignore from embedded templates for `py`/`python`, `go`/`golang`, or generic if no language provided
- `forge repo gitignore` — prompts before overwriting existing .gitignore, returns error on unsupported language argument
- embedded gitignore templates compiled into binary at build time (no external files required)
- `forge repo readme [path]` — generate a README.md scaffold in current or specified directory
- `forge repo readme` — infers project name from directory name, author from `git config user.name` with prompt fallback
- `forge repo readme` — author rendered as a GitHub profile link
- `forge env example` — tolerates `y`, `Y` and case insensitive forms of `yes`
- `forge repo license [license] [path]` — generate a LICENSE file from embedded templates for `mit`, `apache`, `gpl`, `agpl`, `bsd`. defaults to `mit` if omitted
- `forge repo license` — infers author from `git config user.name` with prompt fallback, year from system clock
- `forge repo license` — prompts before overwriting existing LICENSE

### Changed
- reorganised embedded templates into subdirectories — `templates/licenses/` for license templates, `templates/gitignore/` for gitignore templates, `templates/readme/` for readme template

## [1.2.2] - 2026-04-10

### Added
- `forge env` — parent command grouping all env subcommands, prints help when called alone
- `forge env check` — validate a .env file against key naming rules, reports errors and warnings with line numbers
- `forge env check` — warns on commented key=value lines that still contain a value
- `forge env add` — append predefined variable sets to .env (`--db`, `--ai`, `--web`, `--redis`, `--monitoring`, `--neo4j`)
- `forge env add` — skips existing keys with a warning, errors if all keys already exist
- `forge env add` — inserts section comment per preset (e.g. `# db - added by forge env add`)
- `forge env add` — host and port vars get sensible defaults, others default to `""`
- `forge env init` — create a .env file from .env.example (or empty) and automatically register it in .gitignore unless `--no-gitignore` is passed
- `forge env add` — warns if a preset key exists but is commented out in the .env file

### Changed
- `forge env` renamed to `forge env example` — breaking change for existing users
- `cmd/env.go` split into `cmd/env.go` (parent) and `cmd/env_example.go` (subcommand)
- `forge env add` — replaced triple nested loop with a flat `presetKeys` map for O(1) comment key lookup

### Fixed
- `forge env example` — quoted values containing `#` in commented key=value lines now correctly strip the value instead of leaking it
- `forge env check` — empty value detection now correctly handles values that are inline comments (e.g. `KEY= # no value`)
- `forge env check` — panic on lines containing only `=` (empty key) now returns a proper error instead of crashing
- `forge env add` — single `eqIdx` lookup per line instead of recomputing inside each branch

## [1.2.1] - 2026-04-05

### Fixed
- `forge env` — commented key=value lines with inline comments now preserve the comment (e.g. `# KEY=secret # comment` → `# KEY=  # comment`)
- `forge env` — commented key=value lines (e.g. `# API_KEY=secret`) now have their values stripped instead of being returned as-is

## [1.2.0] - 2026-04-03

### Added
- Added key validation into exported `ValidateKey` function
- Added warning for keys starting with digit, containing invalid characters, or lowercase
- `ParseEnv` now strips `export ` prefix from lines before processing

### Changed
- Restructured `internal/` into domain packages: `project/`, `lang/`, `env/`, `git/`, `guard/`, `template/`, `sync/`

### Tests
- Added table-driven tests for `ParseEnv`, `ValidateKey`, and `transformLine`

## [1.1.1] - 2026-04-01

### Fixed
- `forge env` — correctly handles `#` characters inside quoted values (`"val#ue"`, `'val#ue'`)
- Replaced position-based parsing with regex for reliable inline comment detection

## [1.1.0] - 2026-03-30

### Added
- `forge env` — generate a `.env.example` from `.env`, stripping values and preserving comments
- `forge env -y` — overwrite existing `.env.example` without prompt
- Duplicate key detection with warnings during `.env` parsing

## [1.0.0] - 2026-03-20

### Added
- `forge clone` — clone a repo with automatic Python/Go environment setup
- `forge new` — scaffold a fresh local project
- `forge init` — scaffold a project with git initialized
- `forge gitignore` — generate a .gitignore for Python, Go, or generic projects