package matcher

import (
	"testing"

	"github.com/kendalharland/redact/internal/pipeline"
)

func TestMACAddressMatcher_Type(t *testing.T) {
	m := NewMACAddressMatcher()
	if m.Type() != pipeline.MACAddress {
		t.Errorf("expected %v, got %v", pipeline.MACAddress, m.Type())
	}
}

func TestMACAddressMatcher_RequiresLLM(t *testing.T) {
	m := NewMACAddressMatcher()
	if m.RequiresLLM() {
		t.Error("expected RequiresLLM to be false")
	}
}

func TestMACAddressMatcher_Match(t *testing.T) {
	m := NewMACAddressMatcher()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "colon separated",
			input:    "MAC: 00:1A:2B:3C:4D:5E",
			expected: []string{"00:1A:2B:3C:4D:5E"},
		},
		{
			name:     "hyphen separated",
			input:    "MAC: 00-1A-2B-3C-4D-5E",
			expected: []string{"00-1A-2B-3C-4D-5E"},
		},
		{
			name:     "cisco format",
			input:    "MAC: 001A.2B3C.4D5E",
			expected: []string{"001A.2B3C.4D5E"},
		},
		{
			name:     "lowercase",
			input:    "mac: aa:bb:cc:dd:ee:ff",
			expected: []string{"aa:bb:cc:dd:ee:ff"},
		},
		{
			name:     "multiple macs",
			input:    "From 00:11:22:33:44:55 to AA:BB:CC:DD:EE:FF",
			expected: []string{"00:11:22:33:44:55", "AA:BB:CC:DD:EE:FF"},
		},
		{
			name:     "no mac addresses",
			input:    "No MAC addresses here",
			expected: nil,
		},
		{
			name:     "duplicate macs",
			input:    "MAC 00:11:22:33:44:55 and 00:11:22:33:44:55 again",
			expected: []string{"00:11:22:33:44:55"},
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
