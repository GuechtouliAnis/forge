package config

type Config struct {
	Git GitConfig `toml:"git"`
	Env EnvConfig `toml:"env"`
}

// === ENV ===
// ENV
type EnvConfig struct {
	DefaultFile string `toml:"default_file"`
	ExampleFile string `toml:"example_file"`
	Add         EnvAdd
	Check       EnvCheck
	Example     EnvExample
}

// ENV ADD
type EnvAdd struct {
	LineEnding string `toml:"line_ending"`
}

// ENV CHECK
type EnvCheck struct {
	CheckLevel       string   `toml:"check_level"`
	IgnoreKeys       []string `toml:"ignore_keys"`
	IgnoreCodes      []string `toml:"ignore_codes"`
	RequiredKeys     []string `toml:"required_keys"`
	AllowedLowercase []string `toml:"allowed_lowercase"`
	MaxConsBlanks    int32    `toml:"max_consecutive_blanks"`
}

// ENV EXAMPLE
type EnvExample struct {
	InvalidKeys string `toml:"invalid_keys"`
}

// === GIT ===
// GIT
type GitConfig struct {
	Commit GitCommit `toml:"commit"`
	Clean  GitClean  `toml:"clean"`
	Add    GitAdd    `toml:"add"`
}

// GIT ADD
type GitAdd struct {
	MaxFileSize         int      `toml:"max_file_size_mb"`
	OnFileSizeViolation string   `toml:"on_file_size_violation"`
	BlocklistPatterns   []string `toml:"blocklist_patterns"`
}

// GIT COMMIT
type GitCommit struct {
	Format                 string   `toml:"format"`
	Domains                []string `toml:"domain"`
	MessageMaxLen          int      `toml:"message_max_length"`
	DomainCaseSensitive    bool     `toml:"domaine_case_sensitive"`
	StagedIgnoredFilesMode string   `toml:"staged_ignored_files_mode"`
}

// GIT CLEAN
type GitClean struct {
	StaleDays           int      `toml:"stale_days"`
	CommitsBehind       int      `toml:"commits_behind"`
	FetchRemote         bool     `toml:"fetch_remote"`
	FetchTimeoutSeconds int      `toml:"fetch_timeout"`
	MaxWorkers          int      `toml:"max_workers"`
	ProtectedBranches   []string `toml:"protected_branches"`
}

// === DEFAULT VALUES ===
func defaults() *Config {
	return &Config{

		// GIT
		Git: GitConfig{
			Add: GitAdd{
				MaxFileSize:         -1,
				OnFileSizeViolation: "warn",
			},
			Clean: GitClean{
				FetchRemote:         true,
				FetchTimeoutSeconds: 30,
				MaxWorkers:          0,
			},
			Commit: GitCommit{
				DomainCaseSensitive:    false,
				StagedIgnoredFilesMode: "warn",
			},
		},

		// ENV
		Env: EnvConfig{
			DefaultFile: ".env",
			ExampleFile: ".env.example",
			Add: EnvAdd{
				LineEnding: "lf",
			},
			Check: EnvCheck{
				CheckLevel:    "warn",
				MaxConsBlanks: -1,
			},
			Example: EnvExample{
				InvalidKeys: "warn",
			},
		},
	}
}
