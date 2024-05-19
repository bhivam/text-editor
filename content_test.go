package main

import (
	"testing"
)

func rune_cmp(a []rune, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range len(a) {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestReplace(t *testing.T) {
	content := &Content{
		original: []rune{},
		add:      []rune{},
		root:     nil,
		length:   0,
	}

	add_string := []rune("Hello world!")
	content.replace(add_string, 0, 0)

	final_string := content.calculate_content()

	if !rune_cmp(add_string, final_string) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(add_string),
			string(final_string),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(add_string),
			string(final_string),
		)
	}

	content.replace(add_string, 12, 12)

	final_string = content.calculate_content()

	if !rune_cmp(content.add, final_string) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(content.add),
			string(final_string),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(content.add),
			string(final_string),
		)
	}

	content.replace(add_string, 0, 0)

	final_string = content.calculate_content()

	if !rune_cmp(content.add, final_string) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(content.add),
			string(final_string),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(content.add),
			string(final_string),
		)
	}

	content.replace(add_string, 1, 1)

	final_string = content.calculate_content()
	expected_string := []rune(
		"HHello world!ello world!Hello world!Hello world!",
	)

	if !rune_cmp(expected_string, final_string) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final_string),
			string(expected_string),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(final_string),
			string(expected_string),
		)
	}

	content.replace(add_string, 24, 24)

	final_string = content.calculate_content()
	expected_string = []rune(
		"HHello world!ello world!Hello world!Hello world!Hello world!",
	)

	if !rune_cmp(expected_string, final_string) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final_string),
			string(expected_string),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(final_string),
			string(expected_string),
		)
	}
}

func TestReplaceRealistic(t *testing.T) {
	content := &Content{
		original: []rune("hey"),
		add:      []rune{},
		root:     &Piece{0, 3, original, nil},
		length:   3,
	}

	to_add := []rune{'h'}

	content.replace(to_add, 0, 0)
	expected := []rune("hhey")
	final := content.calculate_content()
	if !rune_cmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}

	content.replace(to_add, 1, 1)
	expected = []rune("hhhey")
	final = content.calculate_content()
	if !rune_cmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}

	content.replace(to_add, 2, 2)
	expected = []rune("hhhhey")
	final = content.calculate_content()
	if !rune_cmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}
}
