package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/GuechtouliAnis/forge/internal/config"
)

// CommitValidation is the tri-state outcome of validating a commit message:
// accepted outright, accepted with a domain-case-mismatch warning, or
// rejected. ExpectedDomain/GotDomain are only populated when CaseWarning
// is true, and are used to build the "(expected X, got Y)" warning text.
type CommitValidation struct {
	Valid          bool
	CaseWarning    bool
	ExpectedDomain string
	GotDomain      string
}

// CreatePattern builds a regex pattern from the commit config. Returns an
// empty pattern if no constraints are defined (both format and
// message_max_length are unset), meaning validation is skipped entirely.
//
// hasDomain reports whether {domain} appears in the format, which callers
// use to decide whether a case-insensitive fallback attempt makes sense at
// all — there is no "domain casing" to fall back on if there is no domain
// segment in the first place.
//
// domainCaseInsensitive scopes case-insensitivity to ONLY the domain
// alternation, via Go regexp's inline group flag `(?i:...)`, rather than
// applying it to the pattern as a whole. This is deliberate: the config
// option is named after "domain" specifically, and blanket
// case-insensitivity across the entire message would have unintended
// effects on any literal text in the configured format.
//
// The domain group itself is wrapped in an outer capturing group (not
// left non-capturing) so ValidateCommit can recover exactly what the
// caller typed in that position, which is needed to build a precise
// "expected X, got Y" message on the mismatch-warning path.
func CreatePattern(cfg *config.GitCommit, domainCaseInsensitive bool) (pattern string, hasDomain bool, err error) {
	// Both fields at zero-value means the user opted out of commit validation.
	if strings.TrimSpace(cfg.Format) == "" && cfg.MessageMaxLen == 0 {
		return "", false, nil
	}

	// Escape all regex metacharacters in the format so literal [ ] . etc.
	// are treated as plain characters, not regex syntax.
	pattern = regexp.QuoteMeta(cfg.Format)

	// {domain} in format requires at least one valid domain to build the
	// alternation group (e.g. "FIX|FEAT|REFACT"). Blank entries are
	// filtered out and treated as absent — an all-blank list is the same
	// as no list.
	if strings.Contains(cfg.Format, "{domain}") {
		var validDomains []string
		for _, d := range cfg.Domains {
			if strings.TrimSpace(d) != "" {
				validDomains = append(validDomains, d)
			}
		}
		if len(validDomains) == 0 {
			return "", false, fmt.Errorf("format contains {domain} but no valid domains are defined")
		}

		alternation := strings.Join(validDomains, "|")
		var domainRegex string
		if domainCaseInsensitive {
			// Outer ( ) captures; inner (?i: ) scopes insensitivity to
			// just this group, leaving the rest of the pattern untouched.
			domainRegex = "((?i:" + alternation + "))"
		} else {
			domainRegex = "(" + alternation + ")"
		}
		pattern = strings.ReplaceAll(pattern, `\{domain\}`, domainRegex)
		hasDomain = true
	}

	// {message} is replaced with a character class capped at MessageMaxLen.
	// Newlines and carriage returns are excluded since git commit -m treats
	// them as message delimiters. Skipped if MaxLen is 0 (no length
	// constraint) or if {message} is absent from the format.
	if cfg.MessageMaxLen > 0 && strings.Contains(cfg.Format, "{message}") {
		messageRegex := fmt.Sprintf(`[^\n\r]{1,%d}`, cfg.MessageMaxLen)
		pattern = strings.ReplaceAll(pattern, `\{message\}`, messageRegex)
	}

	// Sanity-check the assembled pattern before returning it. A malformed
	// format string (e.g. unmatched brackets) would produce an invalid
	// regex that would silently match nothing at validation time.
	if _, compileErr := regexp.Compile(pattern); compileErr != nil {
		return "", false, fmt.Errorf("resulting pattern is invalid: %w", compileErr)
	}

	return pattern, hasDomain, nil
}

// validateLength enforces message_max_length independently of the
// compiled pattern. This matters for two distinct cases: when {message}
// is absent from format entirely (so the pattern never constrains length
// on its own), and when there is no pattern at all (format is blank but a
// max length is still configured).
func validateLength(message string, cfg *config.GitCommit) bool {
	if cfg.MessageMaxLen <= 0 {
		return true // length constraint disabled
	}
	if strings.Contains(cfg.Format, "{message}") {
		return true // already enforced by the compiled pattern itself
	}
	return len(message) <= cfg.MessageMaxLen
}

