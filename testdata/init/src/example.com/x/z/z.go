// Package z is used for testing the recursive init command, it is a child
// package of x.
package z

import (
	"other.com/y/a1"
	"other.com/y/c"
)

func main() {
	a.ANop()
	c.CNop()
}
