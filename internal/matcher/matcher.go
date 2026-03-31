// Package matcher provides interfaces and implementations for detecting sensitive data in text.
package matcher

import (
	"github.com/kendalharland/redact/internal/pipeline"
)

// Matcher detects sensitive data of a specific type in text.
type Matcher interface {
	// Type returns the pattern type this matcher detects.
	Type() pipeline.PatternType

	// Match finds all unique occurrences of sensitive data in the text.
	// Returns a slice of unique matched substrings.
	Match(text string) ([]string, error)

	// RequiresLLM returns true if this matcher requires an LLM API call.
	RequiresLLM() bool
}

// Registry holds all available matchers.
type Registry struct {
	matchers []Matcher
}

// NewRegistry creates a new matcher registry.
func NewRegistry() *Registry {
	return &Registry{
		matchers: make([]Matcher, 0),
	}
}

// Register adds a matcher to the registry.
func (r *Registry) Register(m Matcher) {
	r.matchers = append(r.matchers, m)
}

// GetAll returns all registered matchers.
func (r *Registry) GetAll() []Matcher {
	return r.matchers
}

// GetByType returns the matcher for a specific pattern type, or nil if not found.
func (r *Registry) GetByType(pt pipeline.PatternType) Matcher {
	for _, m := range r.matchers {
		if m.Type() == pt {
			return m
		}
	}
	return nil
}

// GetEnabled returns matchers for enabled pattern types.
// If excludeTypes is nil or empty, all matchers are returned.
func (r *Registry) GetEnabled(excludeTypes []pipeline.PatternType) []Matcher {
	if len(excludeTypes) == 0 {
		return r.matchers
	}

	excluded := make(map[pipeline.PatternType]bool)
	for _, pt := range excludeTypes {
		excluded[pt] = true
	}

	var enabled []Matcher
	for _, m := range r.matchers {
		if !excluded[m.Type()] {
			enabled = append(enabled, m)
		}
	}
	return enabled
}

// GetNonLLMMatchers returns all matchers that don't require an LLM.
func (r *Registry) GetNonLLMMatchers() []Matcher {
	var result []Matcher
	for _, m := range r.matchers {
		if !m.RequiresLLM() {
			result = append(result, m)
		}
	}
	return result
}

// GetLLMMatchers returns all matchers that require an LLM.
func (r *Registry) GetLLMMatchers() []Matcher {
	var result []Matcher
	for _, m := range r.matchers {
		if m.RequiresLLM() {
			result = append(result, m)
		}
	}
	return result
}
