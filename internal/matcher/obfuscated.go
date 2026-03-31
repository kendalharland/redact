package matcher

import (
	"fmt"
	"strings"

	"github.com/kendalharland/redact/internal/config"
	"github.com/kendalharland/redact/internal/pipeline"
)

const obfuscatedPromptTemplate = `You are a PII detection system. Your task is to identify obfuscated versions of sensitive data in the given text.

Obfuscated data is data that has been intentionally modified to hide its true form while remaining human-readable.

You are looking for obfuscated versions of these data types: %s

Common obfuscation patterns include:
- Replacing @ with "at", "-at-", "[at]", "(at)", etc.
- Replacing . with "dot", "-dot-", "[dot]", "(dot)", etc.
- Replacing numbers with words (one, two, etc.)
- Adding spaces or hyphens between characters
- Using homoglyphs (0 for O, 1 for l, etc.)
- Spelling out special characters

Examples:

Input: "Contact me at john-at-gmail-DOT-com"
Output:
john-at-gmail-DOT-com

Input: "My IP is one-nine-two dot one-six-eight dot one dot one"
Output:
one-nine-two dot one-six-eight dot one dot one

Input: "Call me at 555 FIVE FIVE FIVE 1234"
Output:
555 FIVE FIVE FIVE 1234

Input: "The server IP is 192.168.1.1"
Output:
NONE

Input: "Contact support@company.com for help"
Output:
NONE

Rules:
1. Only detect OBFUSCATED data, not regular data that matches standard patterns
2. Return one obfuscated item per line
3. If no obfuscated data is found, return "NONE"
4. Return the obfuscated text exactly as it appears (preserve case and spacing)
5. Do NOT add any explanation or commentary

Now extract all obfuscated data from the following text:
`

// ObfuscatedMatcher detects obfuscated versions of sensitive data using an LLM.
type ObfuscatedMatcher struct {
	client       *config.Client
	enabledTypes []pipeline.PatternType
}

// NewObfuscatedMatcher creates a new obfuscated data matcher.
// enabledTypes specifies which pattern types to look for obfuscated versions of.
func NewObfuscatedMatcher(client *config.Client, enabledTypes []pipeline.PatternType) *ObfuscatedMatcher {
	return &ObfuscatedMatcher{
		client:       client,
		enabledTypes: enabledTypes,
	}
}

func (m *ObfuscatedMatcher) Type() pipeline.PatternType {
	return pipeline.Obfuscated
}

func (m *ObfuscatedMatcher) RequiresLLM() bool {
	return true
}

func (m *ObfuscatedMatcher) Match(text string) ([]string, error) {
	if len(m.enabledTypes) == 0 {
		return nil, nil
	}

	// Build the list of enabled types for the prompt
	var typeNames []string
	for _, pt := range m.enabledTypes {
		// Only include pattern-matchable types (not Person or Obfuscated itself)
		if pt != pipeline.Person && pt != pipeline.Obfuscated {
			typeNames = append(typeNames, string(pt))
		}
	}

	if len(typeNames) == 0 {
		return nil, nil
	}

	prompt := fmt.Sprintf(obfuscatedPromptTemplate, strings.Join(typeNames, ", ")) + text

	response, err := m.client.Complete(prompt)
	if err != nil {
		return nil, err
	}

	return parseObfuscatedResponse(response), nil
}

// parseObfuscatedResponse parses the LLM response to extract obfuscated data.
func parseObfuscatedResponse(response string) []string {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	seen := make(map[string]bool)
	var items []string

	for _, line := range lines {
		item := strings.TrimSpace(line)
		if item == "" || item == "NONE" {
			continue
		}
		if !seen[item] {
			seen[item] = true
			items = append(items, item)
		}
	}

	return items
}
