package main

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strings"
)

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

// initc runs the init subcommand, copies all the external packages for the
// package in the current working directory into the specified directory.
// Recurses into sbudirectories depending on the opt.recurse option.
// External packages are packages not located in the standard library, a parent
// directory, or a subdirectory.
// Files are placed in subdirectories based on their package name, if there are
// conflicts the command will fail with a message, those specific packages will
// need to be copied with the cp command, before running init again.
func initc(ctx *build.Context, cwd, dst string) error {
	dst, err := cwdAbs(cwd, dst)
	if err != nil {
		return err
	}

	dsts := make([]string, 0)         // list of destination directories
	cps := make([]cpJob, 0)           // list of pending cp calls
	dups := make(map[string][]string) // package name to import paths
	hasDups := false

	process := func(pkg *build.Package, err error) error {
		// Filter for the imports to copy into dst directory
		f := func(i string) bool {
			switch {
			case isChildPackage(pkg.ImportPath, i):
				return false // in a subdirectory
			case isChildPackage(i, pkg.ImportPath):
				return false // in a parent diretory
			case isStandardPackage(ctx, cwd, i):
				return false
			}
			return true
		}
		imp := filterImports(getImports(pkg, true), f)
		for _, i := range imp {
			cpPkg, _ := getPackage(ctx, cwd, i)
			if len(cpPkg.ImportPath) == 0 {
				return fmt.Errorf("no import path for %s", i)
			} else if len(cpPkg.Name) == 0 || len(cpPkg.Dir) == 0 {
				// Skip packages without a package name, most
				// likely they have not been retrieved.
				fmt.Printf("skipping %s, was not found\n",
					cpPkg.ImportPath)
				continue
			}
			cpDst := filepath.Join(dst, cpPkg.Name)
			if hasString(dsts, cpDst) {
				hasDups = true
				cps = append(cps,
					cpJob{pkg.Dir, cpPkg.ImportPath, cpDst, true})
			} else {
				cps = append(cps,
					cpJob{pkg.Dir, cpPkg.ImportPath, cpDst, false})
			}
			dsts = append(dsts, cpDst)
			dups[cpPkg.Name] = appendUnique(dups[cpPkg.Name], cpPkg.ImportPath)
		}
		return nil
	}
	if opt.recurse {
		if err := recursePackages(ctx, cwd, process); err != nil {
			return err
		}
	} else if pkg, err := getPackage(ctx, cwd, cwd); err != nil {
		return err
	} else if err := process(pkg, nil); err != nil {
		return err
	}
	// Report back if there is any packages with the same package name.
	if hasDups {
		return errDupe(dups)
	}
	// Run copy command on each import
	for _, cj := range cps {
		printBold(fmt.Sprintf("%s => %s", cj.src, cj.dst))
		fmt.Println("update imports in :", cj.cwd)
		if err := cp(ctx, cj.cwd, cj.src, cj.dst, cj.skip); err != nil {
			return err
		}
	}
	return nil
}

// cpJob holds a pending call to cp.
type cpJob struct {
	cwd, src, dst string
	skip          bool
}
