package main

import (
	"fmt"
	"go/build"
	"sort"
	"strings"
)

// list runs the list subcommand, listing all the dependencies of the package
// at the specified path, relative paths are resolved from the current working
// directory.
func list(ctx *build.Context, cwd, path string) error {
	imps := make([]string, 0)
	var parentPkg *build.Package
	process := func(pkg *build.Package, err error) error {
		// Set the parent package so child filters work properly as the
		// command recurses.
		if parentPkg == nil {
			parentPkg = pkg
		}
		f := listFilter(ctx, cwd, parentPkg.ImportPath, opt.child, opt.standard)
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
	} else if pkg, err := getPackage(ctx, cwd, path); err != nil {
		return err
	} else {
		process(pkg, nil)
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
	pkg, err := getPackage(ctx, cwd, path)
	// Error could be that the directory had multiple packages, if the
	// import path was determined proceed.
	if err != nil && len(pkg.ImportPath) == 0 {
		return err
	}
	// Default output
	if len(pkg.Name) == 0 {
		printBold(fmt.Sprintf("%s", pkg.ImportPath))
	} else {
		printBold(fmt.Sprintf("%s (%s)", pkg.ImportPath, pkg.Name))
	}
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
