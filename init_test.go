package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInit tests the init subcommand in non recursive mode. Makes sure that the
// the import paths updated properly, but not altered in a child directory. Also
// tests that all the copied packages can be built.
func TestInit(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "init"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	err := initc(ctx, pkgDir, "lib", false, false)
	if err != nil {
		t.Errorf("error during init : %s", err.Error())
		t.FailNow()
	}
	// Test that the import paths updated.
	testImports(t, pkgDir,
		[]string{"example.com/x/lib/a", "example.com/x/lib/b"}, false)
	// Test that child import path not updated.
	childPkgDir := filepath.Join(pkgDir, "z")
	testImports(t, childPkgDir,
		[]string{"other.com/y/a1", "other.com/y/c"}, false)
	// Test that copied packages build.
	aDir := filepath.Join(pkgDir, "lib", "a")
	testBuild(t, aDir)
	bDir := filepath.Join(pkgDir, "lib", "b")
	testBuild(t, bDir)
}

// TestInitRecursive tests the init subcommand in recursive mode. Makes sure
// that the import paths updated in the child directory and that all the
// packages can be built.
// Tests the case where two packages import the same package, which should not
// throw a duplicate package name error.
func TestInitRecursive(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "init"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	err := initc(ctx, pkgDir, "lib", true, false)
	if err != nil {
		t.Errorf("error during init : %s", err.Error())
		t.FailNow()
	}
	// Test that the import paths updated.
	testImports(t, pkgDir,
		[]string{"example.com/x/lib/a", "example.com/x/lib/b"}, false)
	// Test that child import path not updated.
	childPkgDir := filepath.Join(pkgDir, "z")
	testImports(t, childPkgDir,
		[]string{"example.com/x/lib/a", "example.com/x/lib/c"}, false)
	// Test that copied packages build.
	aDir := filepath.Join(pkgDir, "lib", "a")
	testBuild(t, aDir)
	bDir := filepath.Join(pkgDir, "lib", "b")
	testBuild(t, bDir)
	cDir := filepath.Join(pkgDir, "lib", "c")
	testBuild(t, cDir)
}

// TestInitDupe tests that a duplicate package name error is thrown when two
// packages with the same name are found in a package attempting to init vending.
// Tests that the error message matches expectations.
func TestInitDupe(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "init"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "dupe")
	err := initc(ctx, pkgDir, "lib", false, false)
	dupe, ok := err.(errDupe)
	if err == nil || !ok {
		t.Errorf("should return a duplicate package name error")
		t.FailNow()
	}
	// Test that the import paths didn't update.
	testImports(t, pkgDir,
		[]string{"other.com/y/a1", "other.com/y/a2"}, false)
	// Test the error message.
	expected := "duplicate package names found :\na found at other.com/y/a1, other.com/y/a2"
	if dupe.Error() != expected {
		t.Errorf("error message did not match : got :\n%s\nexpected :\n%s",
			dupe.Error(), expected)
	}
}
