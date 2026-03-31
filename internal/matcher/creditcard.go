package matcher

import (
	"regexp"
	"strings"

	"github.com/kendalharland/redact/internal/pipeline"
)

// CreditCardMatcher detects credit card numbers in text.
type CreditCardMatcher struct{}

// NewCreditCardMatcher creates a new credit card matcher.
func NewCreditCardMatcher() *CreditCardMatcher {
	return &CreditCardMatcher{}
}

func (m *CreditCardMatcher) Type() pipeline.PatternType {
	return pipeline.CreditCard
}

func (m *CreditCardMatcher) RequiresLLM() bool {
	return false
}

// creditCardPatterns matches various credit card number formats.
// Supports formats with spaces, dashes, or no separators.
var creditCardPatterns = []*regexp.Regexp{
	// 16 digits with optional separators (Visa, MasterCard, Discover)
	regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
	// 15 digits with optional separators (Amex)
	regexp.MustCompile(`\b\d{4}[-\s]?\d{6}[-\s]?\d{5}\b`),
}

func (m *CreditCardMatcher) Match(text string) ([]string, error) {
	seen := make(map[string]bool)
	var matches []string

	for _, pattern := range creditCardPatterns {
		found := pattern.FindAllString(text, -1)
		for _, match := range found {
			// Normalize for deduplication (remove separators)
			normalized := strings.ReplaceAll(strings.ReplaceAll(match, "-", ""), " ", "")

			// Validate using Luhn algorithm
			if !isValidLuhn(normalized) {
				continue
			}

			// Keep original form but deduplicate by normalized form
			if !seen[normalized] {
				seen[normalized] = true
				matches = append(matches, match)
			}
		}
	}

	return matches, nil
}

// isValidLuhn checks if a number passes the Luhn algorithm.
func isValidLuhn(number string) bool {
	if len(number) < 13 || len(number) > 19 {
		return false
	}

	var sum int
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
