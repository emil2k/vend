package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/emil2k/vend/lib/ansi"
)

// errColorCode is placed to output an error in color to the terminal.
var errColorCode = ansi.ColorCode("red+b")

// boldColorCode is placed to output bold text to the terminal.
var boldColorCode = ansi.ColorCode("cyan+b")

// endColorCode is placed to signfy an end to a colored output to the terminal.
var endColorCode = ansi.ColorCode("reset")

// printErr prints out an error in color to the terminal.
func printErr(err ...string) {
	fmt.Println(errColorCode + strings.Join(err, " ") + endColorCode)
}

// printBold prints out text in bold to the terminal.
func printBold(b ...string) {
	fmt.Println(boldColorCode + strings.Join(b, " ") + endColorCode)
}

// printWrap prints out the text with a newline each n characters, does not
// split up words.
func printWrap(n int, str ...string) {
	var out int
	for _, v := range str {
		for _, w := range strings.Split(v, " ") {
			if out+len(w) > n {
				out = 0
				fmt.Println()
			}
			fmt.Printf("%s ", w)
			out += len(w)
		}
	}
	fmt.Println()
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

// appendUnique appends the provided strings to the list if they are not already
// present inside.
func appendUnique(list []string, add ...string) []string {
	for _, v := range add {
		if !hasString(list, v) {
			list = append(list, v)
		}
	}
	return list
}

// cwdAbs returns the path as absolute relative to the base directory if it is
// not absolute.
func cwdAbs(base, path string) (string, error) {
	path = filepath.Clean(path)
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Join(base, path), nil
}
