package backend 

import (
	"fmt"
	"testing"
)

func runeCmp(a []rune, b []rune) bool {
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
		original:     []rune{},
		add:          []rune{},
		contentRoot: nil,
		length:       0,
	}

	addString := []rune("Hello world!")
	content.replace(addString, 0, 0)

	finalString := content.calculateContent()

	if !runeCmp(addString, finalString) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(addString),
			string(finalString),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(addString),
			string(finalString),
		)
	}

	content.replace(addString, 12, 12)

	finalString = content.calculateContent()

	if !runeCmp(content.add, finalString) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(content.add),
			string(finalString),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(content.add),
			string(finalString),
		)
	}

	content.replace(addString, 0, 0)

	finalString = content.calculateContent()

	if !runeCmp(content.add, finalString) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(content.add),
			string(finalString),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(content.add),
			string(finalString),
		)
	}

	content.replace(addString, 1, 1)

	finalString = content.calculateContent()
	expectedString := []rune(
		"HHello world!ello world!Hello world!Hello world!",
	)

	if !runeCmp(expectedString, finalString) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(finalString),
			string(expectedString),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(finalString),
			string(expectedString),
		)
	}

	content.replace(addString, 24, 24)

	finalString = content.calculateContent()
	expectedString = []rune(
		"HHello world!ello world!Hello world!Hello world!Hello world!",
	)

	if !runeCmp(expectedString, finalString) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(finalString),
			string(expectedString),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(finalString),
			string(expectedString),
		)
	}
}

func TestReplaceRealistic(t *testing.T) {
	content := &Content{
		original:     []rune("hey"),
		add:          []rune{},
		contentRoot: &Piece{0, 3, original, nil},
		length:       3,
	}

	toAdd := []rune{'h'}

	content.replace(toAdd, 0, 0)
	expected := []rune("hhey")
	final := content.calculateContent()
	if !runeCmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}

	content.replace(toAdd, 1, 1)
	expected = []rune("hhhey")
	final = content.calculateContent()
	if !runeCmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}

	content.replace(toAdd, 2, 2)
	expected = []rune("hhhhey")
	final = content.calculateContent()
	if !runeCmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}
}

func TestReplaceDelete(t *testing.T) {
	content := &Content{
		original:     []rune("hey"),
		add:          []rune{},
		contentRoot: &Piece{0, 3, original, nil},
		length:       3,
	}

	content.replace([]rune{}, 0, 1)

	expected := []rune("ey")
	final := content.calculateContent()
	if !runeCmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}

	content = &Content{
		original:     []rune("hey"),
		add:          []rune{},
		contentRoot: &Piece{0, 3, original, nil},
		length:       3,
	}

	content.replace([]rune{}, 1, 2)

	expected = []rune("hy")
	final = content.calculateContent()
	if !runeCmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}

	content = &Content{
		original:     []rune("hey"),
		add:          []rune{},
		contentRoot: &Piece{0, 3, original, nil},
		length:       3,
	}

	content.replace([]rune{}, 2, 3)

	expected = []rune("he")
	final = content.calculateContent()
	if !runeCmp(expected, final) {
		t.Fatalf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	} else {
		t.Logf(
			"\nFinal String: %s\nExpected String: %s",
			string(final),
			string(expected),
		)
	}
}

func TestReplaceRepeatDelete(t *testing.T) {
	content := &Content{
		original:     []rune("Hello, world!"),
		add:          []rune{},
		contentRoot: &Piece{0, 13, original, nil},
		length:       13,
	}
	content.replace([]rune{}, 0, 1)
	content.replace([]rune{}, 0, 1)
	content.replace([]rune{}, 0, 1)
}

func TestReplaceAddDelete(t *testing.T) {
	fmt.Println("MAIN TEST")
	content := &Content{
		original:     []rune("he\nre"),
		add:          []rune{},
		contentRoot: &Piece{0, 5, original, nil},
		length:       5,
	}

	content.printPieces()
	fmt.Println(content.calculateContent())
	fmt.Println()
	content.replace([]rune{'a'}, 4, 4)
	content.printPieces()
	fmt.Println(content.calculateContent())
}
