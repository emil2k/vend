package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCp tests the cp subcommand checking that imports are updated in the
// package in the current working directory and in an in the copied package
// itself. Also tests that child packages are not updated when not called
// without the recursive option.
func TestCp(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "cp"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	err := cp(ctx, pkgDir, filepath.Join("other.com", "y"),
		filepath.Join("lib", "y"), false, false)
	if err != nil {
		t.Errorf("error during cp : %s", err.Error())
	}
	testImports(t, pkgDir, []string{"example.com/x/lib/y"}, false)
	// Test that the import path updated in the external test package of the
	// copied package.
	cpPkgDir := filepath.Join(pkgDir, "lib", "y")
	testImports(t, cpPkgDir, []string{"example.com/x/lib/y"}, true)
	// Test that imports updated in the child package.
	childPkgDir := filepath.Join(pkgDir, "z")
	testImports(t, childPkgDir, []string{"other.com/y"}, true)
}

// TestCpRecursive tests the recursive cp command, making sure that import paths
// are updated in the child packages.
func TestCpRecursive(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "cp"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	err := cp(ctx, pkgDir, filepath.Join("other.com", "y"),
		filepath.Join("lib", "y"), true, false)
	if err != nil {
		t.Errorf("error during cp : %s", err.Error())
	}
	testImports(t, pkgDir, []string{"example.com/x/lib/y"}, false)
	// Test that imports updated in the child package.
	childPkgDir := filepath.Join(pkgDir, "z")
	testImports(t, childPkgDir, []string{"example.com/x/lib/y"}, true)
}

// TestCpInfinite test calling the cp command to copy a package into its own
// subdirectory, it should not cause an infinite process where it copies copies
// of copies.
func TestCpInfinite(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "cp"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	cpPkgDir := filepath.Join(ctx.GOPATH, "src", "other.com", "y", "lib", "y")
	err := cp(ctx, pkgDir, filepath.Join("other.com", "y"), cpPkgDir, false, false)
	if err != nil {
		t.Errorf("error during cp : %s", err.Error())
	}
	testImports(t, pkgDir, []string{"other.com/y/lib/y"}, false)
	// Test that the import path updated in the external test package of the
	// copied package.
	testImports(t, cpPkgDir, []string{"other.com/y/lib/y"}, true)
}

// TestCpIncludeHiddeFiles tests cp with including hidden files.
func TestCpIncludeHiddenFiles(t *testing.T) {
	testHiddenFiles(t, true)
}

// TestCpIgnoreHiddeFiles tests cp with ignoring hidden files.
func TestCpIgnoreHiddenFiles(t *testing.T) {
	testHiddenFiles(t, false)
}

// testHidden constructs a test based on the parameter whether to keep hidden files
// or not.
func testHiddenFiles(t *testing.T, keepHidden bool) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "cp"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	dstPkgDir := filepath.Join(pkgDir, "lib", "y")
	err := cp(ctx, pkgDir, filepath.Join("other.com", "y"),
		filepath.Join("lib", "y"), false, keepHidden)
	if err != nil {
		t.Errorf("error during cp : %s", err.Error())
	}
	mainHiddenPath := filepath.Join(dstPkgDir, ".hidden")
	subHiddenPath := filepath.Join(dstPkgDir, "sub", ".hidden")
	// Test hidden file presense matches expectations.
	testExists(t, mainHiddenPath, keepHidden)
	testExists(t, subHiddenPath, keepHidden)
}

// testExists whether a file exists or not at the given path, fails the test if
// accessing the file results in an error.
func testExists(t *testing.T, path string, exists bool) {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		fallthrough
	case os.IsExist(err):
		if !exists {
			t.Fatalf("file exists, when it shouldn't : %s : %v", path, err)
		}
	case os.IsNotExist(err):
		if exists {
			t.Fatalf("file does not exist, when it should : %s : %v", path, err)
		}
	}
}
