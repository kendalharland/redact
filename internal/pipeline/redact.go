package pipeline

import (
	"sort"
	"strings"
)

const bleepReplacement = "***"

// RedactOptions configures the redaction behavior.
type RedactOptions struct {
	// Bleep replaces all matches with "***" instead of numbered placeholders.
	Bleep bool
}

// Redact applies labels to the text and returns the redacted output.
// Labels must not overlap.
func Redact(text string, labels Labels, opts RedactOptions) string {
	if len(labels) == 0 {
		return text
	}

	// Sort labels by offset in descending order to process from end to start.
	// This preserves offsets as we make replacements.
	sortedLabels := make(Labels, len(labels))
	copy(sortedLabels, labels)
	sort.Slice(sortedLabels, func(i, j int) bool {
		return sortedLabels[i].Offset > sortedLabels[j].Offset
	})

	result := text
	for _, label := range sortedLabels {
		var replacement string
		if opts.Bleep {
			replacement = bleepReplacement
		} else {
			replacement = "<" + label.Type + ">"
		}

		// Replace the substring at the label's position
		result = result[:label.Offset] + replacement + result[label.Offset+label.Length:]
	}

	return result
}

// Pipeline runs the full redaction pipeline.
type Pipeline struct {
	text       string
	classMap   ClassificationMap
	numberedCM NumberedClassificationMap
	labels     Labels
}

// NewPipeline creates a new redaction pipeline.
func NewPipeline(text string) *Pipeline {
	return &Pipeline{
		text:     text,
		classMap: NewClassificationMap(),
	}
}

// AddMatches adds matched substrings to the classification map.
func (p *Pipeline) AddMatches(pt PatternType, matches []string) {
	for _, match := range matches {
		p.classMap.Add(pt, match)
	}
}

// Process processes the classification map and generates labels.
func (p *Pipeline) Process() {
	p.numberedCM = AssignNumbers(p.classMap, p.text)
	p.labels = GenerateLabels(p.text, p.numberedCM)
	p.labels = RemoveOverlappingLabels(p.labels)
}

// GetClassificationMap returns the classification map.
func (p *Pipeline) GetClassificationMap() ClassificationMap {
	return p.classMap
}

// GetNumberedClassificationMap returns the numbered classification map.
func (p *Pipeline) GetNumberedClassificationMap() NumberedClassificationMap {
	return p.numberedCM
}

// GetLabels returns the generated labels.
func (p *Pipeline) GetLabels() Labels {
	return p.labels
}

// GetRedactedText returns the redacted text.
func (p *Pipeline) GetRedactedText(opts RedactOptions) string {
	return Redact(p.text, p.labels, opts)
}

// ExtractBaseType extracts the base pattern type from a numbered placeholder.
// For example: "PERSON_1" -> "PERSON", "EMAIL_ADDRESS_2" -> "EMAIL_ADDRESS"
func ExtractBaseType(placeholder string) string {
	// Find the last underscore followed by digits
	lastUnderscore := strings.LastIndex(placeholder, "_")
	if lastUnderscore == -1 {
		return placeholder
	}

	// Check if everything after the underscore is digits
	suffix := placeholder[lastUnderscore+1:]
	for _, r := range suffix {
		if r < '0' || r > '9' {
			return placeholder
		}
	}

	return placeholder[:lastUnderscore]
}
