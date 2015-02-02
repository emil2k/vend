package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMv tests that the directory is removed after a move.
func TestMv(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "cp"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	srcDir := filepath.Join(ctx.GOPATH, "src", "other.com", "y")
	err := mv(ctx, pkgDir, "other.com/y", "lib/y", false)
	if err != nil {
		t.Errorf("error during mv : %s", err.Error())
	}
	// Test that the source directory has been removed.
	_, err = os.Stat(srcDir)
	if !os.IsNotExist(err) {
		t.Errorf("error src directory should not exist : %s", srcDir)
	}
}

// TestMvStandardPackage tests that an error is thrown when attempting to move
// a standard package.
func TestMvStandardPackage(t *testing.T) {
	ctx := getTestContextCopy(t, filepath.Join("testdata", "cp"))
	defer os.RemoveAll(ctx.GOPATH)
	pkgDir := filepath.Join(ctx.GOPATH, "src", "example.com", "x")
	err := mv(ctx, pkgDir, "fmt", "lib/fmt", false)
	if err != ErrStandardPackage {
		t.Errorf("moving standard package err : got %v, expected %v",
			err, ErrStandardPackage)
	}
}
