package main

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strings"
)

// initc runs the init subcommand, copies all the external packages for the
// package in the current working directory into the specified directory.
// External packages are packages not located in the standard library, a parent
// directory, or a subdirectory.
// Files are placed in subdirectories based on their package name, if there are
// conflicts the command will fail with a message, those specific packages will
// need to be copied with the cp command, before running init again.
// Includes dependencies from packages located in subdirectories based on the
// `recurse` parameter.
func initc(ctx *build.Context, cwd, dst string, recurse bool) error {
	dst, err := cwdAbs(cwd, dst)
	if err != nil {
		return err
	}
	cwdPkg, _ := getPackage(ctx, cwd, cwd)
	if len(cwdPkg.ImportPath) == 0 {
		return fmt.Errorf("no import path for package in current directory")
	}
	// Filter for the imports to copy into dst directory
	f := func(i string) bool {
		switch {
		case isChildPackage(cwdPkg.ImportPath, i):
			return false // in a subdirectory
		case isChildPackage(i, cwdPkg.ImportPath):
			return false // in a parent diretory
		case isStandardPackage(ctx, cwd, i):
			return false
		}
		return true
	}
	dsts := make([]string, 0)         // list of destination directories
	cps := make([]cpJob, 0)           // list of pending cp calls
	updates := make([]updateJob, 0)   // list of pending update calls
	dups := make(map[string][]string) // package name to import paths
	hasDups := false
	process := func(pkg *build.Package, err error) error {
		imp := filterImports(getImports(pkg, true), f)
		for _, i := range imp {
			cpPkg, _ := getPackage(ctx, cwd, i)
			if len(cpPkg.ImportPath) == 0 {
				return fmt.Errorf("no import path for %s", i)
			} else if len(cpPkg.Name) == 0 || len(cpPkg.Dir) == 0 {
				// Skip packages without a package name, most
				// likely they have not been retreived.
				fmt.Printf("skipping %s, was not found\n",
					cpPkg.ImportPath)
				continue
			}
			cpDst := filepath.Join(dst, cpPkg.Name)
			if hasString(dsts, cpDst) {
				hasDups = true
				if dstImpPath, err := getImportPath(ctx, cwd, cpDst); err != nil {
					return err
				} else {
					updates = append(updates,
						updateJob{pkg.Dir, cpPkg.ImportPath, dstImpPath, false})
				}
			} else {
				cps = append(cps,
					cpJob{pkg.Dir, cpPkg.ImportPath, cpDst, false})
			}
			dsts = append(dsts, cpDst)
			dups[cpPkg.Name] = appendUnique(dups[cpPkg.Name], cpPkg.ImportPath)
		}
		return nil
	}
	if recurse {
		if err := recursePackages(ctx, cwd, process); err != nil {
			return err
		}
	} else if err := process(cwdPkg, nil); err != nil {
		return err
	}
	// Report back if there is any packages with the same package name.
	if hasDups {
		return errDupe(dups)
	}
	// Run copy command on each import.
	for _, cj := range cps {
		printBold(fmt.Sprintf("%s => %s", cj.src, cj.dst))
		if err := cp(ctx, cj.cwd, cj.src, cj.dst, cj.recurse); err != nil {
			return err
		}
	}
	// Run update commands on other packages that need updating.
	for _, uj := range updates {
		if err := update(ctx, uj.src, uj.from, uj.to, uj.recurse); err != nil {
			return err
		}
	}
	return nil
}

// errDupe is returned when there are duplicate package names when trying to
// run the init command.
// Underlying map is package name to a slice of import paths.
type errDupe map[string][]string

func (d errDupe) Error() string {
	errs := make([]string, 0)
	for name, paths := range d {
		if len(paths) > 1 {
			errs = append(errs, fmt.Sprintf("%s found at %s",
				name, strings.Join(paths, ", ")))
		}
	}
	return fmt.Sprintf("duplicate packages names found :\n%s",
		strings.Join(errs, "\n"))

}

// cpJob holds a pending call to cp.
type cpJob struct {
	cwd, src, dst string
	recurse       bool
}

// updateJob holds a pending call to update.
type updateJob struct {
	src, from, to string
	recurse       bool
}
