# Vend
[![Travis
branch](https://img.shields.io/travis/emil2k/vend.svg?style=flat)](https://travis-ci.org/emil2k/vend)
[![Coverage
Status](https://img.shields.io/coveralls/emil2k/vend.svg?style=flat)](https://coveralls.io/r/emil2k/vend)

**WARNING: This is a work in progress, if you want to help jump in.**

A Swiss Army knife for vending your own Go packages.

## Installation

```
go get github.com/emil2k/vend
```

## Compatibility

- Go 1.2+
- Should work on OSX and Linux, someone should test it on Windows.

## Usage

This tool makes a couple of assumptions about a given package :

- The dependencies are present, run `go get -u`.
- Located in the proper location in the `GOPATH`.

  If you are working on a fork, make sure the package is located in the
directory in your `GOPATH` allocated for the original package. You can add your
fork as a remote.

### `vend init`

For the package in the current working directory copies all external packages
into the specified `[directory]`, while updating all the import paths. The
specified `[directory]` is created if necessary. External packages are packages
not located in the standard library, a parent directory, or a subdirectory.

The packages are copied into a subdirectory specified by the package name. If
multiple dependencies have the same package name the command will fail and
provide all the duplicates, the user should use the `vend cp` command to place
those packages in unique directories before running `vend init` again to process
the other packages.

```
vend init [directory]

-f=false: forces copy, replaces destination folder
-i=false: include hidden files, files starting with a dot
-r=false: recurse into subdirectories to vend their imports as well
-v=false: detailed output
```

Example :

```
vend init ./lib
```

### `vend cp`

Copies the package in the `[from]` import path or directory to the `[to]`
directory, updating the necessary import paths for the package in the current
working directory.

```
vend cp [from] [to]

-f=false: forces copy, replaces destination folder
-i=false: include hidden files, files starting with a dot
-v=false: detailed output
```

Example :

```
vend cp image/png ./lib/mypng
```

### `vend mv`

Moves the package in the `[from]` path or directory to the `[to]` directory,
updating the necessary import paths for the package in the current working
directory. The `mv` subcommand cannot be used with standard packages, use
`cp` instead.

```
vend mv [from] [to]

-f=false: forces move, replaces destination folder
-i=false: include hidden files, files starting with a dot
-v=false: detailed output
```

Example :

```
vend mv ./lib/pq ./lib/postgresql
```

### `vend list`

Lists all the dependencies of the package specified by the `[path]`, if ommitted
defaults to the current working directory. The `[path]` can be specified
relative to the current working directory or as an import path resolved through
the `GOPATH`.

```
vend list [arguments] [path]

-c=false: omit child packages, located in subdirectories
-q=false: outputs only import paths
-r=false: include imports from packages located in subdirectories
-s=false: omit standard packages
-t=false: omit test files when compiling imports
-v=false: outputs details for each import
```

### `vend info`

Print out information regarding the package specified by the `[path]`, if
ommitted defaults to the current working directory. The `[path]` can be
specified relative to the current working directory or as an import path
resolved through the `GOPATH`.

```
vend info [arguments] [path]

-v=false: detailed output
```

### TODO: `vend name`

Changes the package name of the package specified by the `[path]` import path or
directory to the `[name]`, updating all the [qualified
identifiers](https://golang.org/ref/spec#Qualified_identifiers) for the package
in the current working directory. Qualified identifiers aren't modified if the
package name is defined during import. The `name` subcommand cannot be used with
standard packages, you must first `cp` the package out of the `GOROOT`.

Example :

```
vend name ./lib/mypq mypq
```

### TODO: `vend each`

Changes to the directory of each dependency, outside of the standard library,
for the package in the current working directory and runs the `[command]`.

Example :

```
vend each go test -v .
```
