package main

// mainUsage describes usage of the overall tool.
const mainUsage string = `
A Swiss Army knife for vending your own Go packages.

  vend [subcommand] [arguments]

Valid subcommands :

  vend list
  vend info
  vend init
  vend cp

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

// cpUsage describes usage of the cp subcommand.
const cpUsage string = `
Copies the package in the [from] import path or directory to the [to]
directory, updating the necessary import paths for the package in the current
working directory.

  vend cp [from] [to]
`
