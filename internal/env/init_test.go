package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	return string(data)
}

func TestInitEnv(t *testing.T) {
	tests := []struct {
		name               string
		targetExists       bool    // whether a .env file already exists at the target path before calling InitEnv
		exampleContent     *string // content to write to .env.example; nil means no .env.example file
		updateGitignore    bool    // whether to pass true for the updateGitignoreFile parameter
		wantErr            string  // substring expected in the error message; empty means no error expected
		wantFileContent    string  // expected content of the created .env file
		wantGitignoreEntry bool    // whether targetPath should appear as a line in .gitignore after the call
	}{
		{
			// InitEnv must refuse to overwrite an existing .env to prevent accidental secret loss.
			name:         "target already exists returns error",
			targetExists: true,
			wantErr:      "already exists",
		},
		{
			// When .env.example is present, InitEnv should copy its content verbatim into the new .env file.
			name:               "creates from example when present",
			exampleContent:     strPtr("API_KEY=secret\nDB_HOST=localhost"),
			updateGitignore:    true,
			wantFileContent:    "API_KEY=secret\nDB_HOST=localhost",
			wantGitignoreEntry: true,
		},
		{
			// When .env.example is absent, InitEnv should still create an empty .env file
			// rather than returning an error.
			name:               "creates empty file when no example",
			exampleContent:     nil,
			updateGitignore:    true,
			wantFileContent:    "",
			wantGitignoreEntry: true,
		},
		{
			// When updateGitignoreFile is false, InitEnv should skip the .gitignore update entirely.
			// The .env file should still be created successfully.
			name:               "skips gitignore update when flag is false",
			exampleContent:     nil,
			updateGitignore:    false,
			wantFileContent:    "",
			wantGitignoreEntry: false,
		},
		{
			// Content should be copied byte-for-byte; InitEnv must not strip or transform
			// "export" prefixes or any other shell syntax present in the example file.
			name:               "example with export prefixes copied verbatim",
			exampleContent:     strPtr("export API_KEY=\nexport DB_PASS="),
			updateGitignore:    true,
			wantFileContent:    "export API_KEY=\nexport DB_PASS=",
			wantGitignoreEntry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fresh temp directory per subtest so file state never leaks between cases.
			dir := t.TempDir()
			origDir, _ := os.Getwd()
			// InitEnv looks up ".env.example" and ".gitignore" relative to CWD,
			// so we must chdir into the temp directory before calling it.
			if err := os.Chdir(dir); err != nil {
				t.Fatal("failed to chdir:", err)
			}
			defer os.Chdir(origDir)

			targetPath := filepath.Join(dir, ".env")

			// Pre-create the target file if the test case requires it to already exist.
			if tt.targetExists {
				if err := os.WriteFile(targetPath, []byte("existing"), 0644); err != nil {
					t.Fatal("setup: failed to write target file:", err)
				}
			}

			// Write .env.example only when the test case supplies content for it.
			if tt.exampleContent != nil {
				if err := os.WriteFile(".env.example", []byte(*tt.exampleContent), 0644); err != nil {
					t.Fatal("setup: failed to write .env.example:", err)
				}
			}

			err := InitEnv(targetPath, tt.updateGitignore)

			// --- Error path ---
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				// Nothing else to assert when an error is expected.
				return
			}

			// --- Happy path ---
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the .env file was written with the expected content.
			got := readFile(t, targetPath)
			if got != tt.wantFileContent {
				t.Errorf("file content = %q, want %q", got, tt.wantFileContent)
			}

			// Verify .gitignore state.
			gitignorePath := filepath.Join(dir, ".gitignore")
			if tt.wantGitignoreEntry {
				// The file must exist and contain targetPath as its own line.
				if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
					t.Errorf("expected .gitignore to exist, but it does not")
				} else {
					content := readFile(t, gitignorePath)
					hasEntry := false
					for _, line := range strings.Split(content, "\n") {
						if strings.TrimSpace(line) == targetPath {
							hasEntry = true
							break
						}
					}
					if !hasEntry {
						t.Errorf("expected .gitignore to contain %q, got:\n%s", targetPath, content)
					}
				}
			} else {
				// If .gitignore was created at all, targetPath must not appear in it.
				if _, err := os.Stat(gitignorePath); err == nil {
					content := readFile(t, gitignorePath)
					for _, line := range strings.Split(content, "\n") {
						if strings.TrimSpace(line) == targetPath {
							t.Errorf("expected .gitignore NOT to contain %q, but it does", targetPath)
						}
					}
				}
			}
		})
	}
}

func TestUpdateGitignore(t *testing.T) {
	tests := []struct {
		name            string
		existingContent *string // content to pre-populate .gitignore with; nil means no file exists yet
		entry           string  // the path to pass to updateGitignore
		wantPresent     bool    // whether entry should appear in .gitignore after the call
		wantCreated     bool    // whether .gitignore should be created from scratch (was absent before)
	}{
		{
			// When no .gitignore exists, the function should create one containing the entry.
			name:        "creates gitignore when absent",
			entry:       ".env",
			wantCreated: true,
			wantPresent: true,
		},
		{
			// Entry should be appended without disturbing pre-existing lines.
			name:            "appends to existing gitignore",
			existingContent: strPtr("node_modules/\ndist/"),
			entry:           ".env",
			wantPresent:     true,
		},
		{
			// If the entry is already in the file, the function should not write it again.
			// wantPresent is true because the entry should still be there — just not added a second time.
			name:            "skips when entry already present",
			existingContent: strPtr("node_modules/\n.env\ndist/"),
			entry:           ".env",
			wantPresent:     true,
		},
		{
			// When the file exists but has no trailing newline, the function must insert one
			// before appending the entry so it appears on its own line.
			name:            "appends newline before entry when file has no trailing newline",
			existingContent: strPtr("node_modules/"),
			entry:           ".env",
			wantPresent:     true,
		},
		{
			// Calling the function when the entry is already the only line (with trailing newline)
			// should be a no-op — the entry must not be duplicated.
			name:            "does not duplicate entry on repeated calls",
			existingContent: strPtr(".env\n"),
			entry:           ".env",
			wantPresent:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			origDir, _ := os.Getwd()
			// updateGitignore resolves ".gitignore" relative to CWD.
			if err := os.Chdir(dir); err != nil {
				t.Fatal("failed to chdir:", err)
			}
			defer os.Chdir(origDir)

			// Write pre-existing content only when the test case requires it.
			if tt.existingContent != nil {
				if err := os.WriteFile(".gitignore", []byte(*tt.existingContent), 0644); err != nil {
					t.Fatal("setup: failed to write .gitignore:", err)
				}
			}

			if err := updateGitignore(tt.entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			data, err := os.ReadFile(".gitignore")
			if err != nil {
				t.Fatal("failed to read .gitignore after call:", err)
			}
			content := string(data)

			// Count how many lines match the entry exactly (after trimming whitespace).
			count := 0
			for _, line := range strings.Split(content, "\n") {
				if strings.TrimSpace(line) == tt.entry {
					count++
				}
			}

			if tt.wantPresent && count == 0 {
				t.Errorf("expected %q in .gitignore, not found:\n%s", tt.entry, content)
			}
			// Writing the entry more than once is always a bug, regardless of wantPresent.
			if count > 1 {
				t.Errorf("entry %q appears %d times in .gitignore, want exactly 1:\n%s", tt.entry, count, content)
			}
		})
	}
}

// strPtr is a convenience helper that returns a pointer to a string literal.
// Used in test tables where *string fields distinguish "not set" (nil) from "empty string" ("").
func strPtr(s string) *string {
	return &s
}
