package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GuechtouliAnis/forge/internal/config"
)

// classifyFile runs a single file through :
// - unconditional rejection,
// - blocklist matching,
// - size-threshold evaluation.
// It returns the file's fileResult and whether it should
// proceed to the staging queue.
//
// Size-threshold handling is fully config-driven via on_file_size_violation,
// there is no interactive prompt. This keeps forge git add safe to run
// non-interactively (CI, scripts, hooks), unlike a stdin-blocking confirm.
func classifyFile(f resolvedFile, cfg config.GitAdd, envDefaultFile string, dryRun bool) (fileResult, bool, error) {

	path := f.Path

	if isHardcodedBlock(path, envDefaultFile) {
		fmt.Printf("[git add]: '%s' cannot be staged via forge git add\n", path)
		return fileResult{
			Path:  path,
			State: stateBlockedHardcoded,
			Code:  f.Code}, false, nil
	}

	if matched, pattern := matchesBlocklist(path, cfg.BlocklistPatterns); matched {
		fmt.Printf("[git add]: blocked — '%s' matches pattern '%s'\n", path, pattern)
		return fileResult{
			Path:   path,
			State:  stateBlockedPattern,
			Reason: pattern,
			Code:   f.Code}, false, nil
	}

	// Deletions have nothing on disk to size-check. Once past the
	// guardrails above, staging a deletion is always approved.
	if f.Deleted {
		return fileResult{
			Path:   path,
			State:  stateStaged,
			Reason: "deleted",
			Code:   f.Code}, true, nil
	}

	sizeMB, statErr := fileSizeMB(path)
	if statErr != nil {
		return fileResult{}, false, fmt.Errorf("[git add]: path does not exist: %s", path)
	}

	exceeds := cfg.MaxFileSize != -1 && sizeMB > float64(cfg.MaxFileSize)
	if !exceeds {
		return fileResult{
			Path:   path,
			State:  stateStaged,
			SizeMB: sizeMB,
			Code:   f.Code}, true, nil
	}

	// Oversized — behavior is entirely determined by on_file_size_violation.
	switch cfg.OnFileSizeViolation {
	case "skip":
		if !dryRun {
			fmt.Printf("%s not staged — exceeds max_file_size_mb (%.1f MB).\n", path, sizeMB)
		}
		return fileResult{
			Path:   path,
			State:  stateSkippedLarge,
			SizeMB: sizeMB,
			Code:   f.Code}, false, nil

	case "fail":
		if dryRun {
			// Never abort during a preview — report what *would* happen instead.
			return fileResult{
				Path:   path,
				State:  stateFailedLarge,
				SizeMB: sizeMB,
				Code:   f.Code}, false, nil
		}
		return fileResult{}, false, fmt.Errorf(
			"[git add]: '%s' exceeds max_file_size_mb (%.1f MB) — on_file_size_violation is set to \"fail\"",
			path,
			sizeMB)

	default: // "warn" — stage anyway, but warn
		if !dryRun {
			fmt.Printf(
				"[git add]: Warning: '%s' is %.1f MB, exceeding max_file_size_mb.\n",
				path,
				sizeMB)
			fmt.Println("[git add]: Tip: raise max_file_size_mb in .forge.toml to avoid this, or set it to -1 to disable.")
		}
		return fileResult{
			Path:   path,
			State:  stateStaged,
			SizeMB: sizeMB,
			Code:   f.Code}, true, nil
	}
}

// isHardcodedBlock reports whether a file is unconditionally rejected:
// the .env file itself, or the project's configured env default file.
// These two names are rejected regardless of blocklist_patterns configuration,
// forge git add can never be the mechanism by which a secrets file is
// staged by accident.
func isHardcodedBlock(path, envDefaultFile string) bool {

	base := filepath.Base(path)

	if strings.EqualFold(base, ".env") {
		return true
	}

	return envDefaultFile != "" && base == filepath.Base(envDefaultFile)
}

// matchesBlocklist tests a file's base name — not its full path — against
// each configured glob pattern, case-insensitively (so "ID_RSA" and
// "id_rsa" are treated identically). Matching on the base name keeps
// extension-style patterns like "*.pem" working regardless of which
// directory the file lives in. Returns the first matching pattern for use
// in the blocked-file message.
func matchesBlocklist(path string, patterns []string) (ok bool, pattern string) {

	base := strings.ToLower(filepath.Base(path))

	for _, p := range patterns {
		if matched, err := filepath.Match(strings.ToLower(p), base); err == nil && matched {
			return true, p
		}
	}
	return false, ""
}

// fileSizeMB stats a file and returns its size in megabytes. Threshold
// comparison is left to the caller, since the meaning of "too large"
// depends on config (and can be disabled entirely via -1) — this function
// stays a pure, single-purpose size lookup.
func fileSizeMB(path string) (float64, error) {

	info, err := os.Stat(path)

	if err != nil {
		return 0, err
	}

	return float64(info.Size()) / (1024 * 1024), nil
}
