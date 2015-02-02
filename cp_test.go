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
		filepath.Join("lib", "y"), false)
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
		filepath.Join("lib", "y"), true)
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
	err := cp(ctx, pkgDir, filepath.Join("other.com", "y"), cpPkgDir, false)
	if err != nil {
		t.Errorf("error during cp : %s", err.Error())
	}
	testImports(t, pkgDir, []string{"other.com/y/lib/y"}, false)
	// Test that the import path updated in the external test package of the
	// copied package.
	testImports(t, cpPkgDir, []string{"other.com/y/lib/y"}, true)
}
