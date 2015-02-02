package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestUpdate test the update subcommand to make sure that it properly changes
// imports, including imports of child packages.
// Also this command tests the call without recursion into subdirectories,
// checks to make sure they are not affected.
func TestUpdate(t *testing.T) {
	ctx := getTestContextCopy(t, "testdata/update")
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	err := update(ctx, pkgDir, "go", "mygo", false)
	if err != nil {
		t.Errorf("update error : %s", err.Error())
	}
	testImports(t, pkgDir,
		[]string{"fmt", "os", "mygo/ast", "mygo/build", "mygo/parser"},
		false)
	// Test that child package was not updated.
	childPkgDir := filepath.Join(pkgDir, "y")
	testImports(t, childPkgDir,
		[]string{"fmt", "os", "go/ast", "go/parser"}, false)
}

// TestUpdateRecurse test the update subcommand with the recurse option. Makes
// sure subdirectory package is also updated.
func TestUpdateRecurse(t *testing.T) {
	ctx := getTestContextCopy(t, "testdata/update")
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	err := update(ctx, pkgDir, "go", "mygo", true)
	if err != nil {
		t.Errorf("update error : %s", err.Error())
	}
	testImports(t, pkgDir,
		[]string{"fmt", "os", "mygo/ast", "mygo/build", "mygo/parser"},
		false)
	// Test child package was updated as well.
	childPkgDir := filepath.Join(pkgDir, "y")
	testImports(t, childPkgDir,
		[]string{"fmt", "os", "mygo/ast", "mygo/parser"}, false)
}