// canonicalDomain finds the configured domain entry that matches got
// case-insensitively, returning it in its originally configured casing —
// this is the "expected" value surfaced in the DOMAIN_CASE_MISMATCH
// warning. Falls back to got itself if no configured entry matches, which
// should not normally occur since got was extracted via the same domain
// alternation the config produced.
func canonicalDomain(got string, domains []string) string {
	for _, d := range domains {
		if strings.EqualFold(d, got) {
			return d
		}
	}
	return got
}

// ValidateCommit checks whether the given commit message satisfies the
// constraints defined in the config. If no pattern can be built (both
// format and max length unset), validation is skipped and the message is
// accepted outright.
//
// Behavior branches on cfg.DomainCaseSensitive:
//
//   - false: the domain segment is matched case-insensitively from the
//     start. A single pass decides acceptance; no warning is ever
//     produced, since case simply doesn't matter under this setting.
//   - true (default): a strict, case-sensitive pass runs first. If it
//     fails AND the format has a domain segment, a second pass retries
//     with only the domain segment relaxed to case-insensitive. Success
//     on that retry means casing was the sole problem — the message is
//     accepted with a CaseWarning. Failure on both passes is a genuine
//     rejection.
func ValidateCommit(message string, cfg *config.GitCommit) (CommitValidation, error) {
	if !cfg.DomainCaseSensitive {
		pattern, _, err := CreatePattern(cfg, true)
		if err != nil {
			return CommitValidation{}, err
		}
		if pattern == "" {
			return CommitValidation{Valid: validateLength(message, cfg)}, nil
		}
		matched, matchErr := regexp.MatchString("^"+pattern+"$", message)
		if matchErr != nil {
			return CommitValidation{}, fmt.Errorf("failed to match commit pattern: %w", matchErr)
		}
		return CommitValidation{Valid: matched && validateLength(message, cfg)}, nil
	}

	// --- Strict pass: domain matching is case-sensitive ---------------------
	strictPattern, hasDomain, err := CreatePattern(cfg, false)
	if err != nil {
		return CommitValidation{}, err
	}
	if strictPattern == "" {
		return CommitValidation{Valid: validateLength(message, cfg)}, nil
	}

	strictRe, compileErr := regexp.Compile("^" + strictPattern + "$")
	if compileErr != nil {
		return CommitValidation{}, fmt.Errorf("failed to match commit pattern: %w", compileErr)
	}
	if strictRe.MatchString(message) {
		return CommitValidation{Valid: validateLength(message, cfg)}, nil
	}

	// Strict match failed. A case-insensitive retry only makes sense if
	// there is a domain segment to relax in the first place — otherwise
	// the rejection stands as-is.
	if !hasDomain {
		return CommitValidation{Valid: false}, nil
	}

	loosePattern, _, err := CreatePattern(cfg, true)
	if err != nil {
		return CommitValidation{}, err
	}
	looseRe, compileErr := regexp.Compile("^" + loosePattern + "$")
	if compileErr != nil {
		return CommitValidation{}, fmt.Errorf("failed to match commit pattern: %w", compileErr)
	}

	submatches := looseRe.FindStringSubmatch(message)
	if submatches == nil || !validateLength(message, cfg) {
		return CommitValidation{Valid: false}, nil
	}

	got := submatches[1]
	return CommitValidation{
		Valid:          true,
		CaseWarning:    true,
		GotDomain:      got,
		ExpectedDomain: canonicalDomain(got, cfg.Domains),
	}, nil
}

