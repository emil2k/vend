package main

import (
	"fmt"
	"go/build"
	"reflect"
	"testing"
)

// TestIsChildPackage tests the isChildPackage function with a posive and
// negative case.
func TestIsChildPackage(t *testing.T) {
	x := isChildPackage("github.com/emil2k/vend", "github.com/emil2k/vend/x")
	if !x {
		t.Error("expected true")
	}
	x = isChildPackage("github.com/emil2k/vend/x", "github.com/emil2k/vend")
	if x {
		t.Error("expected false")
	}
}

// TestIsStandardPackage tests if the image package is a standard package using
// the isStandardPackage function, and tests that this package is not.
func TestIsStandardPackage(t *testing.T) {
	if x := isStandardPackage(&build.Default, "", "image"); !x {
		t.Error("expected true")
	}
	if x := isStandardPackage(&build.Default, "", "C"); !x {
		t.Error("expected true")
	}
	if x := isStandardPackage(&build.Default, "", "github.com/emil2k/vend"); x {
		t.Error("expected false")
	}
}

// filterImportTests table driven tests for the filterImports function.
var filterImportsTests = []struct {
	parent string
	imp    []string
	std    bool // whether to omit standard packages
	child  bool // whether to omit child packages
	out    []string
}{
	{
		"github.com/emil2k/vend",
		[]string{
			"github.com/lib/pq",
			"github.com/emil2k/vend/x",
			"image",
		},
		true,
		false,
		[]string{
			"github.com/lib/pq",
			"github.com/emil2k/vend/x",
		},
	},
	{
		"github.com/emil2k/vend",
		[]string{
			"github.com/lib/pq",
			"github.com/emil2k/vend/x",
			"image",
		},
		false,
		true,
		[]string{
			"github.com/lib/pq",
			"image",
		},
	},
	{
		"github.com/emil2k/vend",
		[]string{
			"github.com/lib/pq",
			"github.com/emil2k/vend/x",
			"image",
		},
		true,
		true,
		[]string{
			"github.com/lib/pq",
		},
	},
}

// TestFilterImports testing the filterImports funtction using a table driven
// tests. Uses the listFilter function to generate the filter.
func TestFilterImports(t *testing.T) {
	for _, tt := range filterImportsTests {
		pre := fmt.Sprintf("std? %t child? %t : ", tt.std, tt.child)
		f := listFilter(&build.Default, "", tt.parent, tt.child, tt.std)
		out := filterImports(tt.imp, f)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf(pre+"got %s, expected %s\n", out, tt.out)
		}
	}
}
