package env

import (
	"os"
	"strings"
	"testing"
)

func TestAddEnv(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		selected    []string
		wantContain []string
		wantAbsent  []string
		wantErr     bool
	}{
		{
			name:        "add db preset",
			initial:     "",
			selected:    []string{"db"},
			wantContain: []string{"# db - added by forge env add", "DB_HOST=localhost", "DB_PORT=5432", `DB_NAME=""`, `DB_USER=""`, `DB_PASSWORD=""`},
		},
		{
			name:        "add ai preset",
			initial:     "",
			selected:    []string{"ai"},
			wantContain: []string{"# ai - added by forge env add", `OPENAI_API_KEY=""`, `ANTHROPIC_API_KEY=""`},
		},
		{
			name:        "add multiple presets",
			initial:     "",
			selected:    []string{"db", "redis"},
			wantContain: []string{"# db - added by forge env add", "DB_HOST=localhost", "# redis - added by forge env add", "REDIS_HOST=localhost", "REDIS_PORT=6379"},
		},
		{
			name:        "skip existing key",
			initial:     "DB_HOST=myhost\n",
			selected:    []string{"db"},
			wantContain: []string{"DB_PORT=5432"},
			wantAbsent:  []string{"DB_HOST=localhost"},
		},
		{
			name:     "all keys exist returns error",
			initial:  "DB_HOST=h\nDB_PORT=p\nDB_NAME=n\nDB_USER=u\nDB_PASSWORD=pw\n",
			selected: []string{"db"},
			wantErr:  true,
		},
		{
			name:        "new line added before preset if file has no trailing newline",
			initial:     "EXISTING=value",
			selected:    []string{"redis"},
			wantContain: []string{"EXISTING=value", "# redis - added by forge env add"},
		},
		{
			name:        "host vars get default values",
			initial:     "",
			selected:    []string{"redis"},
			wantContain: []string{"REDIS_HOST=localhost", "REDIS_PORT=6379"},
		},
		{
			name:        "non-host vars get empty string",
			initial:     "",
			selected:    []string{"db"},
			wantContain: []string{`DB_PASSWORD=""`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFile(t, ".env_add_test", tt.initial)
			defer os.Remove(path)

			err := AddEnv(path, tt.selected)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			result, err := os.ReadFile(path)
			if err != nil {
				t.Fatal("failed to read file:", err)
			}
			content := string(result)

			for _, want := range tt.wantContain {
				if !strings.Contains(content, want) {
					t.Errorf("expected %q in file, not found\ncontent:\n%s", want, content)
				}
			}

			for _, absent := range tt.wantAbsent {
				if strings.Contains(content, absent) {
					t.Errorf("expected %q to be absent, but found it\ncontent:\n%s", absent, content)
				}
			}
		})
	}
}
