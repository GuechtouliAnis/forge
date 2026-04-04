package env

import (
	"os"
	"testing"
)

const testDir = "/tmp/forge"

func TestMain(m *testing.M) {
	os.MkdirAll(testDir, 0755)
	code := m.Run()
	os.RemoveAll(testDir)
	os.Exit(code)
}

// ParseEnv
func TestParseEnv(t *testing.T) {

	// Assert non existent path detected
	_, err := ParseEnv("/tmp/forge/nonexistent.env")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}

	// Create the file for tests
	content := "# ENV\nENV=DEV\n# API KEY\nexport API_KEY= aa11b895&é&&btqsf"

	tmp, err := os.CreateTemp(testDir, ".env*")
	if err != nil {
		t.Fatal("Error creating tmp/forge/.env file.", err)
	}
	defer os.Remove(tmp.Name())

	tmp.WriteString(content)
	tmp.Close()

	// Reading file content through ParseEnv
	file_content, err := ParseEnv(tmp.Name())
	if err != nil {
		t.Fatal("Expected nothing, got error: ", err)
	}

	want := "# ENV\nENV=\n# API KEY\nAPI_KEY="
	if file_content != want {
		t.Errorf("got %q, want %q", file_content, want)
	}
}

// ValidateKey
func TestValidateKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"valid key", "API_KEY", KeyValid},
		{"starts with digit", "1API_KEY", KeyStartsWithDigit},
		{"contains invalid chars", "API$KEY", KeyInvalidChars},
		{"lowercase", "api_key", KeyIsLowercase},
		{"mixed case", "Api_Key", KeyIsLowercase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateKey(tt.input)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

// TransformLine
func TestTransformLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// blank and comment lines
		{"empty line", "", ""},
		{"whitespace only", "   ", ""},
		{"comment line", "# comment", "# comment"},
		{"indented comment", "  # comment", "  # comment"},

		// basic key=value
		{"basic key", "API_KEY=secret", "API_KEY="},
		{"no value", "API_KEY=", "API_KEY="},

		// unquoted with inline comment
		{"unquoted with comment", "API_KEY=secret # comment", "API_KEY=  # comment"},
		{"unquoted comment no space", "API_KEY=secret#comment", "API_KEY=  #comment"},

		// quoted values
		{"double quoted", `API_KEY="secret"`, "API_KEY="},
		{"single quoted", "API_KEY='secret'", "API_KEY="},
		{"quoted with comment", `API_KEY="secret" # comment`, "API_KEY=  # comment"},

		// quoted with # or = inside value
		{"double quoted # inside value", `API_KEY="se#cret"`, "API_KEY="},
		{"single quoted # inside value", `API_KEY='se#cret'`, "API_KEY="},
		{"quoted # inside value", `API_KEY='se#cret' # comment`, "API_KEY=  # comment"},
		{"quoted # inside value with comment", `API_KEY='se=cret' # comment`, "API_KEY=  # comment"},
		{"double quoted = inside value", `API_KEY="se=cret"`, "API_KEY="},
		{"single quoted = inside value", `API_KEY='se=cret'`, "API_KEY="},

		// malformed
		{"no equals", "MALFORMED", ""},
		{"KEY=NAME=VALUE", "KEY=NAME=VALUE", "KEY="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformLine(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
