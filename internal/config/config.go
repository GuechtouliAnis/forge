package config

type Config struct {
	Git GitConfig `toml:"git"`
}

type GitConfig struct {
	Commit CommitConfig `toml:"commit"`
}

type CommitConfig struct {
	Format        string   `toml:"format"`
	Domains       []string `toml:"domain"`
	MessageMaxLen int      `toml:"message_max_length"`
}

func defaults() *Config {
	return &Config{}
}
