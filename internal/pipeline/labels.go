package pipeline

import (
	"sort"
	"strings"
	"unicode"
)

// GenerateLabels finds all occurrences of classified substrings in the text
// and returns a sorted list of labels.
func GenerateLabels(text string, ncm NumberedClassificationMap) Labels {
	var labels Labels

	// For each placeholder -> substring mapping
	for placeholder, substring := range ncm {
		// Find all occurrences of this substring
		positions := findAllOccurrences(text, substring)
		for _, pos := range positions {
			labels = append(labels, Label{
				Offset: pos,
				Length: len(substring),
				Type:   placeholder,
			})
		}
	}

	// Sort by offset
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Offset < labels[j].Offset
	})

	return labels
}

// findAllOccurrences finds all word-bounded occurrences of substring in text.
// It avoids partial word matches (e.g., "Mac" in "MacBook" or "Mac'n'cheese").
func findAllOccurrences(text, substring string) []int {
	var positions []int
	start := 0

	for {
		idx := strings.Index(text[start:], substring)
		if idx == -1 {
			break
		}

		absIdx := start + idx
		endIdx := absIdx + len(substring)

		// Check word boundaries
		if isWordBoundary(text, absIdx, endIdx, substring) {
			positions = append(positions, absIdx)
		}

		start = absIdx + 1
	}

	return positions
}

// isWordBoundary checks if the substring at the given position is at word boundaries.
// This prevents matching "Mac" in "MacBook" or "Mac'n'cheese".
func isWordBoundary(text string, startIdx, endIdx int, substring string) bool {
	// Check character before the match
	if startIdx > 0 {
		prevRune := rune(text[startIdx-1])
		firstRune := rune(substring[0])

		// If previous char is a word character and first char of substring is also
		// a word character, this is not a word boundary
		if isWordChar(prevRune) && isWordChar(firstRune) {
			return false
		}
	}

	// Check character after the match
	if endIdx < len(text) {
		nextRune := rune(text[endIdx])
		lastRune := rune(substring[len(substring)-1])

		// If next char is a word character and last char of substring is also
		// a word character, this is not a word boundary
		if isWordChar(nextRune) && isWordChar(lastRune) {
			return false
		}

		// Handle contractions and compound words with apostrophes.
		// If next char is an apostrophe followed by a word char, not a boundary.
		// Examples: Mac'n'cheese, don't, it's
		if nextRune == '\'' && endIdx+1 < len(text) {
			afterApostrophe := rune(text[endIdx+1])
			if isWordChar(afterApostrophe) && isWordChar(lastRune) {
				return false
			}
		}
	}

	return true
}

// isWordChar returns true if the rune is a word character (letter, digit, or underscore).
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// RemoveOverlappingLabels removes labels that overlap with other labels.
// When labels overlap, the one that appears first in the list is kept.
// The input labels should be sorted by offset.
func RemoveOverlappingLabels(labels Labels) Labels {
	if len(labels) == 0 {
		return labels
	}

	var result Labels
	result = append(result, labels[0])

	for i := 1; i < len(labels); i++ {
		lastLabel := result[len(result)-1]
		currentLabel := labels[i]

		// Check if current label overlaps with the last kept label
		lastEnd := lastLabel.Offset + lastLabel.Length
		if currentLabel.Offset >= lastEnd {
			result = append(result, currentLabel)
		}
	}

	return result
}
