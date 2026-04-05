# Changelog

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