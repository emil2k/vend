package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// cp copies the package at the specified path to the specified destination
// directory.
// TODO update imports and update qualified identifiers.
// Returns errors if the source package cannot be found or built, or if an
// absolute path cannot be determined either for the source or destination.
func cp(path, dst string) error {
	pkg, err := getPackage(path)
	if err != nil {
		return err
	}
	src, err := filepath.Abs(pkg.Dir)
	if err != nil {
		return err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return err
	}
	return copyDir(src, dst)
}

// copyDir recursively copies the src directory to the desination directory.
// Creates directories as necessary. Attempts to chmod everything to the src
// mode.
// With the opt.verbose option set outputs the src and destination of each
// copied file.
func copyDir(src, dst string) error {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
