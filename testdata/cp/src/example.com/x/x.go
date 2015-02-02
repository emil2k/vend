// Package is used as the current working directory package when testing copying
// of a dependency.
package x

import (
	_ "other.com/y" // this path should be modified
)
