package matcher

import (
	"regexp"

	"github.com/kendalharland/redact/internal/pipeline"
)

// IPAddressMatcher detects IP addresses (IPv4 and IPv6) in text.
type IPAddressMatcher struct{}

// NewIPAddressMatcher creates a new IP address matcher.
func NewIPAddressMatcher() *IPAddressMatcher {
	return &IPAddressMatcher{}
}

func (m *IPAddressMatcher) Type() pipeline.PatternType {
	return pipeline.IPAddress
}

func (m *IPAddressMatcher) RequiresLLM() bool {
	return false
}

// IPv4 pattern - matches addresses like 192.168.1.1
var ipv4Pattern = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)

// IPv6 patterns - matches various IPv6 formats
var ipv6Patterns = []*regexp.Regexp{
	// Full IPv6 address
	regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`),
	// Compressed IPv6 with ::
	regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}\b`),
	regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}\b`),
	regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}\b`),
	regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}\b`),
	regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}\b`),
	regexp.MustCompile(`\b[0-9a-fA-F]{1,4}:(?::[0-9a-fA-F]{1,4}){1,6}\b`),
	// Addresses starting with :: (need submatch extraction)
	regexp.MustCompile(`::(?:[0-9a-fA-F]{1,4}:){0,5}[0-9a-fA-F]{1,4}\b`),
	// Loopback ::1
	regexp.MustCompile(`::1\b`),
}

func (m *IPAddressMatcher) Match(text string) ([]string, error) {
	seen := make(map[string]bool)
	var matches []string

	// Match IPv4
	found := ipv4Pattern.FindAllString(text, -1)
	for _, match := range found {
		if !seen[match] {
			seen[match] = true
			matches = append(matches, match)
		}
	}

	// Match IPv6
	for _, pattern := range ipv6Patterns {
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
