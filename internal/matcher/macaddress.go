package matcher

import (
	"regexp"

	"github.com/kendalharland/redact/internal/pipeline"
)

// MACAddressMatcher detects MAC addresses in text.
type MACAddressMatcher struct{}

// NewMACAddressMatcher creates a new MAC address matcher.
func NewMACAddressMatcher() *MACAddressMatcher {
	return &MACAddressMatcher{}
}

func (m *MACAddressMatcher) Type() pipeline.PatternType {
	return pipeline.MACAddress
}

func (m *MACAddressMatcher) RequiresLLM() bool {
	return false
}

// MAC address patterns - supports various formats:
// - Colon separated: AA:BB:CC:DD:EE:FF
// - Hyphen separated: AA-BB-CC-DD-EE-FF
// - Dot separated (Cisco style): AABB.CCDD.EEFF
var macAddressPatterns = []*regexp.Regexp{
	// Colon or hyphen separated (most common)
	regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}[:\-]){5}[0-9a-fA-F]{2}\b`),
	// Cisco style (dot separated, groups of 4)
	regexp.MustCompile(`\b[0-9a-fA-F]{4}\.[0-9a-fA-F]{4}\.[0-9a-fA-F]{4}\b`),
}

func (m *MACAddressMatcher) Match(text string) ([]string, error) {
	seen := make(map[string]bool)
	var matches []string

	for _, pattern := range macAddressPatterns {
		found := pattern.FindAllString(text, -1)
		for _, match := range found {
			if !seen[match] {
				seen[match] = true
				matches = append(matches, match)
			}
		}
	}

	return matches, nil
}
