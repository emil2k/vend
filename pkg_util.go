package main

import (
	"errors"
	"go/build"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ErrPseudoPackage is returned when attempting to access a pseudo package like
// "C".
var ErrPseudoPackage = errors.New("pseudo package")

// cPackage is returned when trying to import the "C" package.
var cPackage = &build.Package{
	ImportPath: "C",
	Goroot:     true,
	Doc:        "Package C is a pseudo package that enables calls to C code via cgo.",
}

// getPackage compiles information about the package at the given path. First,
// it tries to resolve the path as a relative path to the passed current working
// directory and then as an import path inside the GOPATH.
// This function addresses issues with specifying a relative directory without
// a preceding `./` and it attempts to gather the information in a manner that
// allows it to set the ImportPath attribute.
// Returns an error if the resolved directory does not contain a buildable Go
// package.
// For the special pseudo-package "C" it returns a partial package with an
// import path, Goroot true, and doc string along with an ErrPseudoPackage.
// Compiles all the files ignoring build flags and any other build contstraints.
func getPackage(ctx *build.Context, cwd, path string) (pkg *build.Package, err error) {
	var stat os.FileInfo
	if abs, err := cwdAbs(cwd, path); err == nil {
		// Withouth the absolute path, does not set the ImportPath
		// properly.
		if stat, err = os.Stat(path); err == nil && stat.IsDir() {
			return ctx.ImportDir(abs, 0)
		}
	}
	// Handle special case of the pseudo
	if path == "C" {
		return cPackage, ErrPseudoPackage
	}
	return ctx.Import(path, "", 0)
}

// packageResult holds a result from calling importing a package.
type packageResult struct {
	pkg *build.Package
	err error
}

// recursePackages recurses all the directories starting with the specified
// directory, called the passed function for all the packages that are found.
// Compile list of packages then calls function on them, as the function may
// change the packages themselves.
// Returns an error if the passed function returns an error for any of the found
// packages and in case of permissions issues during recursion.
func recursePackages(ctx *build.Context, dir string, f func(p *build.Package, err error) error) error {
	pkgs := make([]packageResult, 0)
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if !info.IsDir() { // only directories
			return nil
		} else if rel, err := filepath.Rel(dir, path); err != nil {
			return err
		} else if rel != "." && rel != ".." &&
			strings.HasPrefix(rel, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		pkg, err := getPackage(ctx, dir, path)
		pkgs = append(pkgs, packageResult{pkg, err})
		return nil
	}
	if err := filepath.Walk(dir, walk); err != nil {
		return err
	}
	// Call function on packages
	for _, p := range pkgs {
		if err := f(p.pkg, p.err); err != nil {
			return err
		}
	}
	return nil
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
	pkg, _ := getPackage(ctx, cwd, path)
	return pkg.Goroot
}
