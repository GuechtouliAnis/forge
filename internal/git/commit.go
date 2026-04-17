package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/GuechtouliAnis/forge/internal/config"
)

// CommitError holds a human-readable reason a commit message failed validation.
type CommitError struct {
	Reason string
}

func (e *CommitError) Error() string { return e.Reason }

// ValidateCommit checks msg against the format and rules in cfg.
// Returns nil if valid, *CommitError otherwise.
func ValidateCommit(msg string, cfg *config.CommitConfig) error {
	pattern, err := buildPattern(cfg.Format)
	if err != nil {
		// Malformed format string in .forge.toml — surface clearly.
		return fmt.Errorf("invalid commit format in .forge.toml: %w", err)
	}

	m := pattern.FindStringSubmatch(msg)
	if m == nil {
		return &CommitError{
			Reason: fmt.Sprintf(
				"commit does not match format %q\n  got: %s",
				cfg.Format, msg,
			),
		}
	}

	domain := m[pattern.SubexpIndex("domain")]
	message := m[pattern.SubexpIndex("message")]

	if !domainAllowed(domain, cfg.Domains) {
		return &CommitError{
			Reason: fmt.Sprintf(
				"unknown domain %q\n  allowed: %s",
				domain, strings.Join(cfg.Domains, ", "),
			),
		}
	}

	if cfg.MessageMaxLen > 0 && len(message) > cfg.MessageMaxLen {
		return &CommitError{
			Reason: fmt.Sprintf(
				"message too long (%d chars, max %d)",
				len(message), cfg.MessageMaxLen,
			),
		}
	}

	return nil
}

// buildPattern converts a format template like "[{domain}] {message}"
// into a compiled regex with named capture groups.
// Literal characters are escaped; placeholders become named groups.
func buildPattern(format string) (*regexp.Regexp, error) {
	// Known placeholders → named capture group patterns.
	// domain: uppercase word; message: anything to EOL.
	placeholders := map[string]string{
		"{domain}":  `(?P<domain>[A-Z]+)`,
		"{message}": `(?P<message>.+)`,
	}

	// Split on placeholders, escape the literal segments, reassemble.
	raw := format
	var b strings.Builder
	for {
		// Find the earliest placeholder in what remains.
		earliest, token := -1, ""
		for ph := range placeholders {
			if i := strings.Index(raw, ph); i != -1 && (earliest == -1 || i < earliest) {
				earliest, token = i, ph
			}
		}
		if earliest == -1 {
			// No more placeholders — escape the remainder.
			b.WriteString(regexp.QuoteMeta(raw))
			break
		}
		// Escape the literal prefix before the placeholder.
		b.WriteString(regexp.QuoteMeta(raw[:earliest]))
		// Emit the named capture group.
		b.WriteString(placeholders[token])
		raw = raw[earliest+len(token):]
	}

	return regexp.Compile(`^` + b.String() + `$`)
}

func domainAllowed(domain string, allowed []string) bool {
	for _, a := range allowed {
		if a == domain {
			return true
		}
	}
	return false
}
