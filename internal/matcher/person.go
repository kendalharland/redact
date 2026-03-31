package matcher

import (
	"strings"

	"github.com/kendalharland/redact/internal/config"
	"github.com/kendalharland/redact/internal/pipeline"
)

const personPrompt = `You are a PII detection system. Your task is to identify all person names in the given text.

Rules:
1. Extract only person names (first names, last names, full names)
2. Include nicknames and informal names
3. Do NOT include company names, product names, or place names
4. Do NOT include titles (Mr., Dr., etc.) in the extracted name
5. Return one name per line
6. If no person names are found, return "NONE"
7. Return names exactly as they appear in the text (preserve case)
8. Do NOT add any explanation or commentary

Examples:

Input: "Today Kendal went for a jog with Dr. Smith."
Output:
Kendal
Smith

Input: "Apple Inc was founded by Steve Jobs."
Output:
Steve Jobs

Input: "The meeting is at 3pm in the conference room."
Output:
NONE

Input: "Call me Bob, said Robert Johnson."
Output:
Bob
Robert Johnson

Now extract all person names from the following text:
`

// PersonMatcher detects person names using an LLM.
type PersonMatcher struct {
	client *config.Client
}

// NewPersonMatcher creates a new person name matcher.
func NewPersonMatcher(client *config.Client) *PersonMatcher {
	return &PersonMatcher{client: client}
}

func (m *PersonMatcher) Type() pipeline.PatternType {
	return pipeline.Person
}

func (m *PersonMatcher) RequiresLLM() bool {
	return true
}

func (m *PersonMatcher) Match(text string) ([]string, error) {
	prompt := personPrompt + text

	response, err := m.client.Complete(prompt)
	if err != nil {
		return nil, err
	}

	return parsePersonResponse(response), nil
}

// parsePersonResponse parses the LLM response to extract person names.
func parsePersonResponse(response string) []string {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	seen := make(map[string]bool)
	var names []string

	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" || name == "NONE" {
			continue
		}
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}

	return names
}
