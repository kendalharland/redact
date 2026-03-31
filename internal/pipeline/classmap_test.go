package pipeline

import (
	"testing"
)

func TestClassificationMap_Add(t *testing.T) {
	cm := NewClassificationMap()

	cm.Add(Person, "John Doe")
	cm.Add(Person, "Jane Smith")
	cm.Add(Person, "John Doe") // Duplicate

	if len(cm[Person]) != 2 {
		t.Errorf("expected 2 entries, got %d", len(cm[Person]))
	}
}

func TestClassificationMap_Merge(t *testing.T) {
	cm1 := NewClassificationMap()
	cm1.Add(Person, "John")
	cm1.Add(EmailAddress, "john@test.com")

	cm2 := NewClassificationMap()
	cm2.Add(Person, "Jane")
	cm2.Add(Person, "John") // Duplicate
	cm2.Add(PhoneNumber, "123-456-7890")

	cm1.Merge(cm2)

	if len(cm1[Person]) != 2 {
		t.Errorf("expected 2 persons, got %d", len(cm1[Person]))
	}
	if len(cm1[EmailAddress]) != 1 {
		t.Errorf("expected 1 email, got %d", len(cm1[EmailAddress]))
	}
	if len(cm1[PhoneNumber]) != 1 {
		t.Errorf("expected 1 phone, got %d", len(cm1[PhoneNumber]))
	}
}

func TestAssignNumbers(t *testing.T) {
	cm := NewClassificationMap()
	cm.Add(Person, "Jane")
	cm.Add(Person, "John")

	text := "Hello John, meet Jane."

	ncm := AssignNumbers(cm, text)

	// John appears first, so should be PERSON_1
	if ncm["PERSON_1"] != "John" {
		t.Errorf("expected PERSON_1 to be John, got %s", ncm["PERSON_1"])
	}
	if ncm["PERSON_2"] != "Jane" {
		t.Errorf("expected PERSON_2 to be Jane, got %s", ncm["PERSON_2"])
	}
}

func TestAssignNumbers_MultipleTypes(t *testing.T) {
	cm := NewClassificationMap()
	cm.Add(Person, "John")
	cm.Add(Person, "Jane")
	cm.Add(EmailAddress, "john@test.com")
	cm.Add(PhoneNumber, "123-456-7890")

	text := "John's email is john@test.com. Jane's phone is 123-456-7890."

	ncm := AssignNumbers(cm, text)

	if ncm["PERSON_1"] != "John" {
		t.Errorf("expected PERSON_1 to be John, got %s", ncm["PERSON_1"])
	}
	if ncm["EMAIL_ADDRESS_1"] != "john@test.com" {
		t.Errorf("expected EMAIL_ADDRESS_1 to be john@test.com, got %s", ncm["EMAIL_ADDRESS_1"])
	}
	if ncm["PHONE_NUMBER_1"] != "123-456-7890" {
		t.Errorf("expected PHONE_NUMBER_1 to be 123-456-7890, got %s", ncm["PHONE_NUMBER_1"])
	}
}

func TestNumberedClassificationMap_GetPlaceholder(t *testing.T) {
	ncm := NumberedClassificationMap{
		"PERSON_1": "John",
		"PERSON_2": "Jane",
	}

	if ncm.GetPlaceholder("John") != "PERSON_1" {
		t.Errorf("expected PERSON_1, got %s", ncm.GetPlaceholder("John"))
	}
	if ncm.GetPlaceholder("Unknown") != "" {
		t.Errorf("expected empty string for unknown")
	}
}

func TestNumberedClassificationMap_GetPlaceholdersForType(t *testing.T) {
	ncm := NumberedClassificationMap{
		"PERSON_1":        "John",
		"PERSON_2":        "Jane",
		"EMAIL_ADDRESS_1": "test@test.com",
	}

	placeholders := ncm.GetPlaceholdersForType(Person)
	if len(placeholders) != 2 {
		t.Errorf("expected 2 placeholders, got %d", len(placeholders))
	}
	if placeholders[0] != "PERSON_1" || placeholders[1] != "PERSON_2" {
		t.Errorf("expected [PERSON_1 PERSON_2], got %v", placeholders)
	}
}

func TestClassificationMap_ToJSON(t *testing.T) {
	cm := NewClassificationMap()
	cm.Add(Person, "John")
	cm.Add(EmailAddress, "john@test.com")

	data, err := cm.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse it back
	parsed, err := ClassificationMapFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed[Person]) != 1 || parsed[Person][0] != "John" {
		t.Errorf("expected Person: [John], got %v", parsed[Person])
	}
}
