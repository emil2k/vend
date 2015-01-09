package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

// cwd is the current working directory.
var cwd string

// ctx is a build context based on build.Default but includes all files.
var ctx build.Context

// init sets the current working directory and initializes the build context.
func init() {
	var err error
	if cwd, err = os.Getwd(); err != nil {
		printErr("Error : " + err.Error())
		os.Exit(1)
	}
	ctx = build.Default
	ctx.UseAllFiles = true
}

// list runs the list subcommand, listing all the dependencies of the package
// at the specified path, relative paths are resolved from the current working
// directory.
func list(path string) error {
	pkg, err := getPackage(path)
	if err != nil {
		return err
	}
	if opt.verbose {
		printBold("GOROOT : " + build.Default.GOROOT)
		printBold("GOPATH : " + build.Default.GOPATH)
		printBold("IMPORT PATH : " + pkg.ImportPath)
		printBold("DIRECTORY : " + pkg.Dir)
		printBold("OBJECT : " + pkg.PkgObj)
	}
	// List imports
	imp := getImports(pkg, opt.tests)
	fimp := filterImports(pkg.ImportPath, imp, opt.standard, opt.child)
	for _, v := range fimp {
		fmt.Println(v)
	}
	return nil
}

// filterImports filters the imports by either ommitting or including standard
// and child packages, returns filtered slice of import paths.
func filterImports(parent string, imp []string, std, child bool) []string {
	r := make([]string, 0, len(imp))
	for _, v := range imp {
		// Filter child packages
		if !child && isChildPackage(parent, v) {
			continue
		}
		// Filter standard packages
		if !std && isStandardPackage(v) {
			continue
		}
		r = append(r, v)
	}
	return r
}

// isChildPackage checks if the child package is stationed in a subdirectory of
// the parent package. Pass in the import paths for both the parent and the
// child.
func isChildPackage(parent string, child string) bool {
	return strings.HasPrefix(child, parent)
}

// isStandardPackage checks if the package is located in the standard library.
// If an error is thrown during import assumes it is not in the standard library.
func isStandardPackage(path string) bool {
	if pkg, err := ctx.Import(path, cwd, build.FindOnly); err != nil {
		return false
	} else {
		return pkg.Goroot
	}
}

// getImports compiles a list of all the imports by appending TestImports and
// XTestImports to Imports as necessary. Returns a slice with unique elements.
func getImports(pkg *build.Package, includeTests bool) []string {
	if !includeTests {
		return pkg.Imports
	}
	imp := pkg.Imports
	for _, v := range pkg.TestImports {
		if !hasString(imp, v) {
			imp = append(imp, v)
		}
	}
	for _, v := range pkg.XTestImports {
		if !hasString(imp, v) {
			imp = append(imp, v)
		}
	}
	return imp
}

// getPackage compiles information about the package at the given path. First,
// it tries to resolve the path as a relative path to the current working
// directory and then as an import path inside the GOPATH.
// This function addresses issues with specifying a relative directory without
// a preceding `./` and it attempts to gather the information in a manner that
// allows it to set the ImportPath attribute.
// Returns an error if the resolved directory does not contain a buildable Go
// package.
// Compiles all the files ignoring build flags and any other build contstraints.
func getPackage(path string) (pkg *build.Package, err error) {
	var stat os.FileInfo
	if stat, err = os.Stat(path); err == nil && stat.IsDir() {
		// Withouth the absolute path, does not set the ImportPath
		// properly.
		pkg, err = ctx.ImportDir(filepath.Join(cwd, path), 0)
		if err != nil {
			return
		}
	} else if pkg, err = ctx.Import(path, cwd, 0); err != nil {
		return
	}
	return
}
