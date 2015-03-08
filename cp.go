package main

import (
	"errors"
	"fmt"
	"go/build"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
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
// Includes hidden files (staring with a dot) when copying files based on the
// `hidden` parameter.
func cp(ctx *build.Context, cwd, src, dst string, recurse, hidden bool) (err error) {
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
	if err = copyDir(src, dst, hidden); err != nil {
		return err
	}
	// Strip the canonical import path from files.
	if err = stripCanonicalImportPathDir(dst); err != nil {
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
	if err := path(ctx, dst, srcImp, dstImp, true); err != nil {
		return err
	}
	// Update the import paths, if the recurse flag is set recurse through
	// the subdirectories and update import paths.
	return path(ctx, cwd, srcImp, dstImp, recurse)
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
// Skips hidden files base on the `hidden` parameter.
func copyDir(src, dst string, hidden bool) error {
	// First compile a list of copies to execute then execute, otherwise
	// infinite copy situations could arise when copying a parent directory
	// into a child directory.
	cjs := make([]copyFileJob, 0)
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Determine whether copying a hidden file and whether to skip
		// it or not.
		if base := filepath.Base(path); base != "." && base != ".." &&
			!hidden && strings.HasPrefix(base, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
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

// stripCanonicalImportPathDir strips the canonical import path from all files
// in the directory.
func stripCanonicalImportPathDir(dir string) error {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() ||
			!strings.HasSuffix(filepath.Base(path), ".go") {
			return nil
		}
		return stripCanonicalImportPathFile(path)
	}
	return filepath.Walk(dir, walk)
}

// stripCanonicalImportPathFile strips the canonical import path from the file
// at the path.
func stripCanonicalImportPathFile(path string) error {
	src, err := getFileContents(path)
	if err != nil {
		return err
	}
	contains, start, end := containsCanonicalImportPath(src)
	if !contains {
		return nil // nothing to do
	}
	// Strip the path and write to the file.
	strip := append(src[:start], src[end:]...)
	return ioutil.WriteFile(path, strip, 0)
}

// containsCanonicalImportPath check whether the src contains a canonical
// import path, and if so returns the file offsets for where the package
// declaration ends to where the comment ends.
// Offset start at 0.
func containsCanonicalImportPath(src []byte) (contains bool, start, end int) {
	fs := token.NewFileSet()
	tf := fs.AddFile("", fs.Base(), len(src))
	var s scanner.Scanner
	s.Init(tf, src, nil, scanner.ScanComments)
	currentLine := 0
	type scanned struct {
		pos token.Position
		tok token.Token
		lit string
	}
	var line []scanned // current line buffer
	for {
		pos, tok, lit := s.Scan()
		posd := fs.Position(pos)
		if tok == token.EOF || posd.Line != currentLine {
			// Does it match the signature of the canonical import
			// path comment, which will be preceded by a PACKAGE,
			// IDENT, and SEMICOLON.
			if len(line) == 4 &&
				line[0].tok == token.PACKAGE &&
				line[1].tok == token.IDENT &&
				line[2].tok == token.SEMICOLON &&
				line[3].tok == token.COMMENT {
				return true,
					line[1].pos.Offset + len(line[1].lit),
					line[3].pos.Offset + len(line[3].lit)
			}
			// Reset the current line.
			line = make([]scanned, 0, 0)
			currentLine = posd.Line
		}
		line = append(line, scanned{posd, tok, lit})
		if tok == token.EOF {
			break
		}
	}
	return false, 0, 0
}

// getFileContens opens the file at the provided path, reads all the content,
// and returns it.
func getFileContents(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return content, nil
}
