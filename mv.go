package main

import (
	"errors"
	"go/build"
	"os"
)

// ErrStandardPackage is returned when a subcommand is attempted on a standard
// package, but it prohibts execution on standard packages.
var ErrStandardPackage = errors.New("standard package specified")

// mv moves the package at the specified path to the specified destination
// directory.
// Just like cp, but cannot be used with standard packages and removes the
// source directory afterwards.
func mv(ctx *build.Context, cwd, src, dst string, recurse bool) (err error) {
	// Ignore the error because the directory itself might not be a package
	// but may contain subdirectories that do, all we want to know here is
	// if it is in the GOROOT.
	srcPkg, _ := getPackage(ctx, cwd, src)
	if srcPkg.Goroot {
		return ErrStandardPackage
	}
	if err := cp(ctx, cwd, src, dst, recurse); err != nil {
		return err
	}
	return os.RemoveAll(srcPkg.Dir)
}
