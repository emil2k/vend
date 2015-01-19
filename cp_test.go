package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestCp copies a standard package to a temporary folder, then runs the cp
// to copy and update imports for a package it imports into its lib directory.
// Ran on a package, i.e. `unicode` that contains two packages in one folder to
// test how it handles a MultiplePackageError.
// Tests copying to a desitination folder where the parent folder
// doesn't exist, and must be created.
// Test makes sure imports and files match in the src and dst packages, and that
// the packages can be built.
func TestCp(t *testing.T) {
	goRoot, err := ioutil.TempDir(os.TempDir(), "testcp")
	if err != nil {
		t.Errorf("error when creating temp dir : %s\n", err.Error())
	}
	defer os.RemoveAll(goRoot)
	// Copy package over to tmp GOROOT
	tmp := build.Default
	tmp.GOROOT = goRoot
	_, cwdDst := copyPackage(t, &tmp, "encoding/json")
	_, cpDst := copyPackage(t, &tmp, "unicode")
	if err := cp(&tmp, cwdDst.Dir, "unicode", "lib/myuni"); err != nil {
		t.Errorf("error during cp : %s\n", err.Error())
	}
	// Check if import paths updated
	cpPkg, err := getPackage(&tmp, cwdDst.Dir, cwdDst.ImportPath)
	if err != nil {
		t.Errorf("error importing cped package : %s\n", err)
	}
	for _, i := range getImports(cpPkg, true) {
		if i == cpDst.ImportPath || isChildPackage(cpDst.ImportPath, i) {
			t.Errorf("old import found : %s\n", i)
		}
	}
}

// TestCopyParentDir tests the copyDir function when copying a parent directory
// into a child directory. Should avoid the situation where an infifinite copy
// procedure, that makes  copies of copies.
func TestCopyParentDir(t *testing.T) {
	parent, err := ioutil.TempDir(os.TempDir(), "testcopyparentdir")
	if err != nil {
		t.Errorf("error when creating temp dir : %s\n", err.Error())
	}
	defer os.RemoveAll(parent)
	child := filepath.Join(parent, "child")
	if err := copyDir(parent, child); err != nil {
		t.Errorf("error during copy : %s\n", err.Error())
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

// copyPackage copies a package in the default context to the GOROOT of the tmp
// package. Fails the test if any errors are reported, returns the
// *build.Package from the src and dst.
// Reports errors in the test if any errors are thrown or if the packages aren't
// equivalent afterwards the test will fail immediateley.
func copyPackage(t *testing.T, tmp *build.Context, path string) (src, dst *build.Package) {
	pkgPath, err := goRootPkgPath()
	if err != nil {
		t.Errorf("error determing package path : %s\n", err.Error())
		t.FailNow()
	}
	dstDir := filepath.Join(tmp.GOROOT, pkgPath, filepath.FromSlash(path))
	if src, err = getPackage(&build.Default, "", path); err != nil {
		t.Errorf("error importing src package : %s\n", err.Error())
		t.FailNow()
	}
	if err = copyDir(src.Dir, dstDir); err != nil {
		t.Errorf("error while copying directory : %s\n", err.Error())
		t.FailNow()
	}
	if dst, err = getPackage(tmp, "", path); err != nil {
		t.Errorf("error importing dst package : %s\n", err.Error())
		t.FailNow()
	}
	testPackagesEqual(t, src, dst)
	return
}

// testPackagesEqual tests if two packages are equal.
func testPackagesEqual(t *testing.T, a, b *build.Package) {
	switch {
	case !reflect.DeepEqual(a.GoFiles, b.GoFiles):
		t.Errorf("package files don't match, %s : %s\n",
			a.GoFiles, b.GoFiles)
	case !reflect.DeepEqual(a.Imports, b.Imports):
		t.Errorf("package imports don't match, %s : %s\n",
			a.Imports, b.Imports)
	case !reflect.DeepEqual(a.TestGoFiles, b.TestGoFiles):
		t.Errorf("package test files don't match, %s : %s\n",
			a.TestGoFiles, b.TestGoFiles)
	case !reflect.DeepEqual(a.TestImports, b.TestImports):
		t.Errorf("package test imports don't match, %s : %s\n",
			a.TestImports, b.TestImports)
	}
}

// goRootPkgPath returns the relative path to the directory containing standard
// packages sources from the GOROOT. This changed starting in 1.4 :
//
//	In Go 1.4, the pkg level of the source tree is now gone, so for example
//	the fmt package's source, once kept in directory src/pkg/fmt, now lives
//	one level higher in src/fmt.
//
// Returns an error if it cannot import the `fmt` package or determine a
// relative path.
func goRootPkgPath() (string, error) {
	pkg, err := build.Import("fmt", "", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("can't import standard package")
	}
	rel, err := filepath.Rel(build.Default.GOROOT, pkg.Dir)
	if err != nil {
		return "", fmt.Errorf("can't determine pkg path")
	}
	return filepath.Dir(rel), nil
}
