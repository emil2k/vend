package main

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
)

// main parses arguments and flags and passes the arguments to the correct
// handler function.
func main() {
	var err error
	if len(os.Args) < 2 {
		err = errors.New("subcommand not specified")
		flagMap["main"].Usage()
	} else {
		switch os.Args[1] {
		case "list":
			f := flagMap["list"]
			f.Parse(os.Args[2:])
			var path string
			if len(f.Args()) > 0 {
				path = f.Arg(0)
			} else {
				path = "."
			}
			err = list(path)
		case "-h":
			flagMap["main"].Usage()
		default:
			err = errors.New("invalid subcommand")
			flagMap["main"].Usage()
		}
	}
	// Output any errors
	if err != nil {
		printErr("Error : " + err.Error())
		os.Exit(1)
	}
}

// list runs the list subcommand, listing all the dependencies of the package in
// at the specified path, relative paths are resolved from the current working
// directory.
func list(path string) error {
	var pkg *build.Package
	var err error
	var cwd string
	var pi os.FileInfo
	// Check if the path given is a directory in the current working
	// directory.
	// Otherwise attempt to import the path, relative path with respect to
	// current working directory.
	if cwd, err = os.Getwd(); err != nil {
		return err
	} else if pi, err = os.Stat(path); err == nil && pi.IsDir() {
		pkg, err = build.ImportDir(filepath.Join(cwd, path), 0)
		if err != nil {
			return err
		}
	} else if pkg, err = build.Import(path, cwd, 0); err != nil {
		return err
	}
	printBold("GOROOT : " + build.Default.GOROOT)
	printBold("GOPATH : " + build.Default.GOPATH)
	printBold("IMPORT PATH : " + pkg.ImportPath)
	printBold("DIRECTORY : " + pkg.Dir)
	printBold("OBJECT : " + pkg.PkgObj)
	// List imports
	for _, v := range pkg.Imports {
		fmt.Println(v)
	}
	// List import line positions
	for k, ps := range pkg.ImportPos {
		printBold(k)
		for _, p := range ps {
			fmt.Println(p)
		}
	}
	return nil
}
