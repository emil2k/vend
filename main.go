package main

import (
	"go/build"
	"os"
)

// main parses arguments and flags and passes the arguments to the correct
// handler function.
func main() {
	var err error
	if len(os.Args) < 2 {
		printErr("Subcommand not specified")
		flagMap["main"].Usage()
		os.Exit(1)
	} else {
		var cwd string
		if cwd, err = os.Getwd(); err != nil {
			printErr("Error : " + err.Error())
			os.Exit(1)
		}
		ctx := &build.Default
		ctx.UseAllFiles = true
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
			err = list(ctx, cwd, path)
		case "info":
			f := flagMap["info"]
			f.Parse(os.Args[2:])
			var path string
			if len(f.Args()) > 0 {
				path = f.Arg(0)
			} else {
				path = "."
			}
			err = info(ctx, cwd, path)
		case "cp":
			f := flagMap["cp"]
			f.Parse(os.Args[2:])
			if len(f.Args()) > 1 {
				err = cp(ctx, cwd, f.Arg(0), f.Arg(1))
			} else {
				printErr("Missing arguments")
				f.Usage()
				os.Exit(1)
			}
		case "-h":
			flagMap["main"].Usage()
		default:
			printErr("Invalid subcommand : " + os.Args[1])
			flagMap["main"].Usage()
			os.Exit(1)
		}
	}
	// Output errors and exit
	if err != nil {
		printErr("Error : " + err.Error())
		os.Exit(1)
	}
}
