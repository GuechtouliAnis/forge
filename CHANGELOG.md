# Changelog

## [Unreleased]

### Added
- `forge config new` — Path existence and directory validation
- `forge config new` — Soft `.git/` reminder printed after successful creation when no `.git/` is detected at target path
- `forge repo` — Added `ResolveCaseInsensitive` to perform case insensitive search for a file and return the actual appearance name

### Changed
- `forge config new` — Replaced `fmt.Scanln` with `bufio.NewReader` for overwrite prompt input — prevents silent failure on empty input or redirected stdin
- `forge config new` — All errors now wrapped with `fmt.Errorf("[config new]: %w", err)` instead of being returned bare
- `forge config new` — Removed redundant `RemoveFileInsensitive` call before overwrite — `os.WriteFile` truncates by default
- `cmd/root` — Removed hardcoded `--username` / `-u` persistent flag
- `cmd/root` — Error handling centralized in `Execute()` — commands return errors bare instead of printing to stderr themselves
- `repo` — Optimized `RemoveFileInsensitive` with an initial `os.Remove` for case sensitive direct remove
- `forge env add` — Replaced `strings.NewReader(string(data))` with `bytes.NewReader(data)` — avoids redundant `[]byte` to `string` conversion
- `forge env add` — Replaced line-by-line `strings.Index` with `strings.Cut` for cleaner key parsing
- `forge env add` — Collapsed triple preset iteration into a single pass using a `toWrite` entry slice
- `forge env add` — Buffered file writes with `bufio.NewWriter` — reduces syscalls per appended line
- `forge env add` — Replaced `fmt.Scanln` based parsing with `bufio.NewScanner` for line reading
- `forge env add` — All errors now wrapped with `fmt.Errorf("[env add]: %w", err)`
- `forge env add` — Error message for missing preset flag aligned with codebase convention

## Fixed
- **The "Duplicate Config" Bug**: Fixed an issue on Linux where overwriting a case-mismatched file (e.g., `.FORGE.toml`) would create a second file (`.forge.toml`) instead of replacing the original.

## [1.4.0] - 2026-05-09

### Added
- `forge repo changelog [path]` — generate a CHANGELOG.md scaffold in the current or specified directory
- `forge repo changelog` — prompts before overwriting existing changelog, handles any casing variant (e.g. `ChangElog.Md`)
- `forge repo init` — now also generates a CHANGELOG.md as part of the scaffold
- `forge repo changelog` — added to README
- `forge config` — new command group for managing `.forge.toml` configuration
- `forge config new [path]` — generate a `.forge.toml` scaffold in the current or specified directory, prompts before overwriting existing file
- `internal/config` — loads and parses `.forge.toml`, merges over defaults
- `forge git` — new command group for git workflow commands with guardrails
- `forge git commit <message>` — validate a commit message against `.forge.toml` rules before committing; checks format, domain allowlist, and message length
- `commit_test.go` — to test the behaviour of `forge git commit <message>`
- `forge git clean` — detect and remove stale local branches by age or commits behind
- `forge git clean` — dry-run by default, `--remove` shows deletions and prompts, `--force` skips confirmation
- `forge git clean` — `main`, `master`, and default branch are always protected
- `forge git clean` — reads `stale_days` and `commits_behind` from `[git.clean]` in `.forge.toml`
- `forge git undo` — revert the last commit with message buffered to `.git/forge/undo_msg.txt` for reuse
- `forge git undo` — soft reset by default, `--hard` wipes uncommitted changes with confirmation prompt
- `forge git undo` — prompts before overwriting an existing buffered message on consecutive undos
- `forge git restore <search>` — recover a deleted or modified file from git history using fuzzy path matching
- `forge git restore` — collision detection blocks overwrites of dirty or staged files unless `--force` is passed
- `forge git restore` — interactive commit picker shows last 10 commits with diff stats, deletion commits filtered out
- `forge git restore` — `--latest` skips the menu and restores from the most recent non-deletion commit
- `forge git restore` — `--commit <hash>` pins restore to a specific commit
- `forge git restore` — `--dry-run` previews match and target commit without touching the filesystem
- `forge git restore` — restored file is left unstaged for user review

### Changed
- `forge repo gitignore`, `forge repo license`, `forge repo readme`, `forge repo changelog` — use shared `CheckFileExists` and `RemoveFileInsensitive` utilities for case-insensitive file detection and safe overwrite
- `forge repo changelog` — date format fixed, using Go birthday format
- `forge git commit <message>` — validation logic replaced: removed named capture groups, `CommitError` type, and `buildPattern`
- `forge git commit` now uses `CreatePattern` (regex assembled via `regexp.QuoteMeta` + placeholder substitution) and `ValidateCommit` returns `(bool, error)` instead of a single error
- `cmd/` refactored from flat file package into grouped subdirectories — `cmd/env/`, `cmd/repo/`, `cmd/git/`, `cmd/config/`
- each command group now lives in its own package with a `Register(root *cobra.Command)` entry point
- `cmd/root.go` now wires all command groups via `Register` calls instead of relying on `init()` side effects across files

### Deleted
- internal/project package which had `clone` and helper function `run` has been dropped and will be replaced

## [1.3.0] - 2026-04-11

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
- `forge repo init [path]` — initialize a new git repository with forge scaffolding in one shot
- `forge repo init` — generates .gitignore, README.md, and LICENSE then makes an initial commit
- `forge repo init` — aborts commit if .env is staged, preventing accidental secret leaks
- `forge repo init` — accepts `--lang` and `--license` flags to override defaults

### Changed
- reorganised embedded templates into subdirectories — `templates/licenses/` for license templates, `templates/gitignore/` for gitignore templates, `templates/readme/` for readme template
- `forge repo gitignore` — accepts optional path argument to generate .gitignore in a specified directory
- `SilenceErrors` and `SilenceUsage` moved to root command — applies globally, eliminates duplicate error output and noisy usage dumps on failure

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