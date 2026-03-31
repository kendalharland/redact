package pipeline

import (
	"testing"
)

func TestGenerateLabels(t *testing.T) {
	text := "Hello John, meet Jane. John says hi."
	ncm := NumberedClassificationMap{
		"PERSON_1": "John",
		"PERSON_2": "Jane",
	}

	labels := GenerateLabels(text, ncm)

	// Should find: John at 6, Jane at 17, John at 23
	// "Hello John, meet Jane. John says hi."
	//  0     6          17    23
	if len(labels) != 3 {
		t.Fatalf("expected 3 labels, got %d: %+v", len(labels), labels)
	}

	// Labels should be sorted by offset
	if labels[0].Offset != 6 {
		t.Errorf("expected first label at offset 6, got %d", labels[0].Offset)
	}
	if labels[0].Type != "PERSON_1" {
		t.Errorf("expected first label to be PERSON_1, got %s", labels[0].Type)
	}

	if labels[1].Offset != 17 {
		t.Errorf("expected second label at offset 17, got %d", labels[1].Offset)
	}
	if labels[1].Type != "PERSON_2" {
		t.Errorf("expected second label to be PERSON_2, got %s", labels[1].Type)
	}

	if labels[2].Offset != 23 {
		t.Errorf("expected third label at offset 23, got %d", labels[2].Offset)
	}
	if labels[2].Type != "PERSON_1" {
		t.Errorf("expected third label to be PERSON_1, got %s", labels[2].Type)
	}
}

func TestGenerateLabels_WordBoundary(t *testing.T) {
	text := "MacBook is not Mac. Mac is a name."
	ncm := NumberedClassificationMap{
		"PERSON_1": "Mac",
	}

	labels := GenerateLabels(text, ncm)

	// Should NOT match Mac in MacBook, but should match standalone Mac twice
	// "MacBook is not Mac. Mac is a name."
	//  0       7      15  19 20
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d: %+v", len(labels), labels)
	}

	// First Mac at position 15 ("not Mac.")
	if labels[0].Offset != 15 {
		t.Errorf("expected first label at offset 15, got %d", labels[0].Offset)
	}

	// Second Mac at position 20 ("Mac is")
	if labels[1].Offset != 20 {
		t.Errorf("expected second label at offset 20, got %d", labels[1].Offset)
	}
}

func TestGenerateLabels_PartialWordMatch(t *testing.T) {
	text := "Mac'n'cheese is delicious. Mac is here."
	ncm := NumberedClassificationMap{
		"PERSON_1": "Mac",
	}

	labels := GenerateLabels(text, ncm)

	// Should NOT match Mac in Mac'n'cheese (n is a word char after apostrophe)
	// Should match standalone Mac
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d: %+v", len(labels), labels)
	}

	if labels[0].Offset != 27 {
		t.Errorf("expected label at offset 27, got %d", labels[0].Offset)
	}
}

func TestRemoveOverlappingLabels(t *testing.T) {
	labels := Labels{
		{Offset: 0, Length: 10, Type: "PERSON_1"},
		{Offset: 5, Length: 10, Type: "PERSON_2"}, // Overlaps with first
		{Offset: 20, Length: 5, Type: "PERSON_3"}, // No overlap
	}

	result := RemoveOverlappingLabels(labels)

	if len(result) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(result))
	}
	if result[0].Type != "PERSON_1" {
		t.Errorf("expected first to be PERSON_1")
	}
	if result[1].Type != "PERSON_3" {
		t.Errorf("expected second to be PERSON_3")
	}
}

func TestLabels_ToJSON(t *testing.T) {
	labels := Labels{
		{Offset: 0, Length: 10, Type: "PERSON_1"},
		{Offset: 20, Length: 5, Type: "EMAIL_ADDRESS_1"},
	}

	data, err := labels.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse it back
	parsed, err := LabelsFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(parsed))
	}
	if parsed[0].Offset != 0 || parsed[0].Length != 10 || parsed[0].Type != "PERSON_1" {
		t.Errorf("first label mismatch: %+v", parsed[0])
	}
}

func TestFindAllOccurrences(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		substring string
		expected  []int
	}{
		{
			name:      "multiple occurrences",
			text:      "abc def abc ghi abc",
			substring: "abc",
			expected:  []int{0, 8, 16},
		},
		{
			name:      "no occurrences",
			text:      "hello world",
			substring: "xyz",
			expected:  nil,
		},
		{
			name:      "word boundary prevents match",
			text:      "MacBook contains Mac",
			substring: "Mac",
			expected:  []int{17},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findAllOccurrences(tt.text, tt.substring)
			if !intSlicesEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func intSlicesEqual(a, b []int) bool {
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
