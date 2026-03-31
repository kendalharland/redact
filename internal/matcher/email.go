package matcher

import (
	"regexp"

	"github.com/kendalharland/redact/internal/pipeline"
)

// EmailMatcher detects email addresses in text.
type EmailMatcher struct{}

// NewEmailMatcher creates a new email matcher.
func NewEmailMatcher() *EmailMatcher {
	return &EmailMatcher{}
}

func (m *EmailMatcher) Type() pipeline.PatternType {
	return pipeline.EmailAddress
}

func (m *EmailMatcher) RequiresLLM() bool {
	return false
}

// emailPattern matches standard email addresses.
// This is a simplified pattern that covers most common email formats.
var emailPattern = regexp.MustCompile(`\b[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}\b`)

func (m *EmailMatcher) Match(text string) ([]string, error) {
	seen := make(map[string]bool)
	var matches []string

	found := emailPattern.FindAllString(text, -1)
	for _, match := range found {
		if !seen[match] {
			seen[match] = true
			matches = append(matches, match)
		}
	}

	return matches, nil
}
