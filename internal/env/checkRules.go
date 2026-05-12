package env

import (
	"fmt"
	"strings"

	"github.com/GuechtouliAnis/forge/internal/config"
)

func CheckTrailingWhitespace(line string, lineNum int) *CheckIssue {
	if line != strings.TrimRight(line, " \t") {
		return &CheckIssue{Line: lineNum, Severity: LevelWarn, Message: "trailing whitespace"}
	}
	return nil
}

func CommentedHasValue(strippedLine string, lineNum int) *CheckIssue {
	if key, value, found := strings.Cut(strippedLine, "="); found {
		if key == "" {
			return nil
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if hashIdx := strings.Index(value, "#"); hashIdx >= 0 {
			value = strings.TrimSpace(value[:hashIdx])
		}
		if value != "" {
			return &CheckIssue{Line: lineNum, Severity: LevelWarn,
				Message: fmt.Sprintf("commented key %q has a value", key)}
		}
	}
	return nil
}

func EmptyValue(key string, value string, lineNum int) *CheckIssue {
	checkEmpty := strings.TrimSpace(value)
	if hashIdx := strings.Index(checkEmpty, "#"); hashIdx >= 0 {
		checkEmpty = strings.TrimSpace(checkEmpty[:hashIdx])
	}
	if checkEmpty == "" {
		return &CheckIssue{Line: lineNum, Severity: LevelWarn,
			Message: fmt.Sprintf("empty value for key: %q", key)}
	}
	return nil
}

func EmptyKey(key string, lineNum int) *CheckIssue {
	if key == "" {
		return &CheckIssue{Line: lineNum, Severity: LevelError,
			Message: "malformed line, empty key"}
	}
	return nil
}

func ValidateValue(key string, value string, lineNum int) *CheckIssue {
	if len(value) > 0 && (value[0] == '"' || value[0] == '\'') {
		quote := value[0]
		if strings.IndexByte(value[1:], quote) < 0 {
			return &CheckIssue{Line: lineNum, Severity: LevelError,
				Message: fmt.Sprintf("unclosed quote for key: %q", key)}
		}
	} else {
		checkVal := value
		if hashIdx := strings.Index(value, "#"); hashIdx >= 0 {
			checkVal = strings.TrimSpace(value[:hashIdx])
		}
		if strings.ContainsAny(checkVal, " \t") {
			return &CheckIssue{Line: lineNum, Severity: LevelError,
				Message: fmt.Sprintf("unquoted value contains spaces for key: %q", key)}
		}
	}
	return nil
}

func ConsecutiveBlanks(start, end int, count uint8) *CheckIssue {
	return &CheckIssue{Line: start, Severity: LevelWarn,
		Message: fmt.Sprintf("%d consecutive blank lines (lines %d–%d)", count, start, end),
	}
}

func FileEndsWithBlank(lineNum int) *CheckIssue {
	return &CheckIssue{Line: lineNum, Severity: LevelWarn, Message: "file ends with blank line"}
}

func IsIgnored(cfg config.EnvCheck, code string) bool {
	for _, c := range cfg.IgnoreCodes {
		if c == code {
			return true
		}
	}
	return false
}

func ShouldAdd(issue *CheckIssue, level int, cfg config.EnvCheck, code string) bool {
	return issue != nil && issue.Severity >= level && !IsIgnored(cfg, code)
}
