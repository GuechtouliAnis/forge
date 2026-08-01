# Changelog

## [Unreleased]

## [1.8.0] - 2026-08-01

### Added
- `forge git add` command: wraps `git add` with a guardrail pipeline applied to every file before it reaches the index.
- `.env` and the configured `env.default_file` are unconditionally rejected by `forge git add`, regardless of `blocklist_patterns` or how the file is named on the command line.
- `git.add.blocklist_patterns` config option (`.forge.toml`) for glob-pattern rejection of sensitive files (e.g. `*.pem`, `id_rsa`), matched case-insensitively against each file's base name.
- `git.add.max_file_size_mb` config option to flag oversized files before staging. Defaults to `10`; set to `-1` to disable.
- `git.add.on_file_size_violation` config option to control oversized-file handling: `warn` (stage anyway, print a warning), `skip` (leave unstaged), or `fail` (abort the run).
- `--dry-run` / `-d` flag on `forge git add` to preview staged/blocked/skipped decisions without touching the index.
- `forge git add` path resolution distinguishes explicitly-named files from directory arguments: named files are always evaluated regardless of `.gitignore` status, so files like `.env` cannot bypass guardrails by being ignored; directory arguments are expanded via `git status --porcelain`, scoped to files with genuine unstaged or untracked changes.
- `forge git add` now supports staging deletions for explicitly-named paths that no longer exist on disk but are still tracked by git, instead of rejecting them as invalid paths.
- `forge git add --dry-run` output now includes a `CODE` column showing each file's raw git status code (e.g. `M`, `??`, `R`) alongside its would-stage/blocked/skipped status.
- `forge git add`'s final summary now reports how many files failed to stage, in addition to staged/blocked/skipped counts.
- `forge git commit` now actually runs `git commit` and reports the resulting commit hash on success — previously, a valid message was only validated and printed `✓ valid`, with no commit ever created.
- `--amend` support for `forge git commit`: reuses the previous commit's message unvalidated when no new message is supplied, or validates and applies a new message when `-m` is given alongside it. Unlike a regular commit, `--amend` does not require staged changes, since it can rewrite the previous commit's message alone with a clean index.
- `git.commit.domaine_case_sensitive` config option: when `true` (default), a domain-only case mismatch (e.g. `fix` instead of `FIX`) is accepted with a warning instead of being rejected outright; when `false`, domain matching is case-insensitive from the start with no warning.
- `forge git commit` now rejects regular commits with nothing staged (`NOTHING_STAGED`), matching `git commit`'s own behavior, instead of allowing an empty commit through validation. `--amend` is exempt from this check.
- `git.commit.staged_ignored_files_mode` config option (`warn`/`block`, default `warn`): controls whether a staged file matching a `.gitignore` rule (possible via an explicit `git add -f`) triggers a warning that still allows the commit, or blocks it outright.
- `--dry-run` flag on `forge git commit`: runs all the same validation and warning/block checks as a real commit, but stops short of actually committing — printing the message that would be used (or the previous message being reused, for a message-less `--amend`), followed by a table of staged files and their status codes.
- When a staged file matches a `.gitignore` rule, `forge git commit` now offers to unstage it interactively (`[y/N]`) before falling back to `staged_ignored_files_mode`; this prompt is skipped during `--dry-run`, which never mutates git state.
- `.forge.toml` lookup now walks upward from the current directory (stopping at the file, a `.git` boundary, or filesystem root) instead of only checking the current directory, so commands work correctly from subdirectories; prints `no .forge.toml found — using defaults` when nothing is found.

### Changed
- `forge git commit`'s commit message is now supplied via `-m`/`--message` instead of a positional argument, matching the spec and git's own convention (`forge git commit -m "..."` instead of `forge git commit "..."`).

### Fixed
- `git.commit.domains` config key corrected to `domain` (singular) to match the field actually read from `.forge.toml` — previously the mismatch silently left the domain list empty, causing every commit with a `{domain}` format to fail with `"format contains {domain} but no valid domains are defined"` regardless of a correctly configured file.
- `gitCheck` (shared preflight validation used by all `forge git` subcommands) now distinguishes "git is not installed" from "not a git repository" — previously a missing git installation was misreported as the latter.
- Domain values in `git.commit.domain` containing regex metacharacters (e.g. `C++`) are now escaped before being compiled into the commit-message validation pattern — previously such characters were interpreted as regex syntax, silently altering match behavior.

