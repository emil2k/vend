package main

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strings"
)

// dupError is returned when there are duplicate package names when trying to
// run the init command.
// Underlying map is package name to a slice of import paths.
type dupError map[string][]string

func (d dupError) Error() string {
	errs := make([]string, 0)
	for name, paths := range d {
		if len(paths) > 1 {
			errs = append(errs, fmt.Sprintf("%s found at %s",
				name, strings.Join(paths, ", ")))
		}
	}
	return fmt.Sprintf("duplicate packages names found : %s",
		strings.Join(errs, ", "))

}

// initc runs the init subcommand, copies all the external packages for the
// package in the current working directory into the specified directory.
// External packages are packages not located in the standard library, a parent
// directory, or a subdirectory.
// Files are placed in subdirectories based on their package name, if there are
// conflicts the command will fail with a message, those specific packages will
// need to be copied with the cp command, before running init again.
func initc(ctx *build.Context, cwd, dst string) error {
	pkg, err := getPackage(ctx, cwd, cwd)
	if err != nil {
		return err
	}
	f := func(i string) bool {
		switch {
		case isChildPackage(pkg.ImportPath, i): // in subdirectory
			return false
		case isChildPackage(i, pkg.ImportPath): // in parent directory
			return false
		case isStandardPackage(ctx, cwd, i):
			return false
		}
		return true
	}
	imp := filterImports(getImports(pkg, true), f)
	cps := make(map[string]string)    // dst directory to import path
	dups := make(map[string][]string) // package name to import paths
	hasDups := false
	for _, i := range imp {
		cpPkg, err := getPackage(ctx, cwd, i)
		if err != nil {
			return err
		}
		cpDst := filepath.Join(dst, cpPkg.Name)
		if _, exist := cps[cpDst]; exist {
			hasDups = true
		}
		cps[cpDst] = cpPkg.ImportPath
		dups[cpPkg.Name] = append(dups[cpPkg.Name], cpPkg.ImportPath)
	}
	if hasDups {
		return dupError(dups)
	}
	// Run copy command on each import
	for cpDst, cpPath := range cps {
		printBold(fmt.Sprintf("%s => %s", cpPath, cpDst))
		if err := cp(ctx, cwd, cpPath, cpDst); err != nil {
			return err
		}
	}
	return nil
}
