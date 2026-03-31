package pipeline

import (
	"fmt"
	"sort"
	"strings"
)

// NumberedClassificationMap maps numbered placeholder types to their original substrings.
// For example: "PERSON_1" -> "John Doe", "PERSON_2" -> "Jane Smith"
type NumberedClassificationMap map[string]string

// AssignNumbers converts a ClassificationMap to a NumberedClassificationMap.
// Numbers are assigned based on first occurrence in the text.
func AssignNumbers(cm ClassificationMap, text string) NumberedClassificationMap {
	ncm := make(NumberedClassificationMap)

	// For each pattern type, order substrings by first occurrence in text
	for pt, substrings := range cm {
		// Create a slice of substrings with their first occurrence position
		type substringPos struct {
			substring string
			pos       int
		}
		var ordered []substringPos

		for _, s := range substrings {
			pos := strings.Index(text, s)
			if pos == -1 {
				pos = len(text) // Put at end if not found
			}
			ordered = append(ordered, substringPos{s, pos})
		}

		// Sort by position
		sort.Slice(ordered, func(i, j int) bool {
			return ordered[i].pos < ordered[j].pos
		})

		// Assign numbers
		for i, sp := range ordered {
			key := fmt.Sprintf("%s_%d", pt, i+1)
			ncm[key] = sp.substring
		}
	}

	return ncm
}

// GetPlaceholder returns the numbered placeholder for a given substring.
// Returns empty string if not found.
func (ncm NumberedClassificationMap) GetPlaceholder(substring string) string {
	for placeholder, s := range ncm {
		if s == substring {
			return placeholder
		}
	}
	return ""
}

// GetSubstring returns the original substring for a given placeholder.
// Returns empty string if not found.
func (ncm NumberedClassificationMap) GetSubstring(placeholder string) string {
	return ncm[placeholder]
}

// GetPlaceholdersForType returns all placeholders of a given type.
// For example, GetPlaceholdersForType("PERSON") returns ["PERSON_1", "PERSON_2", ...]
func (ncm NumberedClassificationMap) GetPlaceholdersForType(pt PatternType) []string {
	prefix := string(pt) + "_"
	var placeholders []string
	for placeholder := range ncm {
		if strings.HasPrefix(placeholder, prefix) {
			placeholders = append(placeholders, placeholder)
		}
	}
	sort.Strings(placeholders)
	return placeholders
}
