package main

import (
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrDstExist is thrown when attempting to copy to a destination that already
// exists with out using the force flag.
var ErrDstExists = errors.New("destination already exists")

// cp copies the package at the specified path to the specified destination
// directory. Update import paths for the copied package in the package
// located in the current working directory. Imports paths for child packages,
// packages located in subdirectories of the copied package, are also updated.
// Returns errors if the source package cannot be found or built, or if an
// absolute path cannot be determined either for the source or destination.
// Updates imports inside the copied package itself, may have a test package
// inside the directory that imports itself.
// Recurses into subdirectories to update the import path of the copied package
// or any of its child packages based on the `recurse` parameter.
func cp(ctx *build.Context, cwd, src, dst string, recurse bool) (err error) {
	// Check if destination folder exists and based on force flag determine
	// action.
	if _, serr := os.Stat(dst); serr == nil {
		if opt.force {
			os.RemoveAll(dst)
		} else {
			return ErrDstExists
		}
	} else if !os.IsNotExist(serr) {
		// Some other error with destination
		return serr
	}
	var srcImp, dstImp string
	var srcPkg, dstPkg *build.Package
	// May fail because there is multiple packages in the folder but all
	// that is necessary here is the directory and the import path.
	// Can't use the build.MultiplePackageError, to detect the error because
	// it was only added in 1.4, and we want 1.2+.
	if srcPkg, err = getPackage(ctx, cwd, src); len(srcPkg.Dir) == 0 {
		if err == nil {
			return fmt.Errorf("package has no directory")
		}
		return err
	} else if srcImp = srcPkg.ImportPath; len(srcImp) == 0 {
		if err == nil {
			return fmt.Errorf("package has no import path")
		}
		return err
	} else if src, err = cwdAbs(cwd, srcPkg.Dir); err != nil {
		return err
	} else if dst, err = cwdAbs(cwd, dst); err != nil {
		return err
	}
	// Copy the package over.
	if err = copyDir(src, dst); err != nil {
		return err
	}
	// Determine import path of the new package, and update import paths in
	// the current working directory.
	// Update the import paths of the new package and its children.
	if dstPkg, err = getPackage(ctx, cwd, dst); len(dstPkg.ImportPath) == 0 {
		return err
	} else {
		dstImp = dstPkg.ImportPath
	}
	// Update import paths in the copied package itself, as it may contain
	// an external _test package that imports itself or may contain packages
	// in its subdirectories that import it, must recurse.
	if err := update(ctx, dst, srcImp, dstImp, true); err != nil {
		return err
	}
	// Update the import paths, if the recurse flag is set recurse through
	// the subdirectories and update import paths.
	return update(ctx, cwd, srcImp, dstImp, recurse)
}

// ErrStandardPackage is returned when a subcommand is attempted on a standard
// package, but it prohibts execution on standard packages.
var ErrStandardPackage = errors.New("standard package specified")

// mv moves the package at the specified path to the specified destination
// directory.
// Just like cp, but cannot be used with standard packages and removes the
// source directory afterwards.
func mv(ctx *build.Context, cwd, src, dst string, recurse bool) (err error) {
	// Ignore the error because the directory itself might not be a package
	// but may contain subdirectories that do, all we want to know here is
	// if it is in the GOROOT.
	srcPkg, _ := getPackage(ctx, cwd, src)
	if srcPkg.Goroot {
		return ErrStandardPackage
	}
	if err := cp(ctx, cwd, src, dst, recurse); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

// copyFileJob holds a pending copyFile call.
type copyFileJob struct {
	si       os.FileInfo
	src, dst string
}

// copyDir recursively copies the src directory to the desination directory.
// Creates directories as necessary. Attempts to chmod everything to the src
// mode.
// With the opt.verbose option set outputs the src and destination of each
// copied file.
// Skips hidden files base on the opt.hidden option.
func copyDir(src, dst string) error {
	// First compile a list of copies to execute then execute, otherwise
	// infinite copy situations could arise when copying a parent directory
	// into a child directory.
	cjs := make([]copyFileJob, 0)
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel != "." && rel != ".." &&
			!opt.hidden && strings.HasPrefix(rel, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		fileDst := filepath.Join(dst, rel)
		if opt.verbose {
			printBold(path)
			fmt.Println(fileDst)
		}
		cjs = append(cjs, copyFileJob{info, path, fileDst})
		return nil
	}
	if err := filepath.Walk(src, walk); err != nil {
		return err
	}
	// Execute copies
	for _, cj := range cjs {
		if err := copyFile(cj.si, cj.src, cj.dst); err != nil {
			return err
		}
	}
	return nil
}

// ErrIrregularFile is returned when attempts are made to copy links, pipes,
// devices, and etc.
var ErrIrregularFile = errors.New("non regular file")

// copyFile copies a file or directory from src to dst. Creates directories as
// necessary. Attempts to chmod to the src mode. Returns an error if the file
// is src file is irregular, i.e. link, pipe, or device.
func copyFile(si os.FileInfo, src, dst string) (err error) {
	switch {
	case si.Mode().IsDir():
		return os.MkdirAll(dst, si.Mode())
	case si.Mode().IsRegular():
		closeErr := func(f *os.File) {
			// Properly return a close error
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
		sf, err := os.Open(src)
		if err != nil {
			return err
		}
		defer closeErr(sf)
		df, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer closeErr(df)
		// Copy contents
		if _, err = io.Copy(df, sf); err != nil {
			return err
		} else if err = df.Sync(); err != nil {
			return err
		} else {
			return df.Chmod(si.Mode())
		}
	default:
		return ErrIrregularFile
	}
}
