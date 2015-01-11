package main

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

// TestUpdate copies a standard package then update its import paths and checks
// to make sure that they were updated.
// Removes the copied package when done with it.
func TestUpdate(t *testing.T) {
	dst, err := ioutil.TempDir(os.TempDir(), "testupdate")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dst)
	if err := cp("encoding/json", dst); err != nil {
		t.Errorf("error while copying standard package : %s\n",
			err.Error())
	} else if pkg, err := getPackage(dst); err != nil {
		t.Errorf("error while importing copied package : %s\n",
			err.Error())
	} else if err := update(pkg.Dir, map[string]string{
		"unicode": "myuni",
	}); err != nil {
		t.Errorf("error during update : %s\n", err.Error())
	} else {
		// Parse the copied package and check that the import path was
		// updated.
		fs := token.NewFileSet()
		mode := parser.AllErrors | parser.ParseComments
		pkgs, err := parser.ParseDir(fs, pkg.Dir, nil, mode)
		if err != nil {
			t.Errorf("error while parsing copied package : %s\n",
				err.Error())
		}
		newFound := false
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, i := range file.Imports {
					// Import path values are quoted
					if i.Path.Value == strconv.Quote("unicode") {
						t.Errorf("old import found : %s", i)
					} else if i.Path.Value == strconv.Quote("myuni") {
						newFound = true
					}
				}
			}
		}
		if !newFound {
			t.Errorf("new import not found")
		}
	}
}
