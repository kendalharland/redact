package matcher

import (
	"testing"

	"github.com/kendalharland/redact/internal/pipeline"
)

func TestPersonMatcher_Type(t *testing.T) {
	m := NewPersonMatcher(nil)
	if m.Type() != pipeline.Person {
		t.Errorf("expected %v, got %v", pipeline.Person, m.Type())
	}
}

func TestPersonMatcher_RequiresLLM(t *testing.T) {
	m := NewPersonMatcher(nil)
	if !m.RequiresLLM() {
		t.Error("expected RequiresLLM to be true")
	}
}

func TestParsePersonResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected []string
	}{
		{
			name:     "single name",
			response: "John Doe",
			expected: []string{"John Doe"},
		},
		{
			name:     "multiple names",
			response: "John Doe\nJane Smith\nBob",
			expected: []string{"John Doe", "Jane Smith", "Bob"},
		},
		{
			name:     "with NONE",
			response: "NONE",
			expected: nil,
		},
		{
			name:     "with empty lines",
			response: "John Doe\n\nJane Smith\n",
			expected: []string{"John Doe", "Jane Smith"},
		},
		{
			name:     "duplicates",
			response: "John\nJohn\nJane",
			expected: []string{"John", "Jane"},
		},
		{
			name:     "with whitespace",
			response: "  John Doe  \n  Jane Smith  ",
			expected: []string{"John Doe", "Jane Smith"},
		},
		{
			name:     "empty response",
			response: "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePersonResponse(tt.response)
			if !stringSlicesEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
