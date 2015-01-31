package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"

	"os"

	"github.com/emil2k/vend/lib/astutil"
)

// update subcommand updates the import paths in the `src` directory that
// contain the `from` import or an import located in its subdirectory to the
// equivalent import in the `to` path.
// Recurses into subdirectories to update import paths based on the `recurse`
// parameter.
func update(ctx *build.Context, src, from, to string, recurse bool) error {
	process := func(srcPkg *build.Package, _ error) error {
		// Get a list of all imports for the package in the src
		// directory, to determine which child package also need to be
		// updated.
		srcImp, srcDir := srcPkg.ImportPath, srcPkg.Dir
		if len(srcImp) == 0 || len(srcDir) == 0 {
			return fmt.Errorf("no import path or directory for src package")
		}
		// Compile map of import paths to change.
		rw := make(map[string]string)
		// Get the children of the src package their import paths will
		// be updated as well.
		for _, a := range getImports(srcPkg, true) {
			switch {
			case from == a:
				rw[from] = to
			case isChildPackage(from, a):
				cp, err := changePathParent(from, to, a)
				if err != nil {
					return err
				}
				rw[a] = cp
			}
		}
		if len(rw) > 0 {
			return rwDir(srcDir, rw)
		}
		return nil
	}
	if recurse {
		// Recurse into subdirectory packages.
		if err := recursePackages(ctx, src, process); err != nil {
			return err
		}
	} else if err := process(getPackage(ctx, src, src)); err != nil {
		return err
	}
	return nil
}

// rwDir goes through the package in the srcDir and updates import path as
// specified by the rw map, from key to value.
// Returns an error if unable to parse the package or if writing to a file.
func rwDir(srcDir string, rw map[string]string) error {
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

// printerConfig configures the AST pretty printing, it should use space for
// alignment and tabs to indent to mirror gofmt.
var printerConfig = &printer.Config{
	Mode:     printer.UseSpaces | printer.TabIndent,
	Tabwidth: 8,
	Indent:   0,
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
		// Print properly
		if err = printerConfig.Fprint(wf, fs, f); err != nil {
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
