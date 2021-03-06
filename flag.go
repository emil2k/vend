package main

import (
	"flag"
	"fmt"
)

// flagMap maps a subcommand to its configured FlagSet.
var flagMap map[string]*flag.FlagSet = make(map[string]*flag.FlagSet)

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
	// recurse flag whether to recurse the command over all packages found
	// in subdirectories.
	recurse bool
	// tests flag omits test files.
	tests bool
	// standard flag omits standard packages.
	standard bool
	// child flag omits packages located in subdirectories.
	child bool
	// force flag forces the command to execute after a warning.
	force bool
	// hidden flag includes hidden files, starting with a dot, when copying
	// or moving files.
	hidden bool
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
	list.BoolVar(&opt.recurse, "r", false,
		"include imports from packages located in subdirectories")
	list.BoolVar(&opt.verbose, "v", false, "outputs details for each import")
	list.BoolVar(&opt.tests, "t", false,
		"omit test files when compiling imports")
	list.BoolVar(&opt.standard, "s", false,
		"omit standard packages")
	list.BoolVar(&opt.child, "c", false,
		"omit child packages, located in subdirectories")
	flagMap["list"] = list
	// Info flagset
	info := flag.NewFlagSet("info", flag.ExitOnError)
	info.Usage = usage(info, infoUsage)
	info.BoolVar(&opt.verbose, "v", false, "detailed output")
	flagMap["info"] = info
	// Init flagset
	init := flag.NewFlagSet("init", flag.ExitOnError)
	init.Usage = usage(init, initUsage)
	init.BoolVar(&opt.verbose, "v", false, "detailed output")
	init.BoolVar(&opt.recurse, "r", false,
		"recurse into subdirectories to include their dependencies")
	init.BoolVar(&opt.force, "f", false,
		"forces copy, replaces destination folder")
	init.BoolVar(&opt.hidden, "i", false,
		"include hidden files, files starting with a dot")
	flagMap["init"] = init
	// Cp flagset
	cp := flag.NewFlagSet("cp", flag.ExitOnError)
	cp.Usage = usage(cp, cpUsage)
	cp.BoolVar(&opt.verbose, "v", false, "detailed output")
	cp.BoolVar(&opt.recurse, "r", false,
		"recurse into subdirectories to update their import paths of the copied packages")
	cp.BoolVar(&opt.force, "f", false,
		"forces copy, replaces destination folder")
	cp.BoolVar(&opt.hidden, "i", false,
		"include hidden files, files starting with a dot")
	flagMap["cp"] = cp
	// Mv flagset
	mv := flag.NewFlagSet("mv", flag.ExitOnError)
	mv.Usage = usage(mv, mvUsage)
	mv.BoolVar(&opt.verbose, "v", false, "detailed output")
	mv.BoolVar(&opt.recurse, "r", false,
		"recurse into subdirectories to update their import paths of the moved packages")
	mv.BoolVar(&opt.force, "f", false,
		"forces move, replaces destination folder")
	mv.BoolVar(&opt.hidden, "i", false,
		"include hidden files, files starting with a dot")
	flagMap["mv"] = mv
	// Path flagset
	path := flag.NewFlagSet("path", flag.ExitOnError)
	path.Usage = usage(path, pathUsage)
	path.BoolVar(&opt.verbose, "v", false, "detailed output")
	path.BoolVar(&opt.recurse, "r", false,
		"recurse into subdirectories to update their import paths")
	flagMap["path"] = path
}
