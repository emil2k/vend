package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"os"

	"github.com/emil2k/y/lib/astutil"
)

// update goes through the package in the srcDir and updates import path as
// specified by the rw map, from key to value.
// Returns an error if unable to parse the package or if writing to a file.
func update(srcDir string, rw map[string]string) error {
	fs := token.NewFileSet()
	mode := parser.AllErrors | parser.ParseComments
	pkgs, err := parser.ParseDir(fs, srcDir, nil, mode)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			if err := rwFile(fs, f, rw); err != nil {
				return err
			}
		}
	}
	return nil
}

// rwFile rewrites the import paths inside a file based on the rw map from keys
// to values.
// Returns an error if writing or closing the file fails.
func rwFile(fs *token.FileSet, f *ast.File, rw map[string]string) error {
	for op, np := range rw {
		if err := rwImport(fs, f, op, np); err != nil {
			return err
		}
	}
	return nil
}

// rwImport rewrites the import path from the old path, op, to the new path, np,
// inside a file and then writes the changes to it.
// If the old path is not used in a file nothing and an nil error is returned.
// Returns an error if writing or closing the file fails.
func rwImport(fs *token.FileSet, f *ast.File, op, np string) (err error) {
	if rw := astutil.RewriteImport(fs, f, op, np); !rw {
		return nil
	}
	// Open up the file and write the changes to it.
	if tf := fs.File(f.Pos()); tf != nil {
		var wf *os.File
		wf, err = os.OpenFile(tf.Name(), os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		// Properly catch a close error
		defer func() {
			if cerr := wf.Close(); err == nil {
				err = cerr
			}
		}()
		if err = printer.Fprint(wf, fs, f); err != nil {
			return err
		}
		// Output
		if opt.verbose {
			printBold(fmt.Sprintf("%s => %s", op, np))
			fmt.Println(tf.Name())
		}
	}
	return nil
}
