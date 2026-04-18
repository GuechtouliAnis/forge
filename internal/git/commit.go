package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/GuechtouliAnis/forge/internal/config"
)

// CreatePattern builds a regex pattern from the commit config.
// Returns an empty string if no constraints are defined (both format and
// message_max_length are unset), meaning validation is skipped entirely.
func CreatePattern(cfg *config.CommitConfig) (string, error) {
	// Both fields at zero-value means the user opted out of commit validation.
	if strings.TrimSpace(cfg.Format) == "" && cfg.MessageMaxLen == 0 {
		return "", nil
	}

	// Escape all regex metacharacters in the format so literal [ ] . etc
	// are treated as plain characters, not regex syntax.
	pattern := regexp.QuoteMeta(cfg.Format)

	// {domain} in format requires at least one valid domain to build the
	// alternation group (e.g. "FIX|FEAT|REFACT"). Blank entries are filtered
	// out and treated as absent — an all-blank list is the same as no list.
	if strings.Contains(cfg.Format, "{domain}") {
		var validDomains []string
		for _, d := range cfg.Domains {
			if strings.TrimSpace(d) != "" {
				validDomains = append(validDomains, d)
			}
		}
		if len(validDomains) == 0 {
			return "", fmt.Errorf("format contains {domain} but no valid domains are defined")
		}
		domainRegex := "(" + strings.Join(validDomains, "|") + ")"
		pattern = strings.ReplaceAll(pattern, `\{domain\}`, domainRegex)
	}

	// {message} is replaced with a character class capped at MessageMaxLen.
	// Newlines and carriage returns are excluded since git commit -m treats
	// them as message delimiters. Skipped if MaxLen is 0 (no length constraint)
	// or if {message} is absent from the format.
	if cfg.MessageMaxLen > 0 && strings.Contains(cfg.Format, "{message}") {
		messageRegex := fmt.Sprintf(`[^\n\r]{1,%d}`, cfg.MessageMaxLen)
		pattern = strings.ReplaceAll(pattern, `\{message\}`, messageRegex)
	}

	// Sanity-check the assembled pattern before returning it. A malformed
	// format string (e.g. unmatched brackets) would produce an invalid regex
	// that would silently match nothing at validation time.
	if _, err := regexp.Compile(pattern); err != nil {
		return "", fmt.Errorf("resulting pattern is invalid: %w", err)
	}

	return pattern, nil
}

// ValidateCommit checks whether the given commit message satisfies the
// constraints defined in the config. If no pattern can be built (both format
// and max length unset), validation is skipped and the message is accepted.
func ValidateCommit(message string, cfg *config.CommitConfig) (bool, error) {
	pattern, err := CreatePattern(cfg)
	if err != nil {
		return false, fmt.Errorf("failed to build commit pattern: %w", err)
	}

	if pattern != "" {
		matched, err := regexp.MatchString("^"+pattern+"$", message)
		if err != nil {
			return false, fmt.Errorf("failed to match commit pattern: %w", err)
		}
		if !matched {
			return false, nil
		}
	}

	// enforce maxlen independently — covers the case where {message} is absent
	// from format or format is empty entirely
	if cfg.MessageMaxLen > 0 && !strings.Contains(cfg.Format, "{message}") {
		if len(message) > cfg.MessageMaxLen {
			return false, nil
		}
	}

	return true, nil
}
