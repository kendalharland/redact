package matcher

import (
	"testing"

	"github.com/kendalharland/redact/internal/pipeline"
)

func TestCreditCardMatcher_Type(t *testing.T) {
	m := NewCreditCardMatcher()
	if m.Type() != pipeline.CreditCard {
		t.Errorf("expected %v, got %v", pipeline.CreditCard, m.Type())
	}
}

func TestCreditCardMatcher_RequiresLLM(t *testing.T) {
	m := NewCreditCardMatcher()
	if m.RequiresLLM() {
		t.Error("expected RequiresLLM to be false")
	}
}

func TestCreditCardMatcher_Match(t *testing.T) {
	m := NewCreditCardMatcher()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "visa with dashes",
			input:    "My card is 4532-0151-1283-0366",
			expected: []string{"4532-0151-1283-0366"},
		},
		{
			name:     "visa with spaces",
			input:    "Card: 4532 0151 1283 0366",
			expected: []string{"4532 0151 1283 0366"},
		},
		{
			name:     "mastercard no separators",
			input:    "Number: 5425233430109903",
			expected: []string{"5425233430109903"},
		},
		{
			name:     "amex with dashes",
			input:    "Amex: 3714-496353-98431",
			expected: []string{"3714-496353-98431"},
		},
		{
			name:     "multiple cards",
			input:    "Cards: 4532-0151-1283-0366 and 5425233430109903",
			expected: []string{"4532-0151-1283-0366", "5425233430109903"},
		},
		{
			name:     "invalid luhn",
			input:    "Invalid: 1234-5678-9012-3456",
			expected: nil,
		},
		{
			name:     "no credit cards",
			input:    "No credit cards here",
			expected: nil,
		},
		{
			name:     "duplicate cards",
			input:    "Card 4532-0151-1283-0366 and 4532-0151-1283-0366 again",
			expected: []string{"4532-0151-1283-0366"},
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

func TestIsValidLuhn(t *testing.T) {
	tests := []struct {
		number string
		valid  bool
	}{
		{"4532015112830366", true},  // Valid Visa
		{"5425233430109903", true},  // Valid MasterCard
		{"371449635398431", true},   // Valid Amex
		{"1234567890123456", false}, // Invalid
		{"123", false},              // Too short
		{"12345678901234567890", false}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.number, func(t *testing.T) {
			if isValidLuhn(tt.number) != tt.valid {
				t.Errorf("isValidLuhn(%s) = %v, expected %v", tt.number, !tt.valid, tt.valid)
			}
		})
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
