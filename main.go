package main

import "os"

// main parses arguments and flags and passes the arguments to the correct
// handler function.
func main() {
	var err error
	if len(os.Args) < 2 {
		printErr("Subcommand not specified")
		flagMap["main"].Usage()
		os.Exit(1)
	} else {
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
			err = list(path)
		case "info":
			f := flagMap["info"]
			f.Parse(os.Args[2:])
			var path string
			if len(f.Args()) > 0 {
				path = f.Arg(0)
			} else {
				path = "."
			}
			err = info(path)
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
