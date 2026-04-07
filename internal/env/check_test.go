package env

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(testDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal("failed to write file:", err)
	}
	return path
}

func TestCheckEnv(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		level    int
		wantMsgs []string
	}{
		{
			name:     "valid file",
			content:  "API_KEY=secret\nDB_HOST=localhost",
			level:    LevelWarn,
			wantMsgs: []string{},
		},
		{
			name:     "duplicate key",
			content:  "API_KEY=secret\nAPI_KEY=other",
			level:    LevelWarn,
			wantMsgs: []string{"duplicate key"},
		},
		{
			name:     "malformed line",
			content:  "MALFORMED",
			level:    LevelWarn,
			wantMsgs: []string{"malformed line"},
		},
		{
			name:     "key starts with digit",
			content:  "1API=secret",
			level:    LevelWarn,
			wantMsgs: []string{"key starts with digit"},
		},
		{
			name:     "key contains invalid chars",
			content:  "API$KEY=secret",
			level:    LevelWarn,
			wantMsgs: []string{"key contains invalid characters"},
		},
		{
			name:     "lowercase key",
			content:  "api_key=secret",
			level:    LevelWarn,
			wantMsgs: []string{"key contains lowercase"},
		},
		{
			name:     "empty value",
			content:  "API_KEY=",
			level:    LevelWarn,
			wantMsgs: []string{"empty value"},
		},
		{
			name:     "unclosed double quote",
			content:  "API_KEY=\"secret",
			level:    LevelWarn,
			wantMsgs: []string{"unclosed quote"},
		},
		{
			name:     "unclosed single quote",
			content:  "API_KEY='secret",
			level:    LevelWarn,
			wantMsgs: []string{"unclosed quote"},
		},
		{
			name:     "unquoted value with spaces",
			content:  "API_KEY=hello world",
			level:    LevelWarn,
			wantMsgs: []string{"unquoted value contains spaces"},
		},
		{
			name:     "key with spaces (API = KEY)",
			content:  "API = KEY",
			level:    LevelWarn,
			wantMsgs: []string{"key contains spaces"},
		},
		{
			name:     "trailing whitespace",
			content:  "API_KEY=secret   ",
			level:    LevelWarn,
			wantMsgs: []string{"trailing whitespace"},
		},
		{
			name:     "consecutive blank lines",
			content:  "API_KEY=secret\n\n\nDB_HOST=localhost",
			level:    LevelWarn,
			wantMsgs: []string{"consecutive blank lines"},
		},
		{
			name:     "ending blank line",
			content:  "API_KEY=secret\n\n",
			level:    LevelWarn,
			wantMsgs: []string{"file ends with blank line"},
		},
		{
			name:     "error level hides warns",
			content:  "api_key=secret\nMALFORMED",
			level:    LevelError,
			wantMsgs: []string{"malformed line"},
		},
		{
			name:     "error level hides lowercase warn",
			content:  "api_key=secret",
			level:    LevelError,
			wantMsgs: []string{},
		},
		{
			name:     "export prefix stripped",
			content:  "export API_KEY=secret",
			level:    LevelWarn,
			wantMsgs: []string{},
		},
		{
			name:     "comment lines skipped",
			content:  "# this is a comment\nAPI_KEY=secret",
			level:    LevelWarn,
			wantMsgs: []string{},
		},
		{
			name:     "empty value with inline comment",
			content:  "EMPTY_WITH_COMMENT= # no value here",
			level:    LevelWarn,
			wantMsgs: []string{"empty value"},
		},
		{
			name:     "commented line with value",
			content:  "# COMMENTED_WITH_VALUE= value # no value here",
			level:    LevelWarn,
			wantMsgs: []string{"commented key has value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFile(t, ".env_check_test", tt.content)
			defer os.Remove(path)

			issues, err := CheckEnv(path, tt.level)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			for _, want := range tt.wantMsgs {
				found := false
				for _, issue := range issues {
					if contains(issue.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected message containing %q, not found in issues: %v", want, issues)
				}
			}

			if len(tt.wantMsgs) == 0 && len(issues) > 0 {
				t.Errorf("expected no issues, got: %v", issues)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
