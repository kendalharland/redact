package matcher

import (
	"testing"

	"github.com/kendalharland/redact/internal/pipeline"
)

func TestEmailMatcher_Type(t *testing.T) {
	m := NewEmailMatcher()
	if m.Type() != pipeline.EmailAddress {
		t.Errorf("expected %v, got %v", pipeline.EmailAddress, m.Type())
	}
}

func TestEmailMatcher_RequiresLLM(t *testing.T) {
	m := NewEmailMatcher()
	if m.RequiresLLM() {
		t.Error("expected RequiresLLM to be false")
	}
}

func TestEmailMatcher_Match(t *testing.T) {
	m := NewEmailMatcher()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple email",
			input:    "Contact me at john@example.com",
			expected: []string{"john@example.com"},
		},
		{
			name:     "email with subdomain",
			input:    "Email: user@mail.example.org",
			expected: []string{"user@mail.example.org"},
		},
		{
			name:     "email with plus",
			input:    "Send to test+filter@gmail.com",
			expected: []string{"test+filter@gmail.com"},
		},
		{
			name:     "email with dots",
			input:    "first.last@company.co.uk",
			expected: []string{"first.last@company.co.uk"},
		},
		{
			name:     "multiple emails",
			input:    "From alice@test.com to bob@test.com",
			expected: []string{"alice@test.com", "bob@test.com"},
		},
		{
			name:     "no emails",
			input:    "No emails here at all",
			expected: nil,
		},
		{
			name:     "duplicate emails",
			input:    "Send to user@test.com and also user@test.com",
			expected: []string{"user@test.com"},
		},
		{
			name:     "email with hyphen",
			input:    "Contact support-team@company-name.com",
			expected: []string{"support-team@company-name.com"},
		},
		{
			name:     "email with numbers",
			input:    "Email: user123@mail2.example456.net",
			expected: []string{"user123@mail2.example456.net"},
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
