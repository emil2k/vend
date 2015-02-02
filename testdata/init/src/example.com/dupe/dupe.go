// Package dupe is used for testing the init command for the duplicate package
// name error.
package dupe

import (
	a1 "other.com/y/a1"
	a2 "other.com/y/a2"
)

func main() {
	a1.ANop()
	a2.ANop()
}
