package git

import (
	"testing"

	"github.com/GuechtouliAnis/forge/internal/config"
)

func TestCreatePattern(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.CommitConfig
		wantPattern string
		wantErr     bool
	}{
		{
			name:        "both unset returns empty",
			cfg:         &config.CommitConfig{},
			wantPattern: "",
			wantErr:     false,
		},
		{
			name: "format without placeholders returns as-is",
			cfg: &config.CommitConfig{
				Format: "hello",
			},
			wantPattern: "hello",
			wantErr:     false,
		},
		{
			name: "domain in format with valid domains",
			cfg: &config.CommitConfig{
				Format:  "[{domain}] {message}",
				Domains: []string{"FIX", "FEAT", "REFACT"},
			},
			wantPattern: `\[(FIX|FEAT|REFACT)\] \{message\}`,
			wantErr:     false,
		},
		{
			name: "domain in format with no valid domains errors",
			cfg: &config.CommitConfig{
				Format:  "[{domain}] {message}",
				Domains: []string{"", "  "},
			},
			wantPattern: "",
			wantErr:     true,
		},
		{
			name: "domain in format with blank entries filtered",
			cfg: &config.CommitConfig{
				Format:  "[{domain}]",
				Domains: []string{"FIX", "", "  ", "FEAT"},
			},
			wantPattern: `\[(FIX|FEAT)\]`,
			wantErr:     false,
		},
		{
			name: "maxlen with message in format",
			cfg: &config.CommitConfig{
				Format:        "{message}",
				MessageMaxLen: 50,
			},
			wantPattern: `[^\n\r]{1,50}`,
			wantErr:     false,
		},
		{
			name: "maxlen without message in format is ignored",
			cfg: &config.CommitConfig{
				Format:        "[{domain}]",
				Domains:       []string{"FIX"},
				MessageMaxLen: 50,
			},
			wantPattern: `\[(FIX)\]`,
			wantErr:     false,
		},
		{
			name: "full format with domain and maxlen",
			cfg: &config.CommitConfig{
				Format:        "[{domain}] {message}",
				Domains:       []string{"FIX", "FEAT"},
				MessageMaxLen: 100,
			},
			wantPattern: `\[(FIX|FEAT)\] [^\n\r]{1,100}`,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreatePattern(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantPattern {
				t.Errorf("CreatePattern() = %q, want %q", got, tt.wantPattern)
			}
		})
	}
}

func TestValidateCommit(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		cfg       *config.CommitConfig
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "no constraints accepts anything",
			message:   "whatever",
			cfg:       &config.CommitConfig{},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:    "valid domain and message",
			message: "[FIX] correct commit",
			cfg: &config.CommitConfig{
				Format:        "[{domain}] {message}",
				Domains:       []string{"FIX", "FEAT"},
				MessageMaxLen: 100,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:    "invalid domain rejected",
			message: "[CHORE] something",
			cfg: &config.CommitConfig{
				Format:  "[{domain}] {message}",
				Domains: []string{"FIX", "FEAT"},
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name:    "message exceeds maxlen",
			message: "this message is way too long and should be rejected by the validator",
			cfg: &config.CommitConfig{
				Format:        "{message}",
				MessageMaxLen: 20,
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name:    "message within maxlen",
			message: "short message",
			cfg: &config.CommitConfig{
				Format:        "{message}",
				MessageMaxLen: 20,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:    "missing domain errors on pattern build",
			message: "[FIX] something",
			cfg: &config.CommitConfig{
				Format:  "[{domain}] {message}",
				Domains: []string{},
			},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateCommit(tt.message, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantValid {
				t.Errorf("ValidateCommit() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}
