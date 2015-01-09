package main

import (
	"flag"
	"fmt"
)

// flagMap maps a subcommand to its configured FlagSet.
var flagMap map[string]*flag.FlagSet = make(map[string]*flag.FlagSet)

// mainUsage describes usage of the overall tool.
const mainUsage string = `
A Swiss Army knife for vending your own Go packages.

  vend [subcommand] [arguments]

Valid subcommands :

  vend list
  vend info
  vend init
  vend mv

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

// usage returns a Usage function that simply prints the passed string, and the
// default usage.
func usage(fs *flag.FlagSet, use string) func() {
	return func() {
		fmt.Println(use)
		fs.PrintDefaults()
	}
}

// optHolder represents argument passed into the command.
type optHolder struct {
	// quite flag to reduce output.
	quite bool
	// verbose flag to increase output.
	verbose bool
	// tests flag includes the test packages.
	tests bool
	// standard flag includes standard library packages.
	standard bool
	// child flag includes child packages.
	child bool
}

// opt argumes passed into the command.
var opt optHolder = optHolder{}

// init initialiazes flags for each subcommand.
func init() {
	// Main flagset
	main := flag.NewFlagSet("main", flag.ExitOnError)
	main.Usage = usage(main, mainUsage)
	flagMap["main"] = main
	// List flagset
	list := flag.NewFlagSet("list", flag.ExitOnError)
	list.Usage = usage(list, listUsage)
	list.BoolVar(&opt.quite, "q", false, "outputs only import paths")
	list.BoolVar(&opt.verbose, "v", false, "outputs details for each import")
	list.BoolVar(&opt.tests, "t", true,
		"include test files when compiling imports")
	list.BoolVar(&opt.standard, "s", true,
		"output standard library packages")
	list.BoolVar(&opt.child, "c", true,
		"output child packages, stationed inside subdirectories")
	flagMap["list"] = list
	// Info flagset
	info := flag.NewFlagSet("info", flag.ExitOnError)
	info.Usage = usage(info, infoUsage)
	info.BoolVar(&opt.verbose, "v", false, "detailed output")
	flagMap["info"] = info
}
