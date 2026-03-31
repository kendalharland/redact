package matcher

import (
	"testing"

	"github.com/kendalharland/redact/internal/pipeline"
)

func TestIPAddressMatcher_Type(t *testing.T) {
	m := NewIPAddressMatcher()
	if m.Type() != pipeline.IPAddress {
		t.Errorf("expected %v, got %v", pipeline.IPAddress, m.Type())
	}
}

func TestIPAddressMatcher_RequiresLLM(t *testing.T) {
	m := NewIPAddressMatcher()
	if m.RequiresLLM() {
		t.Error("expected RequiresLLM to be false")
	}
}

func TestIPAddressMatcher_Match(t *testing.T) {
	m := NewIPAddressMatcher()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple ipv4",
			input:    "Server at 192.168.1.1",
			expected: []string{"192.168.1.1"},
		},
		{
			name:     "ipv4 with zeros",
			input:    "Address: 10.0.0.1",
			expected: []string{"10.0.0.1"},
		},
		{
			name:     "ipv4 max values",
			input:    "IP: 255.255.255.255",
			expected: []string{"255.255.255.255"},
		},
		{
			name:     "multiple ipv4",
			input:    "From 192.168.1.1 to 10.0.0.1",
			expected: []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:     "ipv6 full",
			input:    "IPv6: 2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected: []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		},
		{
			name:     "ipv6 loopback",
			input:    "Loopback: ::1",
			expected: []string{"::1"},
		},
		{
			name:     "no ip addresses",
			input:    "No IP addresses here",
			expected: nil,
		},
		{
			name:     "invalid ipv4 - out of range",
			input:    "Invalid: 300.168.1.1",
			expected: nil,
		},
		{
			name:     "duplicate ips",
			input:    "Server 192.168.1.1 and 192.168.1.1 again",
			expected: []string{"192.168.1.1"},
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
