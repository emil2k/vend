# Y
[![Travis
branch](https://img.shields.io/travis/emil2k/y.svg?style=flat)](https://travis-ci.org/emil2k/y)
[![Coverage
Status](https://img.shields.io/coveralls/emil2k/y.svg?style=flat)](https://coveralls.io/r/emil2k/y)

**WARNING: This is a work in progress, if you want to help jump in.**

A Swiss Army knife for vending your own Go packages.

## Installation

```
go install github.com/emil2k/y -o vend
```

## Usage

### `vend list`

Lists all the dependencies of the package specified by the `[path]`, if ommitted
defaults to the current working directory. The `[path]` can be specified
relative to the current working directory or as an import path resolved through
the `GOPATH`.

```
  vend list [arguments] [path]

  -c=true: output child packages, stationed inside subdirectories
  -q=false: outputs only import paths
  -s=true: output standard library packages
  -t=true: include test files
  -v=false: outputs import details
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

### `vend init [directory]`

Recursively goes through the package in the current working directory and its
dependencies and copies all external packages into the specified `[directory]`,
while updating all the import paths and [qualified
identifiers](https://golang.org/ref/spec#Qualified_identifiers). Qualified
identifiers aren't modified if the package name is defined during import. The
packages are copied into a subdirectory specified by the basename of the
package.

For example, `github.com/lib/pq` would be copied to `[directory]/pq`. If there
are any conflicts, such as two `pq` packages, the packages will be copied to
subdirectories containing their full import paths, i.e.
`[directory]/github.com/lib/pq`, and a warning will be printed. The user can
then use `vend mv` to move the packages into unique subdirectories.

Example :

```
vend init lib
```

### `vend mv [from] [to]`

Moves the package in the `[from]` path or directory to the `[to]` directory,
updating all the necessary import paths and [qualified
identifiers](https://golang.org/ref/spec#Qualified_identifiers) for the package
in the current working directory.

Example :

```
vend mv lib/pq lib/postgresql
```

### `vend cp [from] [to]`

Copies the package in the `[from]` import path or directory to the `[to]`
directory, updating all the necessary import paths and [qualified
identifiers](https://golang.org/ref/spec#Qualified_identifiers) for the package
in the current working directory.

Example :

```
vend cp image/png lib/mypng
```

### `vend each [command]`

Changes to the directory of each dependency, outside of the standard library,
for the package in the current working directory and runs the `[command]`.

Example :

```
vend each go test -v .
```
