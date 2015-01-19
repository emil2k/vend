package main

import (
	"go/build"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// getPackage compiles information about the package at the given path. First,
// it tries to resolve the path as a relative path to the current working
// directory and then as an import path inside the GOPATH.
// This function addresses issues with specifying a relative directory without
// a preceding `./` and it attempts to gather the information in a manner that
// allows it to set the ImportPath attribute.
// Returns an error if the resolved directory does not contain a buildable Go
// package.
// Compiles all the files ignoring build flags and any other build contstraints.
func getPackage(ctx *build.Context, path string) (pkg *build.Package, err error) {
	var stat os.FileInfo
	// TODO pass ina cwd and change this to cwdAbs
	if abs, err := filepath.Abs(path); err == nil {
		// Withouth the absolute path, does not set the ImportPath
		// properly.
		if stat, err = os.Stat(path); err == nil && stat.IsDir() {
			return ctx.ImportDir(abs, 0)
		}
	}
	return ctx.Import(path, "", 0)
}

// getImports compiles a list of all the imports by appending TestImports and
// XTestImports to Imports as necessary. Returns a sorted slice with unique
// elements.
func getImports(pkg *build.Package, includeTests bool) []string {
	imp := pkg.Imports
	if includeTests {
		imp = appendUnique(imp, pkg.TestImports...)
		imp = appendUnique(imp, pkg.XTestImports...)
	}
	sort.Strings(imp)
	return imp
}

// filterImports filters the passed imports based on the passed filter function,
// which is passed in the import path and must return true if the import should
// be included.
func filterImports(imp []string, filter func(imp string) bool) []string {
	r := make([]string, 0, len(imp))
	for _, i := range imp {
		if filter(i) {
			r = append(r, i)
		}
	}
	return r
}

// isChildPackage checks if the child package is stationed in a subdirectory of
// the parent package. Pass in the import paths for both the parent and the
// child.
func isChildPackage(parent, child string) bool {
	return strings.HasPrefix(child, parent)
}

// isStandardPackage checks if the package is located in the standard library.
// If an error is thrown during import assumes it is not in the standard library.
func isStandardPackage(ctx *build.Context, cwd, path string) bool {
	// TODO use getpackage here
	if pkg, err := ctx.Import(path, cwd, build.FindOnly); err != nil {
		return false
	} else {
		return pkg.Goroot
	}
}
