package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// list runs the list subcommand, listing all the dependencies of the package
// at the specified path, relative paths are resolved from the current working
// directory.
func list(ctx *build.Context, cwd, path string) error {
	pkg, err := getPackage(ctx, path)
	if err != nil {
		return err
	}
	// List imports
	f := listFilter(ctx, cwd, pkg.ImportPath, opt.child, opt.standard)
	imp := filterImports(getImports(pkg, opt.tests), f)
	for _, v := range imp {
		info(ctx, cwd, v)
	}
	return nil
}

// listFilter makes an import filter for the list command for the package
// specified by the import path.
// Can specify whether to omit child or standard packages.
func listFilter(ctx *build.Context, cwd, path string, omitChild, omitStd bool) func(i string) bool {
	return func(i string) bool {
		switch {
		case omitChild && isChildPackage(path, i):
			return false
		case omitStd && isStandardPackage(ctx, cwd, i):
			return false
		}
		return true
	}
}

// info runs the info subcommand, printing information about a given package.
// Also used by the list command to output details about imports, the quite and
// verbose flags determine the output.
func info(ctx *build.Context, cwd, path string) error {
	if opt.quite {
		fmt.Println(path)
		return nil
	}
	pkg, err := getPackage(ctx, path)
	if err != nil {
		return err
	}
	// Default output
	printBold(fmt.Sprintf("%s (%s)", pkg.Name, pkg.ImportPath))
	// Print package doc with line breaks
	if len(pkg.Doc) > 0 {
		printWrap(72, pkg.Doc)
	} else {
		fmt.Println("No package documentation.")
	}
	// Verbose output
	if opt.verbose {
		fmt.Println("  Standard :\t", pkg.Goroot)
		fmt.Println("  Directory :\t", pkg.Dir)
		if len(pkg.AllTags) > 0 {
			fmt.Println("  Tags :\t",
				strings.Join(pkg.AllTags, " "))
		}
	}
	return nil
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
func getPackage(ctx *build.Context, path string) (pkg *build.Package, err error) {
	var stat os.FileInfo
	if abs, err := filepath.Abs(path); err == nil {
		// Withouth the absolute path, does not set the ImportPath
		// properly.
		if stat, err = os.Stat(path); err == nil && stat.IsDir() {
			return ctx.ImportDir(abs, 0)
		}
	}
	return ctx.Import(path, "", 0)
}
