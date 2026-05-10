package config

type Config struct {
	Git GitConfig `toml:"git"`
	Env EnvConfig `toml:"env"`
}

// ENV
type EnvConfig struct {
	DefaultFile string `toml:"default_file"`
	ExampleFile string `toml:"example_file"`
	Add         EnvAdd
	Check       EnvCheck
}

// ENV ADD
type EnvAdd struct {
	ExportPrefix bool   `toml:"export_prefix"`
	LineEnding   string `toml:"line_ending"`
}

// ENV CHECK
type EnvCheck struct {
	CheckLevel       string   `toml:"check_level"`
	IgnoreKeys       []string `toml:"ignore_keys"`
	IgnoreCodes      []string `toml:"ignore_codes"`
	RequiredKeys     []string `toml:"required_keys"`
	AllowedLowercase []string `toml:"allowed_lowercase"`
	MaxConsBlanks    int      `toml:"max_consecutive_blanks"`
	EnforceExport    bool     `toml:"enforce_export"`
}

// GIT
type GitConfig struct {
	Commit CommitConfig `toml:"commit"`
	Clean  CleanConfig  `toml:"clean"`
}

// GIT COMMIT
type CommitConfig struct {
	Format        string   `toml:"format"`
	Domains       []string `toml:"domain"`
	MessageMaxLen int      `toml:"message_max_length"`
}

// GIT CLEAN
type CleanConfig struct {
	StaleDays     int `toml:"stale_days"`
	CommitsBehind int `toml:"commits_behind"`
}

func defaults() *Config {
	return &Config{
		Env: EnvConfig{
			DefaultFile: ".env",
			ExampleFile: ".env.example",
			Add: EnvAdd{
				ExportPrefix: false,
				LineEnding:   "lf",
			},
			Check: EnvCheck{
				CheckLevel:    "warn",
				MaxConsBlanks: 1,
				EnforceExport: false,
			},
		},
	}
}