### Removed
- `internal/git/commit_test.go` removed

## [1.7.0] - 2026-07-17

### Added
- `git.clean.fetch_remote` config option (`.forge.toml`) to control whether `forge git clean` runs `git fetch --prune` before evaluating branches. Defaults to `true`.
- `--offline` flag on `forge git clean` to skip the fetch for a single invocation, regardless of `fetch_remote`.
- Local-only diagnosis path for `forge git clean`: when fetch is skipped or fails, branch evaluation now falls back to existing local refs and reports the age of the last successful fetch (via `.git/FETCH_HEAD`).
- `forge env add` now prints a preset table (name + keys) when called with no flags, instead of silently doing nothing.
- `forge env add --<preset> -h` now prints the keys that preset would append, instead of falling back to generic usage text.
- `AddEnv` now prints a success message on completion reporting how many variables were appended and to which file.
- `git.clean.fetch_timeout_seconds` config option to bound `git fetch --prune` with a timeout (default 30s), preventing an indefinite hang on network stalls or credential prompts.
- `git.clean.max_workers` config option to override the branch-evaluation worker pool size; auto-detected from CPU count and file descriptor limits when unset.
- Concurrent "commits behind" evaluation for `forge git clean`, bounded by a dynamically sized worker pool.

### Changed
- `forge git clean` branch deletion loop now explicitly guards unmerged branches; they are safely skipped with a warning unless the `--force` flag is actively passed to authorize a force-delete (`-D`).
- Improved the offline fetch age readout to dynamically format into hours and minutes (`1h15m ago`) if the last sync occurred over an hour ago, replacing the raw minute-only formatting.
- Refactored core branch loop iteration variables (`name`, `tsStr`, `ts`) to self-documenting equivalents (`branchName`, `timestampString`, `lastCommitTimestamp`) to improve long-term maintainability.
- `env package` exposes `PresetKeys(name string) ([]string, bool)` for preset lookup outside the package.
- `envAddCmd` help output now routes through `cmd.Printf`/`cmd.Println` instead of `fmt`, consistent with cobra output conventions.
- `forge git clean` batches branch discovery, age, and merged-status queries into three `git` calls total instead of three per branch (`for-each-ref`, `branch --merged`), keeping only the "commits behind" check per-branch.
- `git fetch --prune` now sets `GIT_TERMINAL_PROMPT=0` to fail fast instead of blocking on an interactive credential prompt.
- `forge git clean` internals split across `clean.go`, `clean_git.go`, and `clean_workers.go` for readability.
- Error and status messages in `git clean` systematically prefixed with `[git clean]:` for cleaner visual tracing and subcommand consistency.

### Fixed
- `forge git clean` no longer aborts with a fatal error when `git fetch --prune` fails (e.g. expired credentials, no network). The failure is now a warning, and the command proceeds using local refs.
- `stale_days = 0` and `commits_behind = 0` in `.forge.toml` now correctly disable their respective detection axis, as documented, instead of matching every branch.
- `config.defaults()` now sets `StaleDays: 30` and `CommitsBehind: 10` for `git.clean`, matching the CLI flag defaults. Previously these were left unset, so projects without a `.forge.toml` silently ran with age/behind detection disabled.
- BEHIND column in `forge git clean` output now shows `-` instead of a misleading `0` when `commits_behind` is disabled.
- Branch names containing `|` no longer break `forge git clean` parsing (switched to NUL-delimited `for-each-ref` output).

## [1.6.0] - 2026-07-10

### Added
- `.forge.toml` — added `[ENV.EXAMPLE]` with `invalid_keys` property
- `forge env example` — `invalid_keys` config option under `[env.example]` — controls handling of invalid keys during `.env.example` generation (`"warn"` | `"skip"` | `"include"`)
- `InvalidKeysMode` — named type with constants `InvalidKeysWarn`, `InvalidKeysSkip`, `InvalidKeysInclude` in `internal/env`
- `invalidKeyReason` — helper for human-readable invalid key diagnostics
- `forge config` — Added `GitAdd` struct that will hold config of `forge git add`

