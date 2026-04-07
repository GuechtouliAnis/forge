package env

// Predefined variable sets for forge env add
var presets = map[string][]string{
	"db": {
		"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
	},
	"ai": {
		// generic
		"AI_MODEL", "AI_BASE_URL",
		// openai
		"OPENAI_API_KEY",
		// google
		"GOOGLE_API_KEY", "GOOGLE_PROJECT_ID", "GOOGLE_LOCATION",
		// anthropic
		"ANTHROPIC_API_KEY",
		// ollama
		"OLLAMA_HOST",
	},
	"web": {
		"APP_HOST", "APP_PORT", "APP_SECRET_KEY",
		// fastapi specific
		"APP_ENV", "APP_DEBUG", "ALLOWED_ORIGINS",
	},
	"redis": {
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
	},
	"monitoring": {
		// grafana
		"GRAFANA_HOST", "GRAFANA_PORT", "GRAFANA_USER", "GRAFANA_PASSWORD",
		// prometheus
		"PROMETHEUS_HOST", "PROMETHEUS_PORT",
	},
	"neo4j": {
		"NEO4J_URI", "NEO4J_USER", "NEO4J_PASSWORD",
	},
}

// Default values
var hostVars = map[string]string{
	// hosts
	"DB_HOST":         "localhost",
	"APP_HOST":        "localhost",
	"REDIS_HOST":      "localhost",
	"GRAFANA_HOST":    "localhost",
	"PROMETHEUS_HOST": "localhost",
	"NEO4J_URI":       "bolt://localhost:7687",
	"OLLAMA_HOST":     "http://localhost:11434",
	// ports
	"DB_PORT":         "5432", // postgres default
	"APP_PORT":        "8000", // fastapi default
	"REDIS_PORT":      "6379",
	"GRAFANA_PORT":    "3000",
	"PROMETHEUS_PORT": "9090",
}