// hasStagedChanges reports whether the index currently has anything
// staged for commit. Both the default and amend flows short-circuit with
// NOTHING_STAGED before any validation or git commit is attempted.
func hasStagedChanges(ctx context.Context) (bool, error) {
	out, err := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only").Output()
	if err != nil {
		return false, fmt.Errorf("[git commit]: could not check staged changes: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// commitHash returns the short hash of HEAD immediately after a commit,
// for inclusion in the success message.
func commitHash(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("[git commit]: could not read commit hash: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// stagedIgnoredFiles returns every file that is currently staged despite
// matching a .gitignore (or other standard exclude) rule. This can only
// happen via an explicit force-add (`git add -f`), since forge git add's
// own guardrails never bypass ignore rules on their own — so surfacing
// it here is a safety net for whatever staged the index, not just files
// forge itself staged.
func stagedIgnoredFiles(ctx context.Context) ([]string, error) {
	out, err := exec.CommandContext(ctx, "git", "ls-files", "-c", "-i", "--exclude-standard").Output()
	if err != nil {
		return nil, fmt.Errorf("[git commit]: could not check staged files against .gitignore: %w", err)
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// CommitGit validates and commits staged changes, or amends the previous
// commit. messageProvided must reflect whether -m was actually passed on
// the command line (via cmd.Flags().Changed("message") at the call site)
// rather than merely whether message is an empty string — this matters
// because an explicit `-m ""` is a different situation from omitting -m
// entirely, and only the latter should trigger the amend-reuse path.
//
// Mirrors AddGit/CleanGit's error convention: only the outcomes in the
// spec's error table (NO_MESSAGE, MESSAGE_REJECTED, NOTHING_STAGED,
// GIT_COMMIT_FAILED, and CreatePattern/ValidateCommit's own errors)
// surface as a non-nil error. Everything else prints its outcome message
// directly and returns nil.
func CommitGit(cfg *config.GitCommit, message string, messageProvided bool, amend bool) error {

	// Set up a cancellable context tied to OS interrupt/termination signals.
	// This allows long-running operations below to be aborted cleanly
	// mid-flight rather than leaving the terminal in an inconsistent state
	// on Ctrl+C.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)

	defer stop()

	// Verify we are actually inside a git repository.
	// Fails fast with a clear error otherwise.
	if err := gitCheck(ctx); err != nil {
		return err
	}

	// --- NO_MESSAGE: required unless amending without a new message -------
	if !messageProvided && !amend {
		return fmt.Errorf("[git commit]: a commit message is required")
	}

	staged, err := hasStagedChanges(ctx)
	if err != nil {
		return err
	}
	if !staged {
		return fmt.Errorf("[git commit]: nothing staged to commit")
	}

	// Non-fatal safety net: warn if anything staged (however it got staged)
	// matches a gitignore rule. Doesn't block the commit — this is a heads-up,
	// not a guardrail rejection like forge git add's blocklist.
	if ignored, err := stagedIgnoredFiles(ctx); err != nil {
		return err
	} else if len(ignored) > 0 {
		fmt.Printf("[git commit]: warning - staged but found in .gitignore: %s\n", strings.Join(ignored, ", "))
	}

	// --- Amend without a new message: reuse the previous message verbatim,
	// completely unvalidated, per the amend flow's step 3.
	if amend && !messageProvided {
		out, err := exec.CommandContext(ctx, "git", "commit", "--amend", "--no-edit").CombinedOutput()
		if err != nil {
			return fmt.Errorf("[git commit]: git commit failed: %s", strings.TrimSpace(string(out)))
		}
		hash, err := commitHash(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("[git commit]: amended previous commit %s\n", hash)
		return nil
	}

	// --- Validate the supplied message (new commit, or amend -m) ----------
	result, err := ValidateCommit(message, cfg)
	if err != nil {
		return fmt.Errorf("[git commit]: %w", err)
	}
	if !result.Valid {
		return fmt.Errorf("[git commit]: commit message does not match required format")
	}
	if result.CaseWarning {
		fmt.Printf("[git commit]: warning - domain case mismatch (expected %s, got %s)\n",
			result.ExpectedDomain, result.GotDomain)
	}

	// --- Run the actual git commit ------------------------------------------
	gitArgs := []string{"commit", "-m", message}
	if amend {
		gitArgs = []string{"commit", "--amend", "-m", message}
	}
	out, err := exec.CommandContext(ctx, "git", gitArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git commit]: git commit failed: %s", strings.TrimSpace(string(out)))
	}

	hash, err := commitHash(ctx)
	if err != nil {
		return err
	}

	if amend {
		fmt.Printf("[git commit]: amended previous commit %s\n", hash)
	} else {
		fmt.Printf("[git commit]: committed as %s\n", hash)
	}

	return nil
}
