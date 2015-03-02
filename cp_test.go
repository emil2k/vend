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
		t.Fatalf("error during cp : %s", err.Error())
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

// TestCpStripCanonicalImportPaths tests that the cannonical import paths are
// stripped from copied packages when running the cp command, otherwise Go will
// not let it build in its new location.
// Checks that the import path is stripped both in the main copied package and
// in a copied child package.
func TestCpStripCanonicalImportPaths(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "cp"))
	//defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	dstDir := filepath.Join(pkgDir, "lib", "y")
	err := cp(ctx, pkgDir, filepath.Join("other.com", "y"),
		dstDir, false, false)
	if err != nil {
		t.Fatalf("error during cp : %s", err.Error())
	}
	// Test to see if there are any cannonical import paths in either the
	// main copied package or its copied child package.
	testStrippedCanonicalImportPath(t, filepath.Join(dstDir, "y.go"))
	testStrippedCanonicalImportPath(t, filepath.Join(dstDir, "sub", "sub.go"))
	// Test that the packages build without error after stripping.
	if _, err := ctx.ImportDir(dstDir, 0); err != nil {
		t.Errorf("main copied package did not build : %v", err)
	}
	if _, err := ctx.ImportDir(filepath.Join(dstDir, "sub"), 0); err != nil {
		t.Errorf("sub copied package did not build : %v", err)
	}
}

// testStrippedCannonicalImportPath tests whether the file contains a canonical
// import path or not.
func testStrippedCanonicalImportPath(t *testing.T, path string) {
	src, err := getFileContents(path)
	if err != nil {
		t.Fatal(err)
	}
	contains, start, end := containsCanonicalImportPath(src)
	if contains {
		lit := string(src[start:end])
		t.Errorf("canonical import path found on line number in file %s : %s",
			path, lit)
	}
}

// TestCpIncludeHiddeFiles tests cp with including hidden files.
func TestCpIncludeHiddenFiles(t *testing.T) {
	testHiddenFiles(t, true)
}

// TestCpIgnoreHiddeFiles tests cp with ignoring hidden files.
func TestCpIgnoreHiddenFiles(t *testing.T) {
	testHiddenFiles(t, false)
}

// testHidden constructs a test based on the parameter whether to keep hidden
// files or not.
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
