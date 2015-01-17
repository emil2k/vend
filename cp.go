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
func cp(ctx *build.Context, cwd, src, dst string) (err error) {
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
	var srcPkg, dstPkg, cwdPkg *build.Package
	// May fail because there is multiple packages in the folder but all
	// that is necessary here is the directory and the import path.
	// Can't use the build.MultiplePackageError, to detect the error because
	// it was only added in 1.4, and we want 1.2+.
	if srcPkg, err = getPackage(ctx, src); len(srcPkg.Dir) == 0 {
		if err == nil {
			return fmt.Errorf("package has no directory")
		}
		return
	} else if srcImp = srcPkg.ImportPath; len(srcImp) == 0 {
		if err == nil {
			return fmt.Errorf("package has no import path")
		}
		return
	} else if src, err = cwdAbs(cwd, srcPkg.Dir); err != nil {
		return
	} else if dst, err = cwdAbs(cwd, dst); err != nil {
		return
	} else if err = copyDir(src, dst); err != nil {
		return
	}
	// Determine import path of the new package, and update import paths in
	// the current working directory.
	// Update the import paths of the new package and its children.
	if dstPkg, err = getPackage(ctx, dst); len(dstPkg.ImportPath) == 0 {
		return
	} else {
		dstImp = dstPkg.ImportPath
	}
	// Get a list of all imports for the package in the current working
	// directory, to determine which child package also need to be updated.
	if cwdPkg, err = getPackage(ctx, cwd); err != nil {
		return
	}
	// Compile map of import paths to change.
	rw := map[string]string{srcImp: dstImp}
	// Get the children of the src package their import paths will be
	// updated as well.
	for _, a := range getImports(cwdPkg, true) {
		if isChildPackage(srcImp, a) {
			rw[a], err = changePathParent(srcImp, dstImp, a)
			if err != nil {
				return err
			}
		}
	}
	// Update import paths in the copied package itself, as it may contain
	// an external _test package that imports itself.
	if err := update(dst, rw); err != nil {
		return err
	}
	// Update the import paths.
	return update(cwd, rw)
}

// changePathParent allows changing of a child import path to a new directory
// by specifiying a their parent packages import path before `a` and after `b`.
func changePathParent(a, b, child string) (string, error) {
	a = filepath.FromSlash(a)
	b = filepath.FromSlash(b)
	child = filepath.FromSlash(child)
	rel, err := filepath.Rel(a, child)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(b, rel)), nil
}

// cwdAbs returns the path as absolute relative to the base directory if it is
// not absolute.
func cwdAbs(base, path string) (string, error) {
	path = filepath.Clean(path)
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Join(base, path), nil
}

// copyDir recursively copies the src directory to the desination directory.
// Creates directories as necessary. Attempts to chmod everything to the src
// mode.
// With the opt.verbose option set outputs the src and destination of each
// copied file.
// Skips hidden files base on the opt.hidden option.
func copyDir(src, dst string) error {
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
		return copyFile(info, path, fileDst)
	}
	return filepath.Walk(src, walk)
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