### Changed
- `forge env check` — rearrange checks
- `forge env example` — changed hard coded `.env` name to be parsed in example to support custom file names defined through `.forge.toml`
- `ParseEnv` — signature now accepts `invalidKeys string` instead of operating without config
- Invalid key handling replaces old commented-out `continue` blocks with explicit mode-driven branching
- Duplicate key detection moved inside `ParseEnv` with consistent `[warn]` formatting
- `EnvExample` — struct added to config with `invalid_keys` field and `"warn"` as default
- `forge git commit` — Renamed `CommitConfig` to `GitCommit` and `CleanConfig` to `GitClean`
- `forge repo init` — the command now stages added files and does not commit them

### Removed
- `forge env init` — removed; insufficient standalone value over `cp .env.example .env`

## [1.5.1] - 2026-05-16

### Added
- `forge env check` — `consecutive_blank_lines` now reads `max_consecutive_blanks` from `[env.check]` in `.forge.toml`
- `forge env check` — `value has leading whitespace` check introduced — errors when a value starts with a space or tab after `=`
- `forge env check` — `lowercase_key` added as a valid `ignore_codes` entry
- `forge env check` — `LowercaseKey` respects `cfg.AllowedLowercase` — exempted keys return no issue
- `forge env check` — `file_ends_with_blank` added as a valid `ignore_codes` entry
- `forge env check` — `ignore_keys` support — keys listed in `[env.check] ignore_keys` are skipped entirely including commented key warns and conformity diff
- `forge env check` — `required_keys` support — keys listed in `[env.check] required_keys` error if missing from `.env` after full scan
- `forge env check` — `checkExample` skip empty lines and validate keys

### Changed
- `forge env check` — consecutive blank lines fires once per run with accurate range — message now reads `N consecutive blank lines (lines X–Y)`
- `forge env check` — `trimmedKey` declaration moved after `!found` guard — avoids trimming garbage when no `=` is present
- `forge env check` — `KeyStartsWithDigit` and `KeyInvalidChars` cases now use direct append — consistent with error handling pattern
- `forge env check` — `LowercaseKey` extracted as a rule function — goes through `ShouldAdd` with `"lowercase_key"` code
- `forge env check` — `DuplicateKey`, `KeyContainsSpace`, `ValueLeadingSpace`, `NoEqualSign`, `FileEndsWithBlank`, `EmptyKey` inlined — pure struct wraps with no logic removed from `rules.go`
- `forge env check` — `example_has_value` and `example_conformity` now respect `ignore_codes` and `--level` flag via `ShouldAdd`
- `forge env check` — `add()` closure removed — all issue appends now use direct struct literals
- `forge env check` — sort comparator extracted to `IssuesByLine` — reusable across future callers
- `forge env check` — `ValidateKey` now uses an allowlist `[A-Za-z0-9_]` instead of a blocklist — catches `.`, `-`, `/`, `:` and any other non-standard character
- `forge env check` — Invalid character check now runs before lowercase check — ensures `KeyInvalidChars` takes priority over `KeyIsLowercase`
- `forge env check` — exits with code 1 when issues are found — previously always exited 0, making it invisible to CI pipelines

### Removed
- `forge env check` — `enforce_export` check removed — `export` prefix breaks Docker `--env-file` and most dotenv libraries
- `.forge.toml` — `enforce_export` key removed from `[env.check]` config

### Fixed
- `forge env check` — consecutive blank lines no longer fires repeatedly across a single blank run
- `forge env check` — consecutive blank lines no longer fires when `max_consecutive_blanks = 0`
- `forge env check` — `consecutiveBlankLines` counter typed as `uint8`, wraps at 256 — changed to `int32`
- `forge env check` — consecutive blank lines at EOF never flushed — post-loop flush added
- `forge env check` — `max_consecutive_blanks = 0` silently disabled the check — condition changed to `>= 0`, use `-1` to disable
- `forge env check` — `file_ends_with_blank` false positive on CRLF files — `TrimRight` now strips `\r\n` instead of `\n` only
- `forge env check` — `file_ends_with_blank` never fired — `lastLine` (formerly `previousLine`) was assigned at the bottom of the loop after all `continue` statements, skipping blank lines entirely — moved to top of loop
- `forge env check` — post-loop consecutive blank lines message reported `lineNum-1` as end of run — corrected to `lineNum`
- `forge env check` — `ValidateKey` called twice per key in `ParseKeysFromExample` — collapsed to single assignment
- `forge env check` — `export` stripping in `ParseKeysFromExample` trimmed leading whitespace before the prefix, diverging from `check.go` behaviour — aligned to `strings.TrimPrefix(line, "export ")`

