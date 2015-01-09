package main

import (
	"fmt"

	"github.com/mgutz/ansi"
)

// errColorCode is placed to output an error in color to the terminal.
var errColorCode = ansi.ColorCode("red+b")

// boldColorCode is placed to output bold text to the terminal.
var boldColorCode = ansi.ColorCode("cyan+b")

// endColorCode is placed to signfy an end to a colored output to the terminal.
var endColorCode = ansi.ColorCode("reset")

// printErr prints out an error in color to the terminal.
func printErr(err string) {
	fmt.Println(errColorCode + err + endColorCode)
}

// printBold prints out text in bold to the terminal.
func printBold(b string) {
	fmt.Println(boldColorCode + b + endColorCode)
}

// hasString checks if the slice has the particular string.
func hasString(hay []string, needle string) bool {
	for _, v := range hay {
		if v == needle {
			return true
		}
	}
	return false
}
