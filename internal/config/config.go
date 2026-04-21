package config

type Config struct {
	Git GitConfig `toml:"git"`
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
	return &Config{}
}
