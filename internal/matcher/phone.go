package matcher

import (
	"regexp"
	"sort"

	"github.com/kendalharland/redact/internal/pipeline"
)

// PhoneMatcher detects US phone numbers in text.
type PhoneMatcher struct{}

// NewPhoneMatcher creates a new phone number matcher.
func NewPhoneMatcher() *PhoneMatcher {
	return &PhoneMatcher{}
}

func (m *PhoneMatcher) Type() pipeline.PatternType {
	return pipeline.PhoneNumber
}

func (m *PhoneMatcher) RequiresLLM() bool {
	return false
}

// Phone number patterns - supports various US formats:
// - (123) 456-7890
// - 123-456-7890
// - 123.456.7890
// - 1234567890
// - +1 123 456 7890
// - +1-123-456-7890
// Patterns are ordered from longest to shortest to prefer longer matches.
var phonePatterns = []*regexp.Regexp{
	// With country code: +1 123 456 7890, +1-123-456-7890 (longest first)
	regexp.MustCompile(`\+1[-.\s]?\d{3}[-.\s]?\d{3}[-.\s]?\d{4}`),
	// With parentheses: (123) 456-7890
	regexp.MustCompile(`\(\d{3}\)\s*\d{3}[-.\s]?\d{4}`),
	// Standard formats: 123-456-7890, 123.456.7890, 123 456 7890
	regexp.MustCompile(`\b\d{3}[-.\s]\d{3}[-.\s]\d{4}\b`),
	// 10 consecutive digits (be careful with this one)
	regexp.MustCompile(`\b\d{10}\b`),
}

type phoneMatch struct {
	start int
	end   int
	text  string
}

func (m *PhoneMatcher) Match(text string) ([]string, error) {
	var allMatches []phoneMatch

	// Collect all matches with their positions
	for _, pattern := range phonePatterns {
		indexes := pattern.FindAllStringIndex(text, -1)
		for _, idx := range indexes {
			allMatches = append(allMatches, phoneMatch{
				start: idx[0],
				end:   idx[1],
				text:  text[idx[0]:idx[1]],
			})
		}
	}

	// Sort by start position, then by length (longer matches first)
	sort.Slice(allMatches, func(i, j int) bool {
		if allMatches[i].start != allMatches[j].start {
			return allMatches[i].start < allMatches[j].start
		}
		return (allMatches[i].end - allMatches[i].start) > (allMatches[j].end - allMatches[j].start)
	})

	// Remove overlapping matches (keep longer ones)
	seen := make(map[string]bool)
	var matches []string
	lastEnd := -1

	for _, match := range allMatches {
		// Skip if this match overlaps with a previous one
		if match.start < lastEnd {
			continue
		}
		// Skip duplicates
		if seen[match.text] {
			continue
		}
		seen[match.text] = true
		matches = append(matches, match.text)
		lastEnd = match.end
	}

	return matches, nil
}
