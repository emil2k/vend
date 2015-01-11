package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sort"
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
	// List imports
	imp := getImports(pkg, opt.tests)
	fimp := filterImports(pkg.ImportPath, imp, opt.standard, opt.child)
	for _, v := range fimp {
		info(v)
	}
	return nil
}

// info runs the info subcommand, printing information about a given package.
// Also used by the list command to output details about imports, the quite and
// verbose flags determine the output.
func info(path string) error {
	if opt.quite {
		fmt.Println(path)
		return nil
	}
	pkg, err := getPackage(path)
	if err != nil {
		return err
	}
	// Default output
	printBold(pkg.ImportPath)
	// Print package doc with line breaks
	if len(pkg.Doc) > 0 {
		printWrap(72, pkg.Doc)
	} else {
		fmt.Println("No package documentation.")
	}
	// Verbose output
	if opt.verbose {
		fmt.Println()
		fmt.Println("  Standard : ", pkg.Goroot)
		fmt.Println("  Directory : ", pkg.Dir)
		if len(pkg.AllTags) > 0 {
			fmt.Println("  Tags : ",
				strings.Join(pkg.AllTags, " "))
		}
	}
	fmt.Println()
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
// XTestImports to Imports as necessary. Returns a sorted slice with unique
// elements.
func getImports(pkg *build.Package, includeTests bool) []string {
	imp := pkg.Imports
	if includeTests {
		imp = appendUnique(imp, pkg.TestImports)
		imp = appendUnique(imp, pkg.XTestImports)
	}
	sort.Strings(imp)
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
	if abs, err := filepath.Abs(path); err == nil {
		// Withouth the absolute path, does not set the ImportPath
		// properly.
		if stat, err = os.Stat(path); err == nil && stat.IsDir() {
			return ctx.ImportDir(abs, 0)
		}
	}
	return ctx.Import(path, cwd, 0)
}
