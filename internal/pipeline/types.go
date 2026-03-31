// Package pipeline implements the redaction pipeline: classification -> labels -> redacted output.
package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PatternType represents a type of sensitive data that can be redacted.
type PatternType string

const (
	CreditCard   PatternType = "CREDIT_CARD"
	EmailAddress PatternType = "EMAIL_ADDRESS"
	IPAddress    PatternType = "IP_ADDRESS"
	MACAddress   PatternType = "MAC_ADDRESS"
	Person       PatternType = "PERSON"
	PhoneNumber  PatternType = "PHONE_NUMBER"
	Obfuscated   PatternType = "OBFUSCATED"
)

// AllPatternTypes returns all supported pattern types.
func AllPatternTypes() []PatternType {
	return []PatternType{
		CreditCard,
		EmailAddress,
		IPAddress,
		MACAddress,
		Person,
		PhoneNumber,
		Obfuscated,
	}
}

// PatternTypeFromString converts a string to a PatternType.
// It handles common variations like "phone", "phone_number", "PHONE_NUMBER".
func PatternTypeFromString(s string) (PatternType, error) {
	normalized := strings.ToUpper(strings.ReplaceAll(s, "-", "_"))

	// Handle short forms
	switch normalized {
	case "CREDIT_CARD", "CREDITCARD", "CC":
		return CreditCard, nil
	case "EMAIL_ADDRESS", "EMAIL":
		return EmailAddress, nil
	case "IP_ADDRESS", "IP":
		return IPAddress, nil
	case "MAC_ADDRESS", "MAC":
		return MACAddress, nil
	case "PERSON", "NAME":
		return Person, nil
	case "PHONE_NUMBER", "PHONE":
		return PhoneNumber, nil
	case "OBFUSCATED":
		return Obfuscated, nil
	default:
		return "", fmt.Errorf("unknown pattern type: %s", s)
	}
}

// ClassificationMap maps pattern types to lists of matched substrings.
// The substrings are unique and case-sensitive.
type ClassificationMap map[PatternType][]string

// NewClassificationMap creates an empty ClassificationMap.
func NewClassificationMap() ClassificationMap {
	return make(ClassificationMap)
}

// Add adds a substring to the classification map for the given type.
// Duplicates are ignored.
func (cm ClassificationMap) Add(pt PatternType, substring string) {
	for _, existing := range cm[pt] {
		if existing == substring {
			return // Already exists
		}
	}
	cm[pt] = append(cm[pt], substring)
}

// Merge combines another ClassificationMap into this one.
func (cm ClassificationMap) Merge(other ClassificationMap) {
	for pt, substrings := range other {
		for _, s := range substrings {
			cm.Add(pt, s)
		}
	}
}

// ToJSON serializes the ClassificationMap to JSON.
func (cm ClassificationMap) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cm, "", "  ")
}

// ClassificationMapFromJSON deserializes a ClassificationMap from JSON.
func ClassificationMapFromJSON(data []byte) (ClassificationMap, error) {
	var cm ClassificationMap
	if err := json.Unmarshal(data, &cm); err != nil {
		return nil, err
	}
	return cm, nil
}

// Label represents a labeled region in the input text.
type Label struct {
	Offset int    `json:"offset"` // Byte offset from start of text
	Length int    `json:"length"` // Length in bytes
	Type   string `json:"type"`   // Numbered placeholder type (e.g., "PERSON_1")
}

// Labels is a slice of Label that can be serialized to JSON.
type Labels []Label

// ToJSON serializes the Labels to JSON.
func (l Labels) ToJSON() ([]byte, error) {
	return json.MarshalIndent(l, "", "  ")
}

// LabelsFromJSON deserializes Labels from JSON.
func LabelsFromJSON(data []byte) (Labels, error) {
	var labels Labels
	if err := json.Unmarshal(data, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}
