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
	imps := make([]string, 0)
	process := func(pkg *build.Package) error {
		f := listFilter(ctx, cwd, pkg.ImportPath, opt.child, opt.standard)
		for _, add := range filterImports(getImports(pkg, opt.tests), f) {
			imps = appendUnique(imps, add)
		}
		return nil
	}
	// Compile list of unique import paths, recurse if asked.
	if opt.recurse {
		if abs, err := cwdAbs(cwd, path); err != nil {
			return err
		} else if err := recursePackages(ctx, abs, process); err != nil {
			return err
		}
	} else if pkg, err := getPackage(ctx, path); err != nil {
		return err
	} else {
		process(pkg)
	}
	// Output the imports
	sort.Strings(imps)
	for _, imp := range imps {
		if opt.quite {
			fmt.Println(imp)
		} else if err := info(ctx, cwd, imp); err != nil {
			return err
		}
	}
	return nil
}

// recursePackages recurses all the directories starting with the specified
// directory, called the passed function for all the packages that are found.
// Returns an error if the passed function returns an error for any of the found
// packages and in case of permissions issues during recursion.
func recursePackages(ctx *build.Context, dir string, f func(p *build.Package) error) error {
	// Compile list of functions first then call function on them, as the
	// function may change the packages themselves.
	pkgs := make([]*build.Package, 0)
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
		// Directories may contain multiple packages ( an external test
		// package ), which causes errors when importing with the
		// go/build package, this ignores those errors and if .go files
		// are found in a directory a package is returned.
		pkg, _ := getPackage(ctx, path)
		if len(pkg.GoFiles) > 0 || len(pkg.CgoFiles) > 0 ||
			len(pkg.TestGoFiles) > 0 || len(pkg.XTestGoFiles) > 0 {
			pkgs = append(pkgs, pkg)
		}
		return nil
	}
	if err := filepath.Walk(dir, walk); err != nil {
		return err
	}
	// Call function on packages
	for _, p := range pkgs {
		if err := f(p); err != nil {
			return err
		}
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
	pkg, err := getPackage(ctx, path)
	// Error could be that the directory had multiple packages, if the
	// import path was determined proceed.
	if err != nil && len(pkg.ImportPath) == 0 {
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
