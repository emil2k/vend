package main

import (
	"fmt"
	"reflect"
	"testing"
)

// TestIsChildPackage tests the isChildPackage function with a posive and
// negative case.
func TestIsChildPackage(t *testing.T) {
	x := isChildPackage("github.com/emil2k/y", "github.com/emil2k/y/x")
	if !x {
		t.Error("expected true")
	}
	x = isChildPackage("github.com/emil2k/y/x", "github.com/emil2k/y")
	if x {
		t.Error("expected false")
	}
}

// TestIsStandardPackage tests if the image package is a standard package using
// the isStandardPackage function, and tests that this package is not.
func TestIsStandardPackage(t *testing.T) {
	if x := isStandardPackage("image"); !x {
		t.Error("expected true")
	}
	if x := isStandardPackage("github.com/emil2k/y"); x {
		t.Error("expected false")
	}
}

// filterImportTests table driven tests for the filterImports function.
var filterImportsTests = []struct {
	parent string
	imp    []string
	std    bool
	child  bool
	out    []string
}{
	{
		"github.com/emil2k/y",
		[]string{
			"github.com/lib/pq",
			"github.com/emil2k/y/x",
			"image",
		},
		false,
		true,
		[]string{
			"github.com/lib/pq",
			"github.com/emil2k/y/x",
		},
	},
	{
		"github.com/emil2k/y",
		[]string{
			"github.com/lib/pq",
			"github.com/emil2k/y/x",
			"image",
		},
		true,
		false,
		[]string{
			"github.com/lib/pq",
			"image",
		},
	},
	{
		"github.com/emil2k/y",
		[]string{
			"github.com/lib/pq",
			"github.com/emil2k/y/x",
			"image",
		},
		false,
		false,
		[]string{
			"github.com/lib/pq",
		},
	},
}

// TestFilterImports testing the filterImports funtction using a table driven
// tests.
func TestFilterImports(t *testing.T) {
	for _, tt := range filterImportsTests {
		pre := fmt.Sprintf("std? %t child? %t : ", tt.std, tt.child)
		out := filterImports(tt.parent, tt.imp, tt.std, tt.child)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf(pre+"got %s, expected %s\n", out, tt.out)
		}
	}
}
