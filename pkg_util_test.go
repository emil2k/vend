package main

import (
	"fmt"
	"go/build"
	"path/filepath"
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

// isStandardPackageTests holds table tests for the isStandardPackage function.
var isStandardPackageTests = []struct {
	path     string
	standard bool // expectation
}{
	{"image", true},
	{"C", true},
	{"github.com/emil2k/vend", false},
	{"runtime", true}, // directory has multiple packages
	{"strconv", true}, // directory has multiple packages
	{"time", true},    // directory has multiple packages
}

// TestIsStandardPackage tests if the image package is a standard package using
// the isStandardPackage function, and tests that this package is not.
// Tests with build context that includes all files to emulate how packages are
// parsed by the tool.
func TestIsStandardPackage(t *testing.T) {
	ctx := &build.Default
	ctx.UseAllFiles = true
	for _, tt := range isStandardPackageTests {
		if x := isStandardPackage(ctx, "", tt.path); x != tt.standard {
			t.Errorf("%s standard %t, expected %t\n",
				tt.path, x, tt.standard)
		}
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

// TestFilterImports testing the filterImports function using a table driven
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

// getImportPathTests holds table driven tests to test getImportPath.
var getImportPathTests = []struct {
	ctx       *build.Context
	cwd, path string
	// expected
	imp string
	err error
}{
	{
		&build.Context{GOPATH: filepath.FromSlash("/gopath")},
		filepath.FromSlash("/somewhere/else/"),
		filepath.FromSlash("/gopath/src/github.com/example/vend"),
		"github.com/example/vend",
		nil,
	},
	{
		&build.Context{GOPATH: filepath.FromSlash("/gopath")},
		filepath.FromSlash("/gopath/src/github.com"),
		filepath.FromSlash("example/vend"),
		"github.com/example/vend",
		nil,
	},
	{
		&build.Context{GOPATH: filepath.FromSlash("/gopath")},
		filepath.FromSlash("/nopath"),
		filepath.FromSlash("example/vend"),
		"",
		ErrNotInGoPath,
	},
}

// TestGetImportPath tests getImportPath, tries passing it an absolute and
// relative path located in the GOPATH and one that is not in the GOPATH.
func TestGetImportPath(t *testing.T) {
	for _, tt := range getImportPathTests {
		imp, err := getImportPath(tt.ctx, tt.cwd, tt.path)
		if err != tt.err {
			t.Errorf("got %v, expected %v", err, tt.err)
		}
		if imp != tt.imp {
			t.Errorf("got %v, expected %v", imp, tt.imp)
		}
	}
}

// changePathParentTests are table tests for the changePathParent function.
var changePathParentTests = []struct {
	a, b, child, out string
	err              error
}{
	{"unicode", "encoding/json/lib/myuni", "unicode/utf8",
		"encoding/json/lib/myuni/utf8", nil},
	{"gopkg.in/yaml.v2/", "myyaml", "gopkg.in/yaml.v2/child",
		"myyaml/child", nil},
}

// TestChangePathParent tests the helper function changePathParent using a table
// driven test.
func TestChangePathParent(t *testing.T) {
	for _, tt := range changePathParentTests {
		if x, err := changePathParent(tt.a, tt.b, tt.child); err != tt.err {
			t.Errorf("error : %s\n", err.Error())
		} else if x != tt.out {
			t.Errorf("expected %s, got %s\n", tt.out, x)
		}
	}
}
