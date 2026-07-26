package git

import (
	"fmt"
	"regexp"
	"strings"

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
			trimmed := strings.TrimSpace(d)
			if trimmed != "" {
				validDomains = append(validDomains, regexp.QuoteMeta(trimmed))
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
