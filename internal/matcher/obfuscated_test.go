package matcher

import (
	"testing"

	"github.com/kendalharland/redact/internal/pipeline"
)

func TestObfuscatedMatcher_Type(t *testing.T) {
	m := NewObfuscatedMatcher(nil, nil)
	if m.Type() != pipeline.Obfuscated {
		t.Errorf("expected %v, got %v", pipeline.Obfuscated, m.Type())
	}
}

func TestObfuscatedMatcher_RequiresLLM(t *testing.T) {
	m := NewObfuscatedMatcher(nil, nil)
	if !m.RequiresLLM() {
		t.Error("expected RequiresLLM to be true")
	}
}

func TestParseObfuscatedResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected []string
	}{
		{
			name:     "single obfuscated email",
			response: "john-at-gmail-DOT-com",
			expected: []string{"john-at-gmail-DOT-com"},
		},
		{
			name:     "multiple items",
			response: "john-at-gmail-DOT-com\none-nine-two dot one dot one dot one",
			expected: []string{"john-at-gmail-DOT-com", "one-nine-two dot one dot one dot one"},
		},
		{
			name:     "with NONE",
			response: "NONE",
			expected: nil,
		},
		{
			name:     "with empty lines",
			response: "john-at-gmail-DOT-com\n\ntest[at]test[dot]com\n",
			expected: []string{"john-at-gmail-DOT-com", "test[at]test[dot]com"},
		},
		{
			name:     "duplicates",
			response: "john-at-gmail-DOT-com\njohn-at-gmail-DOT-com",
			expected: []string{"john-at-gmail-DOT-com"},
		},
		{
			name:     "empty response",
			response: "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseObfuscatedResponse(tt.response)
			if !stringSlicesEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestObfuscatedMatcher_Match_NoEnabledTypes(t *testing.T) {
	m := NewObfuscatedMatcher(nil, nil)
	result, err := m.Match("some text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestObfuscatedMatcher_Match_OnlyPersonEnabled(t *testing.T) {
	// Person and Obfuscated are excluded from obfuscation detection
	m := NewObfuscatedMatcher(nil, []pipeline.PatternType{pipeline.Person})
	result, err := m.Match("some text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}
