package matcher

import (
	"testing"

	"github.com/kendalharland/redact/internal/pipeline"
)

func TestPhoneMatcher_Type(t *testing.T) {
	m := NewPhoneMatcher()
	if m.Type() != pipeline.PhoneNumber {
		t.Errorf("expected %v, got %v", pipeline.PhoneNumber, m.Type())
	}
}

func TestPhoneMatcher_RequiresLLM(t *testing.T) {
	m := NewPhoneMatcher()
	if m.RequiresLLM() {
		t.Error("expected RequiresLLM to be false")
	}
}

func TestPhoneMatcher_Match(t *testing.T) {
	m := NewPhoneMatcher()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "with parentheses",
			input:    "Call me at (123) 456-7890",
			expected: []string{"(123) 456-7890"},
		},
		{
			name:     "dashes only",
			input:    "Phone: 123-456-7890",
			expected: []string{"123-456-7890"},
		},
		{
			name:     "dots",
			input:    "Number: 123.456.7890",
			expected: []string{"123.456.7890"},
		},
		{
			name:     "spaces",
			input:    "Phone: 123 456 7890",
			expected: []string{"123 456 7890"},
		},
		{
			name:     "with country code",
			input:    "International: +1-123-456-7890",
			expected: []string{"+1-123-456-7890"},
		},
		{
			name:     "no separators",
			input:    "Number: 1234567890",
			expected: []string{"1234567890"},
		},
		{
			name:     "multiple phones",
			input:    "Home: 123-456-7890 Work: 098-765-4321",
			expected: []string{"123-456-7890", "098-765-4321"},
		},
		{
			name:     "no phone numbers",
			input:    "No phone numbers here",
			expected: nil,
		},
		{
			name:     "duplicate phones",
			input:    "Call 123-456-7890 or 123-456-7890",
			expected: []string{"123-456-7890"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := m.Match(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !stringSlicesEqual(matches, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, matches)
			}
		})
	}
}
