package pipeline

import (
	"testing"
)

func TestRedact(t *testing.T) {
	text := "Hello John, meet Jane."
	labels := Labels{
		{Offset: 6, Length: 4, Type: "PERSON_1"},
		{Offset: 17, Length: 4, Type: "PERSON_2"},
	}

	result := Redact(text, labels, RedactOptions{})

	expected := "Hello <PERSON_1>, meet <PERSON_2>."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRedact_Bleep(t *testing.T) {
	text := "Hello John, meet Jane."
	labels := Labels{
		{Offset: 6, Length: 4, Type: "PERSON_1"},
		{Offset: 17, Length: 4, Type: "PERSON_2"},
	}

	result := Redact(text, labels, RedactOptions{Bleep: true})

	expected := "Hello ***, meet ***."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRedact_Empty(t *testing.T) {
	text := "Hello world."
	labels := Labels{}

	result := Redact(text, labels, RedactOptions{})

	if result != text {
		t.Errorf("expected %q, got %q", text, result)
	}
}

func TestRedact_MultipleTypes(t *testing.T) {
	text := "John's email is john@test.com and phone is 123-456-7890."
	labels := Labels{
		{Offset: 0, Length: 4, Type: "PERSON_1"},
		{Offset: 16, Length: 13, Type: "EMAIL_ADDRESS_1"},
		{Offset: 43, Length: 12, Type: "PHONE_NUMBER_1"},
	}

	result := Redact(text, labels, RedactOptions{})

	expected := "<PERSON_1>'s email is <EMAIL_ADDRESS_1> and phone is <PHONE_NUMBER_1>."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestPipeline_EndToEnd(t *testing.T) {
	text := "Today Kendal went for a jog."

	p := NewPipeline(text)
	p.AddMatches(Person, []string{"Kendal"})
	p.Process()

	// Check classification map
	cm := p.GetClassificationMap()
	if len(cm[Person]) != 1 || cm[Person][0] != "Kendal" {
		t.Errorf("expected [Kendal], got %v", cm[Person])
	}

	// Check numbered classification map
	ncm := p.GetNumberedClassificationMap()
	if ncm["PERSON_1"] != "Kendal" {
		t.Errorf("expected PERSON_1 = Kendal, got %s", ncm["PERSON_1"])
	}

	// Check labels
	labels := p.GetLabels()
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0].Offset != 6 || labels[0].Length != 6 || labels[0].Type != "PERSON_1" {
		t.Errorf("unexpected label: %+v", labels[0])
	}

	// Check redacted output
	result := p.GetRedactedText(RedactOptions{})
	expected := "Today <PERSON_1> went for a jog."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestPipeline_MultiplePersons(t *testing.T) {
	text := "Octavia Butler and Isaac Asimov are authors."

	p := NewPipeline(text)
	p.AddMatches(Person, []string{"Octavia Butler", "Isaac Asimov"})
	p.Process()

	result := p.GetRedactedText(RedactOptions{})
	expected := "<PERSON_1> and <PERSON_2> are authors."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExtractBaseType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PERSON_1", "PERSON"},
		{"PERSON_12", "PERSON"},
		{"EMAIL_ADDRESS_1", "EMAIL_ADDRESS"},
		{"CREDIT_CARD_99", "CREDIT_CARD"},
		{"PERSON", "PERSON"},
		{"PERSON_ABC", "PERSON_ABC"}, // Not a number suffix
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExtractBaseType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
