package main

// mainUsage describes usage of the overall tool.
const mainUsage string = `
A Swiss Army knife for vending your own Go packages.

  vend [subcommand] [arguments]

Valid subcommands :

  vend init
  vend cp
  vend mv
  vend path
  vend list
  vend info

For help with subcommands run :

  vend [subcommand] -h
`

// listUsage describes usage of the list subcommand.
const listUsage string = `
Lists all the dependencies of the package specified by the [path], if ommitted
defaults to the current working directory. The [path] can be specified
relative to the current working directory or as an import path resolved through
the GOPATH.

  vend list [arguments] [path]
`

// infoUsage describes usage of the info subcommand.
const infoUsage string = `
Print out information about the package specified by the [path], if ommitted
defaults to the current working directory. The [path] can be specified relative
to the current working directory or as an import path resolved through the
GOPATH.

  vend info [arguments] [path]
`

// initUsage describes usage of the init subcommand.
const initUsage string = `
For the package in the current working directory copies all external packages
into the specified [directory], while updating all the import paths. The
specified [directory] is created if necessary. External packages are packages
not located in the standard library, a parent directory, or a subdirectory.

The packages are copied into a subdirectory specified by the package name. If
multiple dependencies have the same package name the command will fail and
provide all the duplicates, the user should use the vend cp command to place
those packages in unique directories before running vend init again to process
the other packages.

  vend init [directory]
`

// cpUsage describes usage of the cp subcommand.
const cpUsage string = `
Copies the package in the [from] import path or directory to the [to]
directory, updating the necessary import paths for the package in the current
working directory.

  vend cp [from] [to]
`

// mvUsage describes usage of the mv subcommand.
const mvUsage string = `
Moves the package in the [from] path or directory to the [to] directory,
updating the necessary import paths for the package in the current working
directory. The mv subcommand cannot be used with standard packages, use
cp instead.

  vend mv [from] [to]
`

// pathUsage describes usage of the path subcommand.
const pathUsage string = `
Updates all the usages of the import path [from] to the import path [to] for
the package in the current working directory. When updating it includes import
paths located in subdirectories of the [from] import path, updating them to
their corresponding location in the [to] import path.

  vend path [from] [to]
`
