package main

import (
	"testing"
)

// TestHasString tests the utility function hasString, checking both when a
// string is present and not present.
func TestHasString(t *testing.T) {
	test := []string{"a", "b", "c"}
	if x := hasString(test, "a"); !x {
		t.Errorf("expected to have")
	}
	if x := hasString(test, "d"); x {
		t.Errorf("expeted no to have")
	}
}