## [1.5.0] - 2026-05-11

### Added
- `forge config new` — Path existence and directory validation
- `forge config new` — Soft `.git/` reminder printed after successful creation when no `.git/` is detected at target path
- `forge repo` — Added `ResolveCaseInsensitive` to perform case insensitive search for a file and return the actual appearance name
- `forge env check` — `.env.example` conformity now detects keys with values set — warns when example file contains non-empty values
- `forge env check` — `ExampleKey` struct introduced to carry `HasValue` metadata from `parseKeysFromExample`
- `forge env check` — `CheckIssue` now carries a `File` field — allows issues to reference `.env.example` path instead of always printing `.env`
- `forge env check` — `shouldAdd` helper introduced — centralizes severity level filtering and ignore code checks at extracted rule call sites
- `forge env check` — `emptyKey`, `emptyValue`, `commentedHasValue`, `validateValue` extracted as standalone rule functions — each returns `*CheckIssue` for testability
- `forge env` — `[env]` section added to `.forge.toml` — `default_file` and `example_file` configurable at project level
- `forge env add` — `[env.add]` config block introduced — `export_prefix` and `line_ending` fields
- `forge env check` — `[env.check]` config block introduced — `check_level`, `ignore_keys`, `ignore_codes`, `required_keys`, `allowed_lowercase`, `max_consecutive_blanks`, `enforce_export` fields
- `forge env check` — `ignore_keys` supports wildcard prefix matching — e.g. `DB_*` skips all keys with that prefix
- `forge env check` — `required_keys` — Forge errors if any listed key is absent from `.env`
- `forge env check` — `allowed_lowercase` — exempts specific lowercase keys from the lowercase warning
- `forge env check` — `max_consecutive_blanks` — configurable blank line tolerance, set to `0` to disable
- `config` — `EnvConfig`, `EnvAdd`, `EnvCheck` structs introduced with sensible defaults

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
- `forge env check` — Replaced `strings.Split` with `bufio.NewScanner` + `bytes.NewReader` for line reading
- `forge env check` — Replaced all `strings.Index` key/value splits with `strings.Cut`
- `forge env check` — `parseKeysFromExample` migrated to `bufio.NewScanner` + `bytes.NewReader` + `strings.Cut`
- `forge env check` — `.env.example` path now derived from the directory of the checked `.env` file — fixes resolution when invoked outside the project root
- `forge env check` — All errors wrapped with `fmt.Errorf("[env check]: %w", err)`
- `forge env check` — Issues now printed to `os.Stderr`, success message remains on `os.Stdout`
- `forge env check` — Empty key guard added to commented line parser — prevents false positives on separator lines containing `=`
- `forge env check` — Invalid keys (`KeyStartsWithDigit`, `KeyInvalidChars`) excluded from `seen` map — prevents noise in conformity diff
- `forge env check` — `validateValue` consolidates unclosed quote and unquoted spaces checks — both share the same quoted/unquoted branch logic
- `forge env check` — `commentedHasValue` message now includes the key name — improves diagnostic precision
- `forge env check` — `validateValue` messages now reference key name instead of value

### Fixed
- **The "Duplicate Config" Bug**: Fixed an issue on Linux where overwriting a case-mismatched file (e.g., `.FORGE.toml`) would create a second file (`.forge.toml`) instead of replacing the original.
- `forge env check` — Separator comment lines (e.g. `# ===`) no longer trigger false `commented key "" has a value` warnings
- `forge env check` — Invalid keys no longer appear in `.env.example` conformity warnings
- `forge env check` — Extracted rule functions now respect `--level` flag — previously bypassed `add()` severity gate and always appended regardless of level
- `forge env check` — Spurious `continue` removed from `emptyValue` call site — previously skipped `validateValue` when `empty_value` was ignored or filtered by level

### Removed
- `forge env` — Removed `add_test`, `check_test`, `init_test`, `example_test` — test coverage to be rewritten against the new config-aware rule functions

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

### Removed
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