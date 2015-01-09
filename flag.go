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

	vend init
	vend list
	vend mv

For help with subcommands run :

	vend [subcommand] -h
`

// listUsage describes usage of the list subcommand.
const listUsage string = `
Lists all the dependencies of the package in the [directory] if ommitted
defaults to the current working directory.

	vend list [directory]
`

// usage returns a Usage functions that simply print the passed string.
func usage(use string) func() {
	return func() {
		fmt.Println(use)
	}
}

// init initialiazes flags for each subcommand.
func init() {
	flagMap["main"] = flag.NewFlagSet("main", flag.ExitOnError)
	flagMap["main"].Usage = usage(mainUsage)
	flagMap["list"] = flag.NewFlagSet("list", flag.ExitOnError)
	flagMap["list"].Usage = usage(listUsage)
}
