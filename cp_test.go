package main

import (
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"code.google.com/p/go-uuid/uuid"
)

// TestCp copies a standard package to a temporary folder, folder is cleaned up
// afterward. Tests copying to a desitination folder where the parent folder
// doesn't exist, and must be created.
// Test makes sure imports and files match in the src and dst packages, and that
// the packages can be built.
func TestCp(t *testing.T) {
	tmp, err := ioutil.TempDir(os.TempDir(), "testcp")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(tmp)
	parent := filepath.Join(tmp, uuid.New())
	dst := filepath.Join(parent, uuid.New())
	if err := cp("image", dst); err != nil {
		t.Errorf("error : %s\n", err.Error())
	} else if srcPkg, err := getPackage("image"); err != nil {
		t.Errorf("error during src import : %s\n", err.Error())
	} else if dstPkg, err := build.ImportDir(dst, 0); err != nil {
		t.Errorf("error during dst import : %s\n", err.Error())
	} else if !reflect.DeepEqual(srcPkg.GoFiles, dstPkg.GoFiles) {
		t.Errorf("package files don't match, %s : %s\n",
			srcPkg.GoFiles, dstPkg.GoFiles)
	} else if !reflect.DeepEqual(srcPkg.Imports, dstPkg.Imports) {
		t.Errorf("package imports don't match, %s : %s\n",
			srcPkg.Imports, dstPkg.Imports)
	} else if !reflect.DeepEqual(srcPkg.TestGoFiles, dstPkg.TestGoFiles) {
		t.Errorf("package test files don't match, %s : %s\n",
			srcPkg.TestGoFiles, dstPkg.TestGoFiles)
	} else if !reflect.DeepEqual(srcPkg.TestImports, dstPkg.TestImports) {
		t.Errorf("package test imports don't match, %s : %s\n",
			srcPkg.TestImports, dstPkg.TestImports)
	}
}
